package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CurrencyResource{}
var _ resource.ResourceWithImportState = &CurrencyResource{}

func NewCurrencyResource() resource.Resource {
	return &CurrencyResource{}
}

// CurrencyResource defines the resource implementation.
type CurrencyResource struct {
	client *EmporixClient
}

// CurrencyResourceModel describes the resource data model.
type CurrencyResourceModel struct {
	Code types.String `tfsdk:"code"`
	Name types.Map    `tfsdk:"name"`
}

// mapCurrencyToModel converts a Currency API response to a CurrencyResourceModel
func mapCurrencyToModel(ctx context.Context, currency *Currency, data *CurrencyResourceModel, diags *diag.Diagnostics) {
	data.Code = types.StringValue(currency.Code)

	// Convert name map from API to Terraform map
	if currency.Name != nil && len(currency.Name) > 0 {
		nameMapValue, d := types.MapValueFrom(ctx, types.StringType, currency.Name)
		diags.Append(d...)
		data.Name = nameMapValue
	} else {
		// If no name in response, set to empty map
		nameMapValue, d := types.MapValueFrom(ctx, types.StringType, map[string]string{})
		diags.Append(d...)
		data.Name = nameMapValue
	}
}

func (r *CurrencyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_currency"
}

func (r *CurrencyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a currency in Emporix. " +
			"The currency code must be compliant with ISO-4217 standard (https://www.iso.org/iso-4217-currency-codes.html). " +
			"The currency code is immutable and cannot be changed after creation.",

		Attributes: map[string]schema.Attribute{
			"code": schema.StringAttribute{
				MarkdownDescription: "Currency code (3-letter uppercase ISO-4217 code, e.g., USD, EUR, GBP). Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.MapAttribute{
				MarkdownDescription: "Currency name as a map of language code to name (e.g., {\"en\": \"US Dollar\", \"de\": \"US-Dollar\"}). " +
					"Provide at least one language translation.",
				ElementType: types.StringType,
				Required:    true,
			},
		},
	}
}

func (r *CurrencyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*EmporixClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *EmporixClient, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *CurrencyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CurrencyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating currency", map[string]interface{}{
		"code": data.Code.ValueString(),
	})

	// Get name translations from user
	nameMap := make(map[string]string)
	resp.Diagnostics.Append(data.Name.ElementsAs(ctx, &nameMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create currency via API
	currencyCreate := &CurrencyCreate{
		Code: data.Code.ValueString(),
		Name: nameMap,
	}

	currency, err := r.client.CreateCurrency(ctx, currencyCreate)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create currency, got error: %s", err))
		return
	}

	// Map API response to model
	mapCurrencyToModel(ctx, currency, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CurrencyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CurrencyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading currency", map[string]interface{}{
		"code": data.Code.ValueString(),
	})

	// Get currency from API
	currency, err := r.client.GetCurrency(ctx, data.Code.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read currency, got error: %s", err))
		return
	}

	// Map API response to model
	mapCurrencyToModel(ctx, currency, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CurrencyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CurrencyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating currency", map[string]interface{}{
		"code": data.Code.ValueString(),
	})

	// Get name translations from user
	nameMap := make(map[string]string)
	resp.Diagnostics.Append(data.Name.ElementsAs(ctx, &nameMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare update payload
	updateData := &CurrencyUpdate{
		Name: nameMap,
	}

	// Update currency via API
	currency, err := r.client.UpdateCurrency(ctx, data.Code.ValueString(), updateData)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update currency, got error: %s", err))
		return
	}

	// Map updated response to model
	mapCurrencyToModel(ctx, currency, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CurrencyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CurrencyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting currency", map[string]interface{}{
		"code": data.Code.ValueString(),
	})

	// Delete currency via API
	err := r.client.DeleteCurrency(ctx, data.Code.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete currency, got error: %s", err),
		)
		return
	}

	// Currency is now deleted and will be removed from Terraform state
}

func (r *CurrencyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by currency code (e.g., "USD", "EUR", "GBP")
	resource.ImportStatePassthroughID(ctx, path.Root("code"), req, resp)
}
