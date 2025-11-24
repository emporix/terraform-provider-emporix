package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &EmporixProvider{}

type EmporixProvider struct {
	version string
}

type EmporixProviderModel struct {
	Tenant      types.String `tfsdk:"tenant"`
	AccessToken types.String `tfsdk:"access_token"`
	ApiUrl      types.String `tfsdk:"api_url"`
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
				Description: "OAuth2 access token for Emporix API. Can be set via EMPORIX_ACCESS_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
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

	if config.ApiUrl.IsNull() {
		apiUrl := os.Getenv("EMPORIX_API_URL")
		if apiUrl == "" {
			apiUrl = "https://api.emporix.io"
		}
		config.ApiUrl = types.StringValue(apiUrl)
	}

	// Validate required fields
	if config.Tenant.IsNull() || config.Tenant.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("tenant"),
			"Missing Tenant Configuration",
			"The provider cannot create the Emporix API client as there is a missing or empty value for the Emporix tenant. "+
				"Set the tenant value in the configuration or use the EMPORIX_TENANT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if config.AccessToken.IsNull() || config.AccessToken.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("access_token"),
			"Missing Access Token Configuration",
			"The provider cannot create the Emporix API client as there is a missing or empty value for the access token. "+
				"Set the access_token value in the configuration or use the EMPORIX_ACCESS_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API client
	client := &EmporixClient{
		Tenant:      config.Tenant.ValueString(),
		AccessToken: config.AccessToken.ValueString(),
		ApiUrl:      config.ApiUrl.ValueString(),
	}
	client.SetContext(ctx)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *EmporixProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSiteSettingsResource,
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
