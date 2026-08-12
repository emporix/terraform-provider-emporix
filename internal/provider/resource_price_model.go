package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &PriceModelResource{}
var _ resource.ResourceWithImportState = &PriceModelResource{}
var _ resource.ResourceWithModifyPlan = &PriceModelResource{}

func NewPriceModelResource() resource.Resource {
	return &PriceModelResource{}
}

// PriceModelResource defines the resource implementation
type PriceModelResource struct {
	client *EmporixClient
}

// PriceModelResourceModel describes the resource data model
type PriceModelResourceModel struct {
	ID              types.String `tfsdk:"id"`
	IncludesTax     types.Bool   `tfsdk:"includes_tax"`
	IncludesMarkup  types.Bool   `tfsdk:"includes_markup"`
	Default         types.Bool   `tfsdk:"default"`
	Name            types.Map    `tfsdk:"name"`
	Description     types.Map    `tfsdk:"description"`
	TierDefinition  types.Object `tfsdk:"tier_definition"`
	MeasurementUnit types.Object `tfsdk:"measurement_unit"`
	ForceDelete     types.Bool   `tfsdk:"force_delete"`
}

// TierDefinitionModel represents the tier_definition nested attribute
type TierDefinitionModel struct {
	TierType types.String `tfsdk:"tier_type"`
	Tiers    types.List   `tfsdk:"tiers"`
}

// TierModel represents a single tier within the tier_definition.tiers list
type TierModel struct {
	ID          types.String `tfsdk:"id"`
	MinQuantity types.Object `tfsdk:"min_quantity"`
}

// MinQuantityModel represents the min_quantity nested attribute of a tier
type MinQuantityModel struct {
	Quantity types.Float64 `tfsdk:"quantity"`
	UnitCode types.String  `tfsdk:"unit_code"`
}

// MeasurementUnitModel represents the measurement_unit nested attribute
type MeasurementUnitModel struct {
	Quantity types.Float64 `tfsdk:"quantity"`
	UnitCode types.String  `tfsdk:"unit_code"`
}

func priceModelMinQuantityAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"quantity":  types.Float64Type,
		"unit_code": types.StringType,
	}
}

func priceModelTierAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":           types.StringType,
		"min_quantity": types.ObjectType{AttrTypes: priceModelMinQuantityAttrTypes()},
	}
}

func priceModelTierDefinitionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"tier_type": types.StringType,
		"tiers":     types.ListType{ElemType: types.ObjectType{AttrTypes: priceModelTierAttrTypes()}},
	}
}

func priceModelMeasurementUnitAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"quantity":  types.Float64Type,
		"unit_code": types.StringType,
	}
}

func (r *PriceModelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_price_model"
}

func (r *PriceModelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Emporix price model, which defines a pricing strategy (basic, volume, or tiered) " +
			"that prices can be assigned to. See the [Price Models API](https://developer.emporix.io/api-references-1/readme/api-reference-26/price-models) for details.\n\n" +
			"**Known upstream issue:** the Price Models API has a confirmed read-after-write consistency bug (reported to Emporix, independent of this provider) " +
			"where a GET immediately following a successful update can briefly reflect the change and then revert to pre-update data on a later read - the reversion " +
			"includes the API's own `metadata.version`, so it is not a caching artifact on this provider's side. This can cause `terraform apply` to report " +
			"\"Provider produced inconsistent result after apply\" on `emporix_price_model` updates (including tier changes) even though the update was sent correctly. " +
			"There is no workaround on the provider side; re-running `apply` after the API's data settles typically resolves it.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the price model. Generated automatically if not provided. Cannot be changed after creation.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"includes_tax": schema.BoolAttribute{
				MarkdownDescription: "Whether prices calculated with this price model are gross (`true`, tax included) or net (`false`, tax excluded).",
				Required:            true,
			},
			"includes_markup": schema.BoolAttribute{
				MarkdownDescription: "Whether the price model operates in markup preview mode. The API requires this field to be non-null, so it defaults to `false` when not set.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"default": schema.BoolAttribute{
				MarkdownDescription: "Whether this is the tenant's default price model. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"name": schema.MapAttribute{
				MarkdownDescription: "Price model name as a map of language codes to translated names. Example: `{en = \"Standard Pricing\"}`. At least one language is required.",
				ElementType:         types.StringType,
				Required:            true,
			},
			"description": schema.MapAttribute{
				MarkdownDescription: "Optional price model description as a map of language codes to translated descriptions.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"tier_definition": schema.SingleNestedAttribute{
				MarkdownDescription: "Defines the pricing strategy and quantity tiers for this price model.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"tier_type": schema.StringAttribute{
						MarkdownDescription: "Pricing strategy. One of `BASIC` (flat price regardless of quantity), `VOLUME` (price depends on the total ordered quantity), " +
							"or `TIERED` (price calculated per tier range).",
						Required: true,
						Validators: []validator.String{
							stringvalidator.OneOf("BASIC", "VOLUME", "TIERED"),
						},
					},
					"tiers": schema.ListNestedAttribute{
						MarkdownDescription: "Quantity tiers for this price model. `BASIC` requires exactly one tier with `min_quantity.quantity = 0`. " +
							"`VOLUME`/`TIERED` require unique, ascending `min_quantity` values, with the first tier starting at `0`.",
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									MarkdownDescription: "Tier identifier, assigned automatically by the API. Shown as `(known after apply)` whenever the tier list changes, " +
										"since Terraform's position-based tracking can't reliably predict which tier keeps which id across insertions or removals - " +
										"the provider resolves the real id-to-tier matching at apply time.",
									Optional: true,
									Computed: true,
								},
								"min_quantity": schema.SingleNestedAttribute{
									MarkdownDescription: "Minimum ordered quantity from which this tier applies.",
									Required:            true,
									Attributes: map[string]schema.Attribute{
										"quantity": schema.Float64Attribute{
											MarkdownDescription: "Minimum quantity value. Must be zero or a positive value.",
											Required:            true,
										},
										"unit_code": schema.StringAttribute{
											MarkdownDescription: "Unit code for the quantity. Must match across all tiers of the same price model.",
											Required:            true,
										},
									},
								},
							},
						},
					},
				},
			},
			"measurement_unit": schema.SingleNestedAttribute{
				MarkdownDescription: "Measurement unit that this price model's tiers are expressed in. Required by the API even for `BASIC` price models.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"quantity": schema.Float64Attribute{
						MarkdownDescription: "Measurement unit quantity. Must be zero or a positive value.",
						Required:            true,
					},
					"unit_code": schema.StringAttribute{
						MarkdownDescription: "Measurement unit code.",
						Required:            true,
					},
				},
			},
			"force_delete": schema.BoolAttribute{
				MarkdownDescription: "If `true`, deleting this resource also deletes (asynchronously) all prices assigned to this price model. " +
					"Requires the `price.pricemodel_manage_admin` scope. Defaults to `false`.",
				Optional: true,
			},
		},
	}
}

func (r *PriceModelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ModifyPlan resolves each planned tier's computed "id" by matching it against the prior
// state's tiers by min_quantity identity, instead of the framework's default positional
// behavior (which has no plan modifier to rely on here, since a tier's real id can only be
// tracked by identity, not list position). Without this, Terraform would either reuse a
// wrong id from a shifted list position, or - lacking any plan modifier at all - mark every
// tier's id unknown on every update regardless of whether the tier list actually changed.
// Since this runs before Update(), priceModelToAPI can read the already-correct id straight
// off the plan with no further reconciliation needed at apply time.
func (r *PriceModelResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Nothing to reconcile on create (no prior state) or destroy (no plan).
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var state PriceModelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan PriceModelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.TierDefinition.IsNull() || state.TierDefinition.IsUnknown() ||
		plan.TierDefinition.IsNull() || plan.TierDefinition.IsUnknown() {
		return
	}

	var stateTierDef, planTierDef TierDefinitionModel
	resp.Diagnostics.Append(state.TierDefinition.As(ctx, &stateTierDef, basetypes.ObjectAsOptions{})...)
	resp.Diagnostics.Append(plan.TierDefinition.As(ctx, &planTierDef, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	var stateTiers []TierModel
	resp.Diagnostics.Append(stateTierDef.Tiers.ElementsAs(ctx, &stateTiers, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var planTiers []TierModel
	resp.Diagnostics.Append(planTierDef.Tiers.ElementsAs(ctx, &planTiers, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idByQuantity := make(map[float64]string, len(stateTiers))
	for _, tierModel := range stateTiers {
		if tierModel.ID.IsNull() || tierModel.ID.IsUnknown() {
			continue
		}

		var minQty MinQuantityModel
		d := tierModel.MinQuantity.As(ctx, &minQty, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(d...)
		if d.HasError() {
			continue
		}

		idByQuantity[minQty.Quantity.ValueFloat64()] = tierModel.ID.ValueString()
	}

	changed := false
	for i := range planTiers {
		var minQty MinQuantityModel
		d := planTiers[i].MinQuantity.As(ctx, &minQty, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(d...)
		if d.HasError() {
			continue
		}

		if id, ok := idByQuantity[minQty.Quantity.ValueFloat64()]; ok {
			planTiers[i].ID = types.StringValue(id)
			changed = true
		}
	}

	if !changed {
		return
	}

	newTiersList, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: priceModelTierAttrTypes()}, planTiers)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	planTierDef.Tiers = newTiersList

	newTierDefObj, d := types.ObjectValueFrom(ctx, priceModelTierDefinitionAttrTypes(), planTierDef)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("tier_definition"), newTierDefObj)...)
}

func (r *PriceModelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PriceModelResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	priceModel, diags := priceModelToAPI(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreatePriceModel(ctx, priceModel)
	if err != nil {
		resp.Diagnostics.AddError("Error creating price model", fmt.Sprintf("Could not create price model: %s", err.Error()))
		return
	}

	resp.Diagnostics.Append(priceModelFromAPI(ctx, created, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PriceModelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PriceModelResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	priceModel, err := r.client.GetPriceModel(ctx, state.ID.ValueString())
	if err != nil {
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading price model", fmt.Sprintf("Could not read price model %s: %s", state.ID.ValueString(), err.Error()))
		return
	}

	resp.Diagnostics.Append(priceModelFromAPI(ctx, priceModel, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PriceModelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PriceModelResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	priceModel, diags := priceModelToAPI(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdatePriceModel(ctx, plan.ID.ValueString(), priceModel)
	if err != nil {
		resp.Diagnostics.AddError("Error updating price model", fmt.Sprintf("Could not update price model: %s", err.Error()))
		return
	}

	resp.Diagnostics.Append(priceModelFromAPI(ctx, updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PriceModelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PriceModelResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	forceDelete := !state.ForceDelete.IsNull() && state.ForceDelete.ValueBool()

	err := r.client.DeletePriceModel(ctx, state.ID.ValueString(), forceDelete)
	if err != nil && !IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting price model", fmt.Sprintf("Could not delete price model: %s", err.Error()))
		return
	}
}

func (r *PriceModelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// priceModelToAPI converts a PriceModelResourceModel (Terraform plan) into a PriceModelUpsert (API payload)
func priceModelToAPI(ctx context.Context, model *PriceModelResourceModel) (*PriceModelUpsert, diag.Diagnostics) {
	var diags diag.Diagnostics

	pm := &PriceModelUpsert{
		IncludesTax: model.IncludesTax.ValueBool(),
	}

	if !model.ID.IsNull() && !model.ID.IsUnknown() {
		pm.ID = model.ID.ValueString()
	}

	if !model.IncludesMarkup.IsNull() {
		v := model.IncludesMarkup.ValueBool()
		pm.IncludesMarkup = &v
	}

	if !model.Default.IsNull() {
		v := model.Default.ValueBool()
		pm.Default = &v
	}

	nameMap := make(map[string]string)
	diags.Append(model.Name.ElementsAs(ctx, &nameMap, false)...)
	pm.Name = nameMap

	if !model.Description.IsNull() {
		descMap := make(map[string]string)
		diags.Append(model.Description.ElementsAs(ctx, &descMap, false)...)
		pm.Description = descMap
	}

	if diags.HasError() {
		return nil, diags
	}

	var tierDefModel TierDefinitionModel
	diags.Append(model.TierDefinition.As(ctx, &tierDefModel, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}

	var tierModels []TierModel
	diags.Append(tierDefModel.Tiers.ElementsAs(ctx, &tierModels, false)...)
	if diags.HasError() {
		return nil, diags
	}

	tiers := make([]Tier, len(tierModels))
	for i, tierModel := range tierModels {
		var minQtyModel MinQuantityModel
		diags.Append(tierModel.MinQuantity.As(ctx, &minQtyModel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		tier := Tier{
			MinQuantity: &MinQuantity{
				Quantity: minQtyModel.Quantity.ValueFloat64(),
				UnitCode: minQtyModel.UnitCode.ValueString(),
			},
		}
		if !tierModel.ID.IsNull() && !tierModel.ID.IsUnknown() {
			tier.ID = tierModel.ID.ValueString()
		}
		tiers[i] = tier
	}

	pm.TierDefinition = &TierDefinition{
		TierType: tierDefModel.TierType.ValueString(),
		Tiers:    tiers,
	}

	if !model.MeasurementUnit.IsNull() && !model.MeasurementUnit.IsUnknown() {
		var muModel MeasurementUnitModel
		diags.Append(model.MeasurementUnit.As(ctx, &muModel, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		pm.MeasurementUnit = &MeasurementUnit{
			Quantity: muModel.Quantity.ValueFloat64(),
			UnitCode: muModel.UnitCode.ValueString(),
		}
	}

	return pm, diags
}

// priceModelFromAPI populates a PriceModelResourceModel (Terraform state) from a PriceModel (API response).
// It never touches model.ForceDelete, which is a client-side-only field the API doesn't return.
func priceModelFromAPI(ctx context.Context, pm *PriceModel, model *PriceModelResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(pm.ID)
	model.IncludesTax = types.BoolValue(pm.IncludesTax)

	if pm.IncludesMarkup != nil {
		model.IncludesMarkup = types.BoolValue(*pm.IncludesMarkup)
	} else {
		model.IncludesMarkup = types.BoolNull()
	}

	if pm.Default != nil {
		model.Default = types.BoolValue(*pm.Default)
	} else {
		model.Default = types.BoolValue(false)
	}

	nameMap, d := localizedFieldToMap(pm.Name)
	diags.Append(d...)
	nameMapValue, d := types.MapValueFrom(ctx, types.StringType, nameMap)
	diags.Append(d...)
	model.Name = nameMapValue

	if pm.Description != nil {
		descMap, d := localizedFieldToMap(pm.Description)
		diags.Append(d...)
		descMapValue, d := types.MapValueFrom(ctx, types.StringType, descMap)
		diags.Append(d...)
		model.Description = descMapValue
	} else {
		model.Description = types.MapNull(types.StringType)
	}

	if diags.HasError() {
		return diags
	}

	tierDefObj, d := priceModelTierDefinitionFromAPI(ctx, pm.TierDefinition)
	diags.Append(d...)
	model.TierDefinition = tierDefObj

	if pm.MeasurementUnit != nil {
		muModel := MeasurementUnitModel{
			Quantity: types.Float64Value(pm.MeasurementUnit.Quantity),
			UnitCode: types.StringValue(pm.MeasurementUnit.UnitCode),
		}
		muObj, d := types.ObjectValueFrom(ctx, priceModelMeasurementUnitAttrTypes(), muModel)
		diags.Append(d...)
		model.MeasurementUnit = muObj
	} else {
		model.MeasurementUnit = types.ObjectNull(priceModelMeasurementUnitAttrTypes())
	}

	return diags
}

// priceModelTierDefinitionFromAPI converts a TierDefinition (API) into a types.Object (Terraform)
func priceModelTierDefinitionFromAPI(ctx context.Context, td *TierDefinition) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if td == nil {
		return types.ObjectNull(priceModelTierDefinitionAttrTypes()), diags
	}

	tierModels := make([]TierModel, len(td.Tiers))
	for i, tier := range td.Tiers {
		quantity := 0.0
		unitCode := ""
		if tier.MinQuantity != nil {
			quantity = tier.MinQuantity.Quantity
			unitCode = tier.MinQuantity.UnitCode
		}

		minQtyObj, d := types.ObjectValueFrom(ctx, priceModelMinQuantityAttrTypes(), MinQuantityModel{
			Quantity: types.Float64Value(quantity),
			UnitCode: types.StringValue(unitCode),
		})
		diags.Append(d...)

		tierModels[i] = TierModel{
			ID:          types.StringValue(tier.ID),
			MinQuantity: minQtyObj,
		}
	}

	tiersList, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: priceModelTierAttrTypes()}, tierModels)
	diags.Append(d...)

	tierDefObj, d := types.ObjectValueFrom(ctx, priceModelTierDefinitionAttrTypes(), TierDefinitionModel{
		TierType: types.StringValue(td.TierType),
		Tiers:    tiersList,
	})
	diags.Append(d...)

	return tierDefObj, diags
}

// localizedFieldToMap normalizes a "string or map[string]string" API field (decoded as interface{}
// by encoding/json) into a map[string]string, following the convention used by other localized
// fields in this provider (e.g. TaxClass.Name).
func localizedFieldToMap(v interface{}) (map[string]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := make(map[string]string)

	switch val := v.(type) {
	case nil:
		// leave empty
	case map[string]interface{}:
		for k, raw := range val {
			if strVal, ok := raw.(string); ok {
				result[k] = strVal
			}
		}
	case map[string]string:
		result = val
	case string:
		result["en"] = val
	default:
		diags.AddError(
			"Unexpected Localized Field Type",
			fmt.Sprintf("Expected a string or map of strings, got: %T", v),
		)
	}

	return result, diags
}
