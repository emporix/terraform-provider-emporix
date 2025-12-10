package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CountryResource{}
var _ resource.ResourceWithImportState = &CountryResource{}

func NewCountryResource() resource.Resource {
	return &CountryResource{}
}

// CountryResource defines the resource implementation.
type CountryResource struct {
	client *EmporixClient
}

// CountryResourceModel describes the resource data model.
type CountryResourceModel struct {
	Code    types.String `tfsdk:"code"`
	Name    types.Map    `tfsdk:"name"`
	Regions types.List   `tfsdk:"regions"`
	Active  types.Bool   `tfsdk:"active"`
}

// mapCountryToModel converts a Country API response to a CountryResourceModel
func mapCountryToModel(ctx context.Context, country *Country, data *CountryResourceModel, diags *diag.Diagnostics) {
	data.Code = types.StringValue(country.Code)
	data.Active = types.BoolValue(country.Active)

	// Convert name map to Terraform map
	if country.Name != nil {
		nameMapValue, d := types.MapValueFrom(ctx, types.StringType, country.Name)
		diags.Append(d...)
		data.Name = nameMapValue
	}

	// Convert regions to Terraform list
	regions := country.Regions
	if regions == nil {
		regions = []string{}
	}
	regionList, d := types.ListValueFrom(ctx, types.StringType, regions)
	diags.Append(d...)
	data.Regions = regionList
}

func (r *CountryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_country"
}

func (r *CountryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a country's active status in Emporix. " +
			"Countries are pre-populated by Emporix. " +
			"When you add this resource, it automatically adopts the existing country and allows you to manage its active status. " +
			"When you remove the resource from Terraform, the country is deactivated (active = false).",

		Attributes: map[string]schema.Attribute{
			"code": schema.StringAttribute{
				MarkdownDescription: "Country code (ISO 3166-1 alpha-2, 2-letter code). Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.MapAttribute{
				MarkdownDescription: "Localized country names (read-only).",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"regions": schema.ListAttribute{
				MarkdownDescription: "Regions the country belongs to (read-only).",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether the country is active for the tenant. Only active countries are visible in the system. Defaults to true.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
		},
	}
}

func (r *CountryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CountryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CountryResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating (adopting) country", map[string]interface{}{
		"code":   data.Code.ValueString(),
		"active": data.Active.ValueBool(),
	})

	// Countries are pre-populated by Emporix, so we "adopt" the existing country
	// First, fetch the current state
	country, err := r.client.GetCountry(ctx, data.Code.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read country, got error: %s", err))
		return
	}

	// If user specified active status and it differs from current, update it
	if !data.Active.IsNull() && !data.Active.IsUnknown() && data.Active.ValueBool() != country.Active {
		tflog.Debug(ctx, "Updating country active status during create", map[string]interface{}{
			"code":        data.Code.ValueString(),
			"from_active": country.Active,
			"to_active":   data.Active.ValueBool(),
		})

		active := data.Active.ValueBool()
		updateData := &CountryUpdate{
			Active: &active,
		}

		country, err = r.client.UpdateCountry(ctx, data.Code.ValueString(), updateData)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update country, got error: %s", err))
			return
		}
	}

	// Map API response to model
	mapCountryToModel(ctx, country, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CountryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CountryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading country", map[string]interface{}{
		"code": data.Code.ValueString(),
	})

	// Get country from API
	country, err := r.client.GetCountry(ctx, data.Code.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read country, got error: %s", err))
		return
	}

	// Map API response to model
	mapCountryToModel(ctx, country, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CountryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CountryResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating country", map[string]interface{}{
		"code":   data.Code.ValueString(),
		"active": data.Active.ValueBool(),
	})

	// Prepare update payload (only active field can be updated)
	active := data.Active.ValueBool()
	updateData := &CountryUpdate{
		Active: &active,
	}

	// Update country via API
	country, err := r.client.UpdateCountry(ctx, data.Code.ValueString(), updateData)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update country, got error: %s", err))
		return
	}

	// Map updated response to model
	mapCountryToModel(ctx, country, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CountryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CountryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deactivating country", map[string]interface{}{
		"code": data.Code.ValueString(),
	})

	// Deactivate the country (set active = false)
	active := false
	updateData := &CountryUpdate{
		Active: &active,
	}

	_, err := r.client.UpdateCountry(ctx, data.Code.ValueString(), updateData)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to deactivate country, got error: %s", err))
		return
	}

	// Country is now deactivated and will be removed from Terraform state
}

func (r *CountryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by country code (e.g., "US", "GB", "DE")
	resource.ImportStatePassthroughID(ctx, path.Root("code"), req, resp)
}
