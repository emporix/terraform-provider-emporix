package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &TaxResource{}
var _ resource.ResourceWithImportState = &TaxResource{}

func NewTaxResource() resource.Resource {
	return &TaxResource{}
}

// TaxResource defines the resource implementation
type TaxResource struct {
	client *EmporixClient
}

// TaxResourceModel describes the resource data model
type TaxResourceModel struct {
	CountryCode types.String `tfsdk:"country_code"`
	TaxClasses  types.List   `tfsdk:"tax_classes"`
}

// TaxClassModel represents a single tax class in Terraform
type TaxClassModel struct {
	Code        types.String  `tfsdk:"code"`
	Name        types.Map     `tfsdk:"name"`
	Rate        types.Float64 `tfsdk:"rate"`
	Description types.Map     `tfsdk:"description"`
	Order       types.Int64   `tfsdk:"order"`
	IsDefault   types.Bool    `tfsdk:"is_default"`
}

func (r *TaxResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tax"
}

func (r *TaxResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages tax configurations for countries. Each country can have multiple tax classes with different rates. " +
			"Tax classes are sorted by their order value in ascending order. Only one tax class per country can be marked as default.",

		Attributes: map[string]schema.Attribute{
			"country_code": schema.StringAttribute{
				MarkdownDescription: "Country code (e.g., 'US', 'DE', 'GB'). Must follow Country Service standards (ISO 3166-1 alpha-2). Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tax_classes": schema.ListNestedAttribute{
				MarkdownDescription: "List of tax classes for this country. Tax classes are sorted by their order value. At least one tax class is required.",
				Required:            true,
				Validators: []validator.List{
					singleDefaultTaxClassValidator{},
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"code": schema.StringAttribute{
							MarkdownDescription: "Unique code for this tax class (e.g., 'STANDARD', 'REDUCED', 'ZERO').",
							Required:            true,
						},
						"name": schema.MapAttribute{
							MarkdownDescription: "Tax class name as a map of language codes to translated names. " +
								"Example: {en = \"Standard Rate\", de = \"Normalsteuersatz\"}. At least one language is required.",
							ElementType: types.StringType,
							Required:    true,
						},
						"rate": schema.Float64Attribute{
							MarkdownDescription: "Tax rate as a decimal (e.g., 0.19 for 19%, 0.07 for 7%).",
							Required:            true,
						},
						"description": schema.MapAttribute{
							MarkdownDescription: "Optional description as a map of language codes to translated descriptions. " +
								"Example: {en = \"Standard VAT rate\", de = \"Standard-Mehrwertsteuersatz\"}.",
							ElementType: types.StringType,
							Optional:    true,
						},
						"order": schema.Int64Attribute{
							MarkdownDescription: "Display order for this tax class. Tax classes are sorted by this value in ascending order. " +
								"Lower values appear first.",
							Optional: true,
						},
						"is_default": schema.BoolAttribute{
							MarkdownDescription: "Whether this is the default tax class for the country. Only one tax class can be default. " +
								"Defaults to false if not specified.",
							Optional: true,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (r *TaxResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TaxResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TaxResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating tax configuration", map[string]interface{}{
		"country_code": data.CountryCode.ValueString(),
	})

	// Parse tax_classes from plan
	var taxClassModels []TaxClassModel
	resp.Diagnostics.Append(data.TaxClasses.ElementsAs(ctx, &taxClassModels, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to API types
	taxClasses := make([]TaxClass, len(taxClassModels))
	for i, tcModel := range taxClassModels {
		// Convert name map
		nameMap := make(map[string]string)
		resp.Diagnostics.Append(tcModel.Name.ElementsAs(ctx, &nameMap, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		taxClass := TaxClass{
			Code:      tcModel.Code.ValueString(),
			Name:      nameMap,
			Rate:      tcModel.Rate.ValueFloat64(),
			IsDefault: tcModel.IsDefault.ValueBool(),
		}

		// Optional description
		if !tcModel.Description.IsNull() {
			descMap := make(map[string]string)
			resp.Diagnostics.Append(tcModel.Description.ElementsAs(ctx, &descMap, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			taxClass.Description = descMap
		}

		// Optional order
		if !tcModel.Order.IsNull() {
			order := int(tcModel.Order.ValueInt64())
			taxClass.Order = &order
		}

		taxClasses[i] = taxClass
	}

	// Create API request
	taxCreate := &TaxCreate{
		Location: &TaxLocation{
			CountryCode: data.CountryCode.ValueString(),
		},
		TaxClasses: taxClasses,
	}

	// Create via API
	tax, err := r.client.CreateTax(ctx, taxCreate)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create tax configuration, got error: %s", err))
		return
	}

	// Map API response to model
	mapTaxToModel(ctx, tax, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaxResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TaxResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading tax configuration", map[string]interface{}{
		"country_code": data.CountryCode.ValueString(),
	})

	// Get tax from API
	tax, err := r.client.GetTax(ctx, data.CountryCode.ValueString())
	if err != nil {
		// If resource not found, remove from state (drift detection)
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read tax configuration, got error: %s", err))
		return
	}

	// Map API response to model
	mapTaxToModel(ctx, tax, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaxResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TaxResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating tax configuration", map[string]interface{}{
		"country_code": data.CountryCode.ValueString(),
	})

	// Parse tax_classes from plan
	var taxClassModels []TaxClassModel
	resp.Diagnostics.Append(data.TaxClasses.ElementsAs(ctx, &taxClassModels, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to API types
	taxClasses := make([]TaxClass, len(taxClassModels))
	for i, tcModel := range taxClassModels {
		// Convert name map
		nameMap := make(map[string]string)
		resp.Diagnostics.Append(tcModel.Name.ElementsAs(ctx, &nameMap, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		taxClass := TaxClass{
			Code:      tcModel.Code.ValueString(),
			Name:      nameMap,
			Rate:      tcModel.Rate.ValueFloat64(),
			IsDefault: tcModel.IsDefault.ValueBool(),
		}

		// Optional description
		if !tcModel.Description.IsNull() {
			descMap := make(map[string]string)
			resp.Diagnostics.Append(tcModel.Description.ElementsAs(ctx, &descMap, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			taxClass.Description = descMap
		}

		// Optional order
		if !tcModel.Order.IsNull() {
			order := int(tcModel.Order.ValueInt64())
			taxClass.Order = &order
		}

		taxClasses[i] = taxClass
	}

	// Prepare update payload
	updateData := &TaxUpdate{
		Location: &TaxLocation{
			CountryCode: data.CountryCode.ValueString(),
		},
		TaxClasses: taxClasses,
	}

	// Update via API
	tax, err := r.client.UpdateTax(ctx, data.CountryCode.ValueString(), updateData)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update tax configuration, got error: %s", err))
		return
	}

	// Map updated response to model
	mapTaxToModel(ctx, tax, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaxResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TaxResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting tax configuration", map[string]interface{}{
		"country_code": data.CountryCode.ValueString(),
	})

	// Delete tax via API
	err := r.client.DeleteTax(ctx, data.CountryCode.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete tax configuration, got error: %s", err),
		)
		return
	}

	// Tax is now deleted and will be removed from Terraform state
}

func (r *TaxResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by country code (e.g., "US", "DE", "GB")
	resource.ImportStatePassthroughID(ctx, path.Root("country_code"), req, resp)
}

// mapTaxToModel converts a Tax API response to a TaxResourceModel
func mapTaxToModel(ctx context.Context, tax *Tax, data *TaxResourceModel, diags *diag.Diagnostics) {
	// Set country code
	data.CountryCode = types.StringValue(tax.Location.CountryCode)

	// Convert tax classes
	taxClassModels := make([]TaxClassModel, len(tax.TaxClasses))
	for i, tc := range tax.TaxClasses {
		model := TaxClassModel{
			Code:      types.StringValue(tc.Code),
			Rate:      types.Float64Value(tc.Rate),
			IsDefault: types.BoolValue(tc.IsDefault),
		}

		// Convert name (always map from API)
		if nameMap, ok := tc.Name.(map[string]interface{}); ok {
			nameStrMap := make(map[string]string)
			for k, v := range nameMap {
				if strVal, ok := v.(string); ok {
					nameStrMap[k] = strVal
				}
			}
			nameMapValue, d := types.MapValueFrom(ctx, types.StringType, nameStrMap)
			diags.Append(d...)
			model.Name = nameMapValue
		}

		// Convert description if present
		if tc.Description != nil {
			if descMap, ok := tc.Description.(map[string]interface{}); ok {
				descStrMap := make(map[string]string)
				for k, v := range descMap {
					if strVal, ok := v.(string); ok {
						descStrMap[k] = strVal
					}
				}
				descMapValue, d := types.MapValueFrom(ctx, types.StringType, descStrMap)
				diags.Append(d...)
				model.Description = descMapValue
			}
		} else {
			model.Description = types.MapNull(types.StringType)
		}

		// Set order if present
		if tc.Order != nil {
			model.Order = types.Int64Value(int64(*tc.Order))
		} else {
			model.Order = types.Int64Null()
		}

		taxClassModels[i] = model
	}

	// Convert to Terraform list
	taxClassesAttrType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"code":        types.StringType,
			"name":        types.MapType{ElemType: types.StringType},
			"rate":        types.Float64Type,
			"description": types.MapType{ElemType: types.StringType},
			"order":       types.Int64Type,
			"is_default":  types.BoolType,
		},
	}

	taxClassesList, d := types.ListValueFrom(ctx, taxClassesAttrType, taxClassModels)
	diags.Append(d...)
	data.TaxClasses = taxClassesList
}

// singleDefaultTaxClassValidator validates that at most one tax class has is_default = true
type singleDefaultTaxClassValidator struct{}

func (v singleDefaultTaxClassValidator) Description(ctx context.Context) string {
	return "Ensures at most one tax class is marked as default"
}

func (v singleDefaultTaxClassValidator) MarkdownDescription(ctx context.Context) string {
	return "Ensures at most one tax class is marked as default. The Emporix API constraint allows only one default tax class per country."
}

func (v singleDefaultTaxClassValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var taxClasses []TaxClassModel
	diags := req.ConfigValue.ElementsAs(ctx, &taxClasses, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Count defaults
	defaultCount := 0
	for _, tc := range taxClasses {
		if !tc.IsDefault.IsNull() && !tc.IsDefault.IsUnknown() && tc.IsDefault.ValueBool() {
			defaultCount++
		}
	}

	// Report error if more than one default
	if defaultCount > 1 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Multiple Default Tax Classes",
			fmt.Sprintf("Only one tax class can be marked as default, but %d tax classes have is_default = true. "+
				"Please set is_default = true for only one tax class.", defaultCount),
		)
	}
}
