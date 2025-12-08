package provider

import (
	"context"
	"fmt"

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

func (r *CountryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_country"
}

func (r *CountryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a country's active status in Emporix. " +
			"**Important**: Countries are pre-populated by Emporix and cannot be created or deleted. " +
			"This resource only manages the `active` field. You must import existing countries before managing them.",

		Attributes: map[string]schema.Attribute{
			"code": schema.StringAttribute{
				MarkdownDescription: "Country code (ISO 3166-1 alpha-2, 2-letter code). Cannot be changed after import.",
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
				MarkdownDescription: "Whether the country is active for the tenant. Only active countries are visible in the system.",
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
	// Countries cannot be created - they must be imported
	resp.Diagnostics.AddError(
		"Cannot Create Country",
		"Countries are pre-populated by Emporix and cannot be created. "+
			"Please import an existing country using: terraform import emporix_country.<name> <country_code>\n\n"+
			"Example: terraform import emporix_country.usa US",
	)
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
	data.Code = types.StringValue(country.Code)
	data.Active = types.BoolValue(country.Active)

	// Convert name map to Terraform map
	if country.Name != nil {
		nameMap := make(map[string]types.String)
		for k, v := range country.Name {
			nameMap[k] = types.StringValue(v)
		}
		nameMapValue, diags := types.MapValue(types.StringType, map[string]types.Value{})
		resp.Diagnostics.Append(diags...)
		for k, v := range nameMap {
			nameMapValue.Elements()[k] = v
		}
		// Recreate the map properly
		nameElements := make(map[string]types.Value)
		for k, v := range nameMap {
			nameElements[k] = v
		}
		nameMapValue, diags = types.MapValue(types.StringType, nameElements)
		resp.Diagnostics.Append(diags...)
		data.Name = nameMapValue
	}

	// Convert regions to Terraform list
	if country.Regions != nil {
		regionElements := make([]types.Value, len(country.Regions))
		for i, region := range country.Regions {
			regionElements[i] = types.StringValue(region)
		}
		regionList, diags := types.ListValue(types.StringType, regionElements)
		resp.Diagnostics.Append(diags...)
		data.Regions = regionList
	} else {
		regionList, diags := types.ListValue(types.StringType, []types.Value{})
		resp.Diagnostics.Append(diags...)
		data.Regions = regionList
	}

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
	data.Active = types.BoolValue(country.Active)

	// Convert name map to Terraform map
	if country.Name != nil {
		nameElements := make(map[string]types.Value)
		for k, v := range country.Name {
			nameElements[k] = types.StringValue(v)
		}
		nameMapValue, diags := types.MapValue(types.StringType, nameElements)
		resp.Diagnostics.Append(diags...)
		data.Name = nameMapValue
	}

	// Convert regions to Terraform list
	if country.Regions != nil {
		regionElements := make([]types.Value, len(country.Regions))
		for i, region := range country.Regions {
			regionElements[i] = types.StringValue(region)
		}
		regionList, diags := types.ListValue(types.StringType, regionElements)
		resp.Diagnostics.Append(diags...)
		data.Regions = regionList
	} else {
		regionList, diags := types.ListValue(types.StringType, []types.Value{})
		resp.Diagnostics.Append(diags...)
		data.Regions = regionList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CountryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Countries cannot be deleted - just remove from state
	var data CountryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Country removed from Terraform state (countries cannot be deleted from Emporix)", map[string]interface{}{
		"code": data.Code.ValueString(),
	})

	// Note: We don't call any API here because countries cannot be deleted
	// Terraform will simply remove it from state
}

func (r *CountryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by country code (e.g., "US", "GB", "DE")
	resource.ImportStatePassthroughID(ctx, path.Root("code"), req, resp)
}
