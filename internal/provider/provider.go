package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ provider.Provider = &EmporixProvider{}

type EmporixProvider struct {
	version string
}

type EmporixProviderModel struct {
	Tenant       types.String `tfsdk:"tenant"`
	AccessToken  types.String `tfsdk:"access_token"`
	ClientId     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	Scope        types.String `tfsdk:"scope"`
	ApiUrl       types.String `tfsdk:"api_url"`
}

func (p *EmporixProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "emporix"
	resp.Version = p.version
}

func (p *EmporixProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for Emporix API",
		Attributes: map[string]schema.Attribute{
			"tenant": schema.StringAttribute{
				Description: "Emporix tenant name (lowercase). Can be set via EMPORIX_TENANT environment variable.",
				Optional:    true,
			},
			"access_token": schema.StringAttribute{
				Description: "OAuth2 access token for Emporix API. Can be set via EMPORIX_ACCESS_TOKEN environment variable. If not provided, will be generated using client_id and client_secret.",
				Optional:    true,
				Sensitive:   true,
			},
			"client_id": schema.StringAttribute{
				Description: "OAuth2 client ID for generating access token. Can be set via EMPORIX_CLIENT_ID environment variable. Required if access_token is not provided.",
				Optional:    true,
				Sensitive:   true,
			},
			"client_secret": schema.StringAttribute{
				Description: "OAuth2 client secret for generating access token. Can be set via EMPORIX_CLIENT_SECRET environment variable. Required if access_token is not provided.",
				Optional:    true,
				Sensitive:   true,
			},
			"scope": schema.StringAttribute{
				Description: "OAuth2 scopes (space-separated). Optional. If not provided, no scope parameter is sent to the OAuth endpoint. Example: 'tenant=mytenant site.site_read site.site_manage'",
				Optional:    true,
			},
			"api_url": schema.StringAttribute{
				Description: "Emporix API base URL. Defaults to https://api.emporix.io. Can be set via EMPORIX_API_URL environment variable.",
				Optional:    true,
			},
		},
	}
}

func (p *EmporixProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config EmporixProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check environment variables if values not set in config
	if config.Tenant.IsNull() {
		config.Tenant = types.StringValue(os.Getenv("EMPORIX_TENANT"))
	}

	if config.AccessToken.IsNull() {
		config.AccessToken = types.StringValue(os.Getenv("EMPORIX_ACCESS_TOKEN"))
	}

	if config.ClientId.IsNull() {
		config.ClientId = types.StringValue(os.Getenv("EMPORIX_CLIENT_ID"))
	}

	if config.ClientSecret.IsNull() {
		config.ClientSecret = types.StringValue(os.Getenv("EMPORIX_CLIENT_SECRET"))
	}

	if config.Scope.IsNull() {
		config.Scope = types.StringValue(os.Getenv("EMPORIX_SCOPE"))
	}

	if config.ApiUrl.IsNull() {
		apiUrl := os.Getenv("EMPORIX_API_URL")
		if apiUrl == "" {
			apiUrl = "https://api.emporix.io"
		}
		config.ApiUrl = types.StringValue(apiUrl)
	}

	// Validate tenant (always required)
	if config.Tenant.IsNull() || config.Tenant.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("tenant"),
			"Missing Tenant Configuration",
			"The provider cannot create the Emporix API client as there is a missing or empty value for the Emporix tenant. "+
				"Set the tenant value in the configuration or use the EMPORIX_TENANT environment variable.",
		)
		return
	}

	// Check if we need to generate an access token
	needsTokenGeneration := config.AccessToken.IsNull() || config.AccessToken.ValueString() == ""

	if needsTokenGeneration {
		// Validate client credentials are provided
		if config.ClientId.IsNull() || config.ClientId.ValueString() == "" {
			resp.Diagnostics.AddError(
				"Missing Authentication Configuration",
				"Either access_token or client_id must be provided. "+
					"Set via provider configuration or environment variables (EMPORIX_ACCESS_TOKEN or EMPORIX_CLIENT_ID).",
			)
			return
		}

		if config.ClientSecret.IsNull() || config.ClientSecret.ValueString() == "" {
			resp.Diagnostics.AddError(
				"Missing Authentication Configuration",
				"Either access_token or client_secret must be provided. "+
					"Set via provider configuration or environment variables (EMPORIX_ACCESS_TOKEN or EMPORIX_CLIENT_SECRET).",
			)
			return
		}

		// Build scope - only if user provided it
		scope := ""
		if !config.Scope.IsNull() && config.Scope.ValueString() != "" {
			scope = config.Scope.ValueString()
		}

		// Generate access token
		token, err := generateAccessToken(
			ctx,
			config.ApiUrl.ValueString(),
			config.ClientId.ValueString(),
			config.ClientSecret.ValueString(),
			scope,
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Generate Access Token",
				fmt.Sprintf("Could not generate OAuth access token: %s", err.Error()),
			)
			return
		}

		config.AccessToken = types.StringValue(token)
		tflog.Debug(ctx, "Successfully generated OAuth access token")
	}

	// Create API client with timeout
	client := NewEmporixClient(
		config.Tenant.ValueString(),
		config.AccessToken.ValueString(),
		config.ApiUrl.ValueString(),
	)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *EmporixProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSiteSettingsResource,
		NewPaymentModeResource,
		NewCountryResource,
		NewCurrencyResource,
		NewTenantConfigurationResource,
		NewShippingZoneResource,
		NewSchemaResource,
		NewDeliveryTimeResource,
		NewShippingMethodResource,
	}
}

func (p *EmporixProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &EmporixProvider{
			version: version,
		}
	}
}
