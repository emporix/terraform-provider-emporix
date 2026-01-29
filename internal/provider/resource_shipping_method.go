package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &ShippingMethodResource{}
	_ resource.ResourceWithConfigure   = &ShippingMethodResource{}
	_ resource.ResourceWithImportState = &ShippingMethodResource{}
)

func NewShippingMethodResource() resource.Resource {
	return &ShippingMethodResource{}
}

type ShippingMethodResource struct {
	client *EmporixClient
}

type ShippingMethodResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Site            types.String `tfsdk:"site"`
	ZoneID          types.String `tfsdk:"zone_id"`
	Name            types.Map    `tfsdk:"name"`
	Active          types.Bool   `tfsdk:"active"`
	MaxOrderValue   types.Object `tfsdk:"max_order_value"`
	Fees            types.List   `tfsdk:"fees"`
	ShippingTaxCode types.String `tfsdk:"shipping_tax_code"`
	ShippingGroupID types.String `tfsdk:"shipping_group_id"`
}

type MonetaryAmountModel struct {
	Amount   types.Float64 `tfsdk:"amount"`
	Currency types.String  `tfsdk:"currency"`
}

type ShippingFeeModel struct {
	MinOrderValue   types.Object `tfsdk:"min_order_value"`
	Cost            types.Object `tfsdk:"cost"`
	ShippingGroupID types.String `tfsdk:"shipping_group_id"`
}

// API structs for ShippingMethod

// ShippingMethod represents a shipping method
type ShippingMethod struct {
	ID              string            `json:"id"`
	Name            map[string]string `json:"name"` // Localized name map
	Active          bool              `json:"active"`
	MaxOrderValue   *MonetaryAmount   `json:"maxOrderValue,omitempty"`
	Fees            []ShippingFee     `json:"fees"`
	ShippingTaxCode string            `json:"shippingTaxCode,omitempty"`
	ShippingGroupID string            `json:"shippingGroupId,omitempty"`
}

// UnmarshalJSON handles the Name field which can be string or map[string]string from the API
func (sm *ShippingMethod) UnmarshalJSON(data []byte) error {
	// Define a raw struct with Name as json.RawMessage
	type shippingMethodRaw struct {
		ID              string          `json:"id"`
		Name            json.RawMessage `json:"name"`
		Active          bool            `json:"active"`
		MaxOrderValue   *MonetaryAmount `json:"maxOrderValue,omitempty"`
		Fees            []ShippingFee   `json:"fees"`
		ShippingTaxCode string          `json:"shippingTaxCode,omitempty"`
		ShippingGroupID string          `json:"shippingGroupId,omitempty"`
	}

	var raw shippingMethodRaw
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Copy all fields except Name
	sm.ID = raw.ID
	sm.Active = raw.Active
	sm.MaxOrderValue = raw.MaxOrderValue
	sm.Fees = raw.Fees
	sm.ShippingTaxCode = raw.ShippingTaxCode
	sm.ShippingGroupID = raw.ShippingGroupID

	// Handle Name field - try as string first, then as map
	var nameStr string
	if err := json.Unmarshal(raw.Name, &nameStr); err == nil {
		// It's a string - convert to map
		sm.Name = map[string]string{"en": nameStr}
		return nil
	}

	// Try as map[string]string
	var nameMap map[string]string
	if err := json.Unmarshal(raw.Name, &nameMap); err == nil {
		sm.Name = nameMap
		return nil
	}

	// Fallback: empty map
	sm.Name = map[string]string{}
	return nil
}

// ShippingFee represents a shipping fee configuration
type ShippingFee struct {
	MinOrderValue   *MonetaryAmount `json:"minOrderValue"`
	Cost            *MonetaryAmount `json:"cost"`
	ShippingGroupID string          `json:"shippingGroupId,omitempty"`
}

// MonetaryAmount represents an amount of money
type MonetaryAmount struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

func (r *ShippingMethodResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_shipping_method"
}

func (r *ShippingMethodResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages shipping methods for a shipping zone. Shipping methods define delivery options (e.g., standard, express) with associated costs and rules.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Shipping method identifier.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"site": schema.StringAttribute{
				MarkdownDescription: "Site code (typically 'main' for single-shop tenants).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"zone_id": schema.StringAttribute{
				MarkdownDescription: "Shipping zone ID this method belongs to. Must reference an existing shipping zone.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.MapAttribute{
				MarkdownDescription: "Localized names for the shipping method (e.g., {\"en\": \"Standard Shipping\", \"de\": \"Standardversand\"}).",
				ElementType:         types.StringType,
				Required:            true,
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether the shipping method is active. Defaults to true.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"max_order_value": schema.SingleNestedAttribute{
				MarkdownDescription: "Maximum order value for this shipping method. Orders above this value cannot use this method.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"amount": schema.Float64Attribute{
						MarkdownDescription: "Amount value.",
						Required:            true,
					},
					"currency": schema.StringAttribute{
						MarkdownDescription: "Currency code (e.g., 'USD', 'EUR').",
						Required:            true,
					},
				},
			},
			"fees": schema.ListNestedAttribute{
				MarkdownDescription: "Shipping fee tiers based on order value. Multiple tiers can be defined for different order value ranges.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"min_order_value": schema.SingleNestedAttribute{
							MarkdownDescription: "Minimum order value for this fee tier.",
							Required:            true,
							Attributes: map[string]schema.Attribute{
								"amount": schema.Float64Attribute{
									MarkdownDescription: "Amount value.",
									Required:            true,
								},
								"currency": schema.StringAttribute{
									MarkdownDescription: "Currency code (e.g., 'USD', 'EUR').",
									Required:            true,
								},
							},
						},
						"cost": schema.SingleNestedAttribute{
							MarkdownDescription: "Shipping cost for this tier.",
							Required:            true,
							Attributes: map[string]schema.Attribute{
								"amount": schema.Float64Attribute{
									MarkdownDescription: "Amount value.",
									Required:            true,
								},
								"currency": schema.StringAttribute{
									MarkdownDescription: "Currency code (e.g., 'USD', 'EUR').",
									Required:            true,
								},
							},
						},
						"shipping_group_id": schema.StringAttribute{
							MarkdownDescription: "Optional shipping group ID for this fee tier.",
							Optional:            true,
						},
					},
				},
			},
			"shipping_tax_code": schema.StringAttribute{
				MarkdownDescription: "Tax code for shipping fees.",
				Optional:            true,
			},
			"shipping_group_id": schema.StringAttribute{
				MarkdownDescription: "Shipping group ID to associate with this method.",
				Optional:            true,
			},
		},
	}
}

func (r *ShippingMethodResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ShippingMethodResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ShippingMethodResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	apiMethod, diags := r.toAPIModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create shipping method via API
	createdMethod, err := r.client.CreateShippingMethod(ctx, data.Site.ValueString(), data.ZoneID.ValueString(), apiMethod)
	if err != nil {
		resp.Diagnostics.AddError("Error creating shipping method", err.Error())
		return
	}

	tflog.Debug(ctx, "Created shipping method", map[string]interface{}{
		"id":      createdMethod.ID,
		"site":    data.Site.ValueString(),
		"zone_id": data.ZoneID.ValueString(),
	})

	// Use the ID from the API response (may differ from requested ID)
	createdID := data.ID.ValueString()
	if createdMethod != nil && createdMethod.ID != "" {
		createdID = createdMethod.ID
	}

	// Read back the created resource using the actual ID from API
	actualMethod, err := r.client.GetShippingMethod(ctx, data.Site.ValueString(), data.ZoneID.ValueString(), createdID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading created shipping method", err.Error())
		return
	}

	// Convert API model to Terraform state
	var stateModel ShippingMethodResourceModel
	stateModel.Site = data.Site
	stateModel.ZoneID = data.ZoneID
	stateModel.ID = types.StringValue(createdID) // Use the actual ID from API
	r.syncModelFromAPI(ctx, &stateModel, actualMethod, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &stateModel)...)
}

func (r *ShippingMethodResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ShippingMethodResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	method, err := r.client.GetShippingMethod(ctx, data.Site.ValueString(), data.ZoneID.ValueString(), data.ID.ValueString())
	if err != nil {
		if _, ok := err.(*NotFoundError); ok {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading shipping method", err.Error())
		return
	}

	var stateModel ShippingMethodResourceModel
	stateModel.Site = data.Site
	stateModel.ZoneID = data.ZoneID
	r.syncModelFromAPI(ctx, &stateModel, method, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &stateModel)...)
}

func (r *ShippingMethodResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ShippingMethodResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiMethod, diags := r.toAPIModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UpdateShippingMethod(ctx, data.Site.ValueString(), data.ZoneID.ValueString(), data.ID.ValueString(), apiMethod)
	if err != nil {
		resp.Diagnostics.AddError("Error updating shipping method", err.Error())
		return
	}

	// Read back updated resource
	actualMethod, err := r.client.GetShippingMethod(ctx, data.Site.ValueString(), data.ZoneID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading updated shipping method", err.Error())
		return
	}

	var stateModel ShippingMethodResourceModel
	stateModel.Site = data.Site
	stateModel.ZoneID = data.ZoneID
	r.syncModelFromAPI(ctx, &stateModel, actualMethod, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &stateModel)...)
}

func (r *ShippingMethodResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ShippingMethodResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteShippingMethod(ctx, data.Site.ValueString(), data.ZoneID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting shipping method", err.Error())
		return
	}
}

func (r *ShippingMethodResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: "site:zone_id:method_id"
	// Example: "main:zone-us:standard-shipping"
	parts := strings.Split(req.ID, ":")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'site:zone_id:method_id', got: %s", req.ID),
		)
		return
	}

	site := parts[0]
	zoneID := parts[1]
	methodID := parts[2]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site"), site)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone_id"), zoneID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), methodID)...)
}

// Helper functions

func (r *ShippingMethodResource) toAPIModel(ctx context.Context, model *ShippingMethodResourceModel) (*ShippingMethod, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiMethod := &ShippingMethod{
		ID:     model.ID.ValueString(),
		Active: model.Active.ValueBool(),
	}

	// Convert name map
	nameMap := make(map[string]string)
	d := model.Name.ElementsAs(ctx, &nameMap, false)
	diags.Append(d...)
	apiMethod.Name = nameMap

	// Convert max order value
	if !model.MaxOrderValue.IsNull() && !model.MaxOrderValue.IsUnknown() {
		var maxOrderValue MonetaryAmountModel
		d := model.MaxOrderValue.As(ctx, &maxOrderValue, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		apiMethod.MaxOrderValue = &MonetaryAmount{
			Amount:   maxOrderValue.Amount.ValueFloat64(),
			Currency: maxOrderValue.Currency.ValueString(),
		}
	}

	// Convert fees
	if !model.Fees.IsNull() && !model.Fees.IsUnknown() {
		var feeModels []ShippingFeeModel
		d = model.Fees.ElementsAs(ctx, &feeModels, false)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		apiFees := make([]ShippingFee, 0, len(feeModels))
		for _, feeModel := range feeModels {
			var minOrderValue MonetaryAmountModel
			d = feeModel.MinOrderValue.As(ctx, &minOrderValue, basetypes.ObjectAsOptions{})
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}

			var cost MonetaryAmountModel
			d = feeModel.Cost.As(ctx, &cost, basetypes.ObjectAsOptions{})
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}

			apiFee := ShippingFee{
				MinOrderValue: &MonetaryAmount{
					Amount:   minOrderValue.Amount.ValueFloat64(),
					Currency: minOrderValue.Currency.ValueString(),
				},
				Cost: &MonetaryAmount{
					Amount:   cost.Amount.ValueFloat64(),
					Currency: cost.Currency.ValueString(),
				},
			}

			if !feeModel.ShippingGroupID.IsNull() {
				apiFee.ShippingGroupID = feeModel.ShippingGroupID.ValueString()
			}

			apiFees = append(apiFees, apiFee)
		}
		apiMethod.Fees = apiFees
	}

	if !model.ShippingTaxCode.IsNull() {
		apiMethod.ShippingTaxCode = model.ShippingTaxCode.ValueString()
	}

	if !model.ShippingGroupID.IsNull() {
		apiMethod.ShippingGroupID = model.ShippingGroupID.ValueString()
	}

	return apiMethod, diags
}

func (r *ShippingMethodResource) syncModelFromAPI(ctx context.Context, model *ShippingMethodResourceModel, api *ShippingMethod, diags *diag.Diagnostics) {
	model.ID = types.StringValue(api.ID)
	model.Active = types.BoolValue(api.Active)

	// Handle name (always map[string]string after unmarshalling)
	if api.Name != nil && len(api.Name) > 0 {
		nameValue, d := types.MapValueFrom(ctx, types.StringType, api.Name)
		diags.Append(d...)
		model.Name = nameValue
	} else {
		model.Name = types.MapNull(types.StringType)
	}

	// Max order value
	if api.MaxOrderValue != nil {
		maxOrderValue := MonetaryAmountModel{
			Amount:   types.Float64Value(api.MaxOrderValue.Amount),
			Currency: types.StringValue(api.MaxOrderValue.Currency),
		}
		maxOrderValueObj, d := types.ObjectValueFrom(ctx, maxOrderValue.AttributeTypes(), maxOrderValue)
		diags.Append(d...)
		model.MaxOrderValue = maxOrderValueObj
	} else {
		model.MaxOrderValue = types.ObjectNull(MonetaryAmountModel{}.AttributeTypes())
	}

	// Fees
	feeModels := make([]ShippingFeeModel, 0, len(api.Fees))
	for _, apiFee := range api.Fees {
		var minOrderValueObj types.Object
		if apiFee.MinOrderValue != nil {
			minOrderValue := MonetaryAmountModel{
				Amount:   types.Float64Value(apiFee.MinOrderValue.Amount),
				Currency: types.StringValue(apiFee.MinOrderValue.Currency),
			}
			var d diag.Diagnostics
			minOrderValueObj, d = types.ObjectValueFrom(ctx, minOrderValue.AttributeTypes(), minOrderValue)
			diags.Append(d...)
		} else {
			minOrderValueObj = types.ObjectNull(MonetaryAmountModel{}.AttributeTypes())
		}

		var costObj types.Object
		if apiFee.Cost != nil {
			cost := MonetaryAmountModel{
				Amount:   types.Float64Value(apiFee.Cost.Amount),
				Currency: types.StringValue(apiFee.Cost.Currency),
			}
			var d diag.Diagnostics
			costObj, d = types.ObjectValueFrom(ctx, cost.AttributeTypes(), cost)
			diags.Append(d...)
		} else {
			costObj = types.ObjectNull(MonetaryAmountModel{}.AttributeTypes())
		}

		feeModel := ShippingFeeModel{
			MinOrderValue: minOrderValueObj,
			Cost:          costObj,
		}

		if apiFee.ShippingGroupID != "" {
			feeModel.ShippingGroupID = types.StringValue(apiFee.ShippingGroupID)
		} else {
			feeModel.ShippingGroupID = types.StringNull()
		}

		feeModels = append(feeModels, feeModel)
	}

	feesList, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: ShippingFeeModel{}.AttributeTypes()}, feeModels)
	diags.Append(d...)
	model.Fees = feesList

	if api.ShippingTaxCode != "" {
		model.ShippingTaxCode = types.StringValue(api.ShippingTaxCode)
	} else {
		model.ShippingTaxCode = types.StringNull()
	}

	if api.ShippingGroupID != "" {
		model.ShippingGroupID = types.StringValue(api.ShippingGroupID)
	} else {
		model.ShippingGroupID = types.StringNull()
	}
}

// AttributeTypes helper methods
func (m MonetaryAmountModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"amount":   types.Float64Type,
		"currency": types.StringType,
	}
}

func (m ShippingFeeModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"min_order_value":   types.ObjectType{AttrTypes: MonetaryAmountModel{}.AttributeTypes()},
		"cost":              types.ObjectType{AttrTypes: MonetaryAmountModel{}.AttributeTypes()},
		"shipping_group_id": types.StringType,
	}
}
