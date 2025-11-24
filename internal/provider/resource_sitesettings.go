package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &SiteSettingsResource{}
var _ resource.ResourceWithImportState = &SiteSettingsResource{}

func NewSiteSettingsResource() resource.Resource {
	return &SiteSettingsResource{}
}

type SiteSettingsResource struct {
	client *EmporixClient
}

type SiteSettingsResourceModel struct {
	Code                      types.String `tfsdk:"code"`
	Name                      types.String `tfsdk:"name"`
	Active                    types.Bool   `tfsdk:"active"`
	Default                   types.Bool   `tfsdk:"default"`
	IncludesTax               types.Bool   `tfsdk:"includes_tax"`
	DefaultLanguage           types.String `tfsdk:"default_language"`
	Languages                 types.List   `tfsdk:"languages"`
	Currency                  types.String `tfsdk:"currency"`
	AvailableCurrencies       types.List   `tfsdk:"available_currencies"`
	ShipToCountries           types.List   `tfsdk:"ship_to_countries"`
	TaxCalculationAddressType types.String `tfsdk:"tax_calculation_address_type"`
	DecimalPoints             types.Int64  `tfsdk:"decimal_points"`
	HomeBase                  types.Object `tfsdk:"home_base"`
	AssistedBuying            types.Object `tfsdk:"assisted_buying"`
	Mixins                    types.Map    `tfsdk:"mixins"`
}

func (r *SiteSettingsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sitesettings"
}

func (r *SiteSettingsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Emporix site settings configuration.",
		Attributes: map[string]schema.Attribute{
			"code": schema.StringAttribute{
				Description: "Site unique identifier (code). Cannot be changed after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Site name.",
				Required:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Flag indicating whether the site is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"default": schema.BoolAttribute{
				Description: "Flag indicating whether the site is the tenant default site.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"includes_tax": schema.BoolAttribute{
				Description: "Indicates whether prices for the site should be returned in gross (true) or net (false).",
				Optional:    true,
			},
			"default_language": schema.StringAttribute{
				Description: "Site's default language, compliant with the ISO 639-1 standard (2-letter lowercase code).",
				Required:    true,
			},
			"languages": schema.ListAttribute{
				Description: "Languages supported by the site. Must be compliant with the ISO 639-1 standard.",
				ElementType: types.StringType,
				Required:    true,
			},
			"currency": schema.StringAttribute{
				Description: "Currency used by the site, compliant with the ISO 4217 standard (3-letter uppercase code).",
				Required:    true,
			},
			"available_currencies": schema.ListAttribute{
				Description: "List of the currencies supported by the site.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"ship_to_countries": schema.ListAttribute{
				Description: "Codes of countries to which the site ships products. Must be compliant with the ISO 3166-1 alpha-2 standard.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"tax_calculation_address_type": schema.StringAttribute{
				Description: "Specifies whether tax calculation is based on customer billing address or shipping address. Default value is BILLING_ADDRESS.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("BILLING_ADDRESS"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"decimal_points": schema.Int64Attribute{
				Description: "Number of decimal points used in the cart calculation. Must be zero or a positive value.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(2),
			},
			"home_base": schema.SingleNestedAttribute{
				Description: "Home base configuration for the site.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"address": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"street": schema.StringAttribute{
								Optional: true,
							},
							"street_number": schema.StringAttribute{
								Optional: true,
							},
							"zip_code": schema.StringAttribute{
								Optional: true,
							},
							"city": schema.StringAttribute{
								Optional: true,
							},
							"country": schema.StringAttribute{
								Required: true,
							},
							"state": schema.StringAttribute{
								Optional: true,
							},
						},
					},
					"location": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"latitude": schema.Float64Attribute{
								Optional: true,
							},
							"longitude": schema.Float64Attribute{
								Optional: true,
							},
						},
					},
				},
			},
			"assisted_buying": schema.SingleNestedAttribute{
				Description: "Assisted buying configuration for the site.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"storefront_url": schema.StringAttribute{
						Optional: true,
					},
				},
			},
			"mixins": schema.MapAttribute{
				Description: "Custom mixins for extending site configuration.",
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

func (r *SiteSettingsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*EmporixClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *EmporixClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *SiteSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SiteSettingsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Store the original ship_to_countries order from plan
	var originalShipToCountries []string
	if !plan.ShipToCountries.IsNull() {
		resp.Diagnostics.Append(plan.ShipToCountries.ElementsAs(ctx, &originalShipToCountries, false)...)
	}

	// Convert Terraform model to API model
	site, diags := r.terraformToApi(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the site
	err := r.client.CreateSite(site)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating site",
			fmt.Sprintf("Could not create site: %s", err.Error()),
		)
		return
	}

	// Read back the created site to get computed values
	createdSite, err := r.client.GetSite(plan.Code.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading created site",
			fmt.Sprintf("Could not read created site: %s", err.Error()),
		)
		return
	}

	// Update state with computed values
	r.apiToTerraform(ctx, createdSite, &plan, &plan, &resp.Diagnostics)

	// Restore the original order for ship_to_countries if they contain the same elements
	if len(originalShipToCountries) > 0 && createdSite.ShipToCountries != nil {
		if sameElements(originalShipToCountries, createdSite.ShipToCountries) {
			orderedList, d := types.ListValueFrom(ctx, types.StringType, originalShipToCountries)
			resp.Diagnostics.Append(d...)
			plan.ShipToCountries = orderedList
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SiteSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SiteSettingsResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	site, err := r.client.GetSite(state.Code.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading site",
			fmt.Sprintf("Could not read site %s: %s", state.Code.ValueString(), err.Error()),
		)
		return
	}

	if site == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Store the original ship_to_countries order and tax_calculation_address_type from state
	var originalShipToCountries []string
	var originalTaxCalculationType types.String
	if !state.ShipToCountries.IsNull() {
		resp.Diagnostics.Append(state.ShipToCountries.ElementsAs(ctx, &originalShipToCountries, false)...)
	}
	originalTaxCalculationType = state.TaxCalculationAddressType

	r.apiToTerraform(ctx, site, &state, &state, &resp.Diagnostics)

	// Restore the original order for ship_to_countries if they contain the same elements
	if len(originalShipToCountries) > 0 && site.ShipToCountries != nil {
		if sameElements(originalShipToCountries, site.ShipToCountries) {
			orderedList, d := types.ListValueFrom(ctx, types.StringType, originalShipToCountries)
			resp.Diagnostics.Append(d...)
			state.ShipToCountries = orderedList
		}
	}

	// Restore tax_calculation_address_type from state (API doesn't return this value)
	if !originalTaxCalculationType.IsNull() {
		state.TaxCalculationAddressType = originalTaxCalculationType
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Helper function to check if two slices contain the same elements (order-independent)
func sameElements(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	counts := make(map[string]int)
	for _, v := range a {
		counts[v]++
	}

	for _, v := range b {
		counts[v]--
		if counts[v] < 0 {
			return false
		}
	}

	for _, count := range counts {
		if count != 0 {
			return false
		}
	}

	return true
}

func (r *SiteSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SiteSettingsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Store the original ship_to_countries order from plan
	var originalShipToCountries []string
	if !plan.ShipToCountries.IsNull() {
		resp.Diagnostics.Append(plan.ShipToCountries.ElementsAs(ctx, &originalShipToCountries, false)...)
	}

	// Convert Terraform model to API model
	site, diags := r.terraformToApi(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the site
	err := r.client.UpdateSite(plan.Code.ValueString(), site)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating site",
			fmt.Sprintf("Could not update site: %s", err.Error()),
		)
		return
	}

	// Read back the updated site
	updatedSite, err := r.client.GetSite(plan.Code.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading updated site",
			fmt.Sprintf("Could not read updated site: %s", err.Error()),
		)
		return
	}

	r.apiToTerraform(ctx, updatedSite, &plan, &plan, &resp.Diagnostics)

	// Restore the original order for ship_to_countries if they contain the same elements
	if len(originalShipToCountries) > 0 && updatedSite.ShipToCountries != nil {
		if sameElements(originalShipToCountries, updatedSite.ShipToCountries) {
			orderedList, d := types.ListValueFrom(ctx, types.StringType, originalShipToCountries)
			resp.Diagnostics.Append(d...)
			plan.ShipToCountries = orderedList
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SiteSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SiteSettingsResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSite(state.Code.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting site",
			fmt.Sprintf("Could not delete site: %s", err.Error()),
		)
		return
	}
}

func (r *SiteSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("code"), req, resp)
}
