package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SchemaResource{}
var _ resource.ResourceWithImportState = &SchemaResource{}

func NewSchemaResource() resource.Resource {
	return &SchemaResource{}
}

// SchemaResource defines the resource implementation.
type SchemaResource struct {
	client *EmporixClient
}

// SchemaResourceModel describes the resource data model.
type SchemaResourceModel struct {
	ID         types.String  `tfsdk:"id"`
	Name       types.Map     `tfsdk:"name"`
	Types      types.List    `tfsdk:"types"`
	Attributes types.Dynamic `tfsdk:"attributes"`
	SchemaUrl  types.String  `tfsdk:"schema_url"`
}

func (r *SchemaResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schema"
}

func (r *SchemaResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a schema in Emporix. " +
			"Schemas define the structure and validation rules for various entity types in the system. " +
			"The schema ID is immutable and cannot be changed after creation. " +
			"Supports unlimited nesting of OBJECT type attributes.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Schema identifier. Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.MapAttribute{
				MarkdownDescription: "Schema name as a map of language code to name (e.g., {\"en\": \"Product Schema\", \"de\": \"Produktschema\"}). Provide at least one language translation.",
				ElementType:         types.StringType,
				Required:            true,
			},
			"types": schema.ListAttribute{
				MarkdownDescription: "List of schema types. Valid values: CART, CATEGORY, COMPANY, COUPON, CUSTOMER, CUSTOMER_ADDRESS, ORDER, PRODUCT, QUOTE, RETURN, PRICE_LIST, SITE, CUSTOM_ENTITY, VENDOR.",
				ElementType:         types.StringType,
				Required:            true,
			},
			"schema_url": schema.StringAttribute{
				MarkdownDescription: "The URL of the schema, as returned by the API in the metadata.url field.",
				Computed:            true,
			},
			"attributes": schema.DynamicAttribute{
				MarkdownDescription: "List of schema attributes defining the structure. Supports unlimited nesting of OBJECT types. " +
					"Each attribute is an object with: key (string), name (map), type (string: TEXT, NUMBER, DECIMAL, BOOLEAN, DATE, TIME, DATE_TIME, ENUM, ARRAY, OBJECT, REFERENCE), " +
					"metadata (object with read_only, localized, required, nullable booleans), and optional: description (map), values (list for ENUM/REFERENCE), " +
					"attributes (list for OBJECT type - can be nested infinitely), array_type (object for ARRAY type with type, localized, values).",
				Required: true,
			},
		},
	}
}

func (r *SchemaResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SchemaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SchemaResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating schema", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Parse name map
	nameMap := make(map[string]string)
	resp.Diagnostics.Append(data.Name.ElementsAs(ctx, &nameMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse types list
	var typesList []string
	resp.Diagnostics.Append(data.Types.ElementsAs(ctx, &typesList, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse attributes from dynamic value
	attributes, diags := convertDynamicToAttributes(ctx, data.Attributes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create schema via API
	schemaCreate := &SchemaCreate{
		ID:         data.ID.ValueString(),
		Name:       nameMap,
		Types:      typesList,
		Attributes: attributes,
	}

	schema, err := r.client.CreateSchema(ctx, schemaCreate)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create schema, got error: %s", err))
		return
	}

	// Preserve the original attributes from the plan to maintain exact type structure
	// This is necessary because Terraform compares the returned value type with the planned value type
	originalAttributes := data.Attributes

	// Map API response to model (updates name, types, schema_url, etc.)
	mapSchemaToModel(ctx, schema, &data, &resp.Diagnostics)

	// Restore the original attributes to preserve the exact type structure from the plan
	data.Attributes = originalAttributes

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SchemaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SchemaResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading schema", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Get schema from API
	schema, err := r.client.GetSchema(ctx, data.ID.ValueString())
	if err != nil {
		// If resource not found, remove from state (drift detection)
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read schema, got error: %s", err))
		return
	}

	// Preserve the original attributes from state to maintain exact type structure
	// This is necessary because Terraform's dynamic type system requires exact type matching
	originalAttributes := data.Attributes

	// Map API response to model
	mapSchemaToModel(ctx, schema, &data, &resp.Diagnostics)

	// If we have original attributes (normal read), restore them to preserve type structure
	// If original is null/unknown (import case), keep the reconstructed attributes from API
	if !originalAttributes.IsNull() && !originalAttributes.IsUnknown() {
		data.Attributes = originalAttributes
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SchemaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SchemaResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating schema", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Parse name map
	nameMap := make(map[string]string)
	resp.Diagnostics.Append(data.Name.ElementsAs(ctx, &nameMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse types list
	var typesList []string
	resp.Diagnostics.Append(data.Types.ElementsAs(ctx, &typesList, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse attributes from dynamic value
	attributes, diags := convertDynamicToAttributes(ctx, data.Attributes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare update payload
	updateData := &SchemaUpdate{
		Name:       nameMap,
		Types:      typesList,
		Attributes: attributes,
	}

	// Update schema via API
	schema, err := r.client.UpdateSchema(ctx, data.ID.ValueString(), updateData)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update schema, got error: %s", err))
		return
	}

	// Preserve the original attributes from the plan to maintain exact type structure
	originalAttributes := data.Attributes

	// Map updated response to model
	mapSchemaToModel(ctx, schema, &data, &resp.Diagnostics)

	// Restore the original attributes to preserve the exact type structure from the plan
	data.Attributes = originalAttributes

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SchemaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SchemaResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting schema", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Delete schema via API
	err := r.client.DeleteSchema(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete schema, got error: %s", err),
		)
		return
	}

	// Schema is now deleted and will be removed from Terraform state
}

func (r *SchemaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by schema ID
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapSchemaToModel converts a Schema API response to a SchemaResourceModel
func mapSchemaToModel(ctx context.Context, schema *Schema, data *SchemaResourceModel, diags *diag.Diagnostics) {
	data.ID = types.StringValue(schema.ID)

	// Convert name map from API to Terraform map
	if schema.Name != nil && len(schema.Name) > 0 {
		nameMapValue, d := types.MapValueFrom(ctx, types.StringType, schema.Name)
		diags.Append(d...)
		data.Name = nameMapValue
	} else {
		nameMapValue, d := types.MapValueFrom(ctx, types.StringType, map[string]string{})
		diags.Append(d...)
		data.Name = nameMapValue
	}

	// Convert types list
	if len(schema.Types) > 0 {
		typesList, d := types.ListValueFrom(ctx, types.StringType, schema.Types)
		diags.Append(d...)
		data.Types = typesList
	} else {
		typesList, d := types.ListValueFrom(ctx, types.StringType, []string{})
		diags.Append(d...)
		data.Types = typesList
	}

	// Convert attributes to dynamic value
	attrsDynamic, d := convertAttributesToDynamic(ctx, schema.Attributes)
	diags.Append(d...)
	data.Attributes = attrsDynamic

	// Set schema_url from metadata.url
	if schema.Metadata != nil && schema.Metadata.URL != "" {
		data.SchemaUrl = types.StringValue(schema.Metadata.URL)
	} else {
		data.SchemaUrl = types.StringNull()
	}
}

// convertAttributesToDynamic converts API SchemaAttributes to a dynamic Terraform value
// Returns a tuple value since HCL [...] in dynamic contexts is interpreted as a tuple
func convertAttributesToDynamic(ctx context.Context, attributes []SchemaAttribute) (types.Dynamic, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(attributes) == 0 {
		// Return empty tuple for empty attributes
		tupleValue, d := types.TupleValue([]attr.Type{}, []attr.Value{})
		diags.Append(d...)
		return types.DynamicValue(tupleValue), diags
	}

	attrValues := make([]attr.Value, len(attributes))
	attrTypes := make([]attr.Type, len(attributes))
	for i, schemaAttr := range attributes {
		attrObj, d := convertAttributeToObject(ctx, schemaAttr)
		diags.Append(d...)
		attrValues[i] = attrObj
		attrTypes[i] = attrObj.Type(ctx)
	}

	tupleValue, d := types.TupleValue(attrTypes, attrValues)
	diags.Append(d...)

	return types.DynamicValue(tupleValue), diags
}

// convertAttributeToObject converts a single SchemaAttribute to a Terraform object value
func convertAttributeToObject(ctx context.Context, schemaAttr SchemaAttribute) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Convert name map
	nameMap, d := types.MapValueFrom(ctx, types.StringType, schemaAttr.Name)
	diags.Append(d...)

	// Convert description map
	var descMap types.Map
	if len(schemaAttr.Description) > 0 {
		descMap, d = types.MapValueFrom(ctx, types.StringType, schemaAttr.Description)
		diags.Append(d...)
	} else {
		descMap = types.MapNull(types.StringType)
	}

	// Convert metadata
	metadataAttrTypes := map[string]attr.Type{
		"read_only":  types.BoolType,
		"localized":  types.BoolType,
		"required":   types.BoolType,
		"nullable":   types.BoolType,
	}
	var metadataObj types.Object
	if schemaAttr.Metadata != nil {
		metadataAttrValues := map[string]attr.Value{
			"read_only":  types.BoolValue(schemaAttr.Metadata.ReadOnly),
			"localized":  types.BoolValue(schemaAttr.Metadata.Localized),
			"required":   types.BoolValue(schemaAttr.Metadata.Required),
			"nullable":   types.BoolValue(schemaAttr.Metadata.Nullable),
		}
		metadataObj, d = types.ObjectValue(metadataAttrTypes, metadataAttrValues)
		diags.Append(d...)
	} else {
		metadataObj = types.ObjectNull(metadataAttrTypes)
	}

	// Convert values
	valueAttrTypes := map[string]attr.Type{"value": types.StringType}
	var valuesList types.List
	if len(schemaAttr.Values) > 0 {
		valueObjs := make([]attr.Value, len(schemaAttr.Values))
		for i, val := range schemaAttr.Values {
			valueAttrValues := map[string]attr.Value{"value": types.StringValue(val.Value)}
			valueObj, d := types.ObjectValue(valueAttrTypes, valueAttrValues)
			diags.Append(d...)
			valueObjs[i] = valueObj
		}
		valuesList, d = types.ListValue(types.ObjectType{AttrTypes: valueAttrTypes}, valueObjs)
		diags.Append(d...)
	} else {
		valuesList = types.ListNull(types.ObjectType{AttrTypes: valueAttrTypes})
	}

	// Convert nested attributes (recursive!)
	// Return tuple since HCL [...] in dynamic contexts is interpreted as a tuple
	var nestedAttrsTuple types.Tuple
	if len(schemaAttr.Attributes) > 0 {
		nestedAttrValues := make([]attr.Value, len(schemaAttr.Attributes))
		nestedAttrTypes := make([]attr.Type, len(schemaAttr.Attributes))
		for i, nestedAttr := range schemaAttr.Attributes {
			nestedObj, d := convertAttributeToObject(ctx, nestedAttr)
			diags.Append(d...)
			nestedAttrValues[i] = nestedObj
			nestedAttrTypes[i] = nestedObj.Type(ctx)
		}
		nestedAttrsTuple, d = types.TupleValue(nestedAttrTypes, nestedAttrValues)
		diags.Append(d...)
	} else {
		nestedAttrsTuple, d = types.TupleValue([]attr.Type{}, []attr.Value{})
		diags.Append(d...)
	}

	// Convert array_type
	arrayTypeAttrTypes := map[string]attr.Type{
		"type":      types.StringType,
		"localized": types.BoolType,
		"values":    types.ListType{ElemType: types.ObjectType{AttrTypes: valueAttrTypes}},
	}
	var arrayTypeObj types.Object
	if schemaAttr.ArrayType != nil {
		var arrayValuesList types.List
		if len(schemaAttr.ArrayType.Values) > 0 {
			arrayValueObjs := make([]attr.Value, len(schemaAttr.ArrayType.Values))
			for i, val := range schemaAttr.ArrayType.Values {
				valAttrValues := map[string]attr.Value{"value": types.StringValue(val.Value)}
				valObj, d := types.ObjectValue(valueAttrTypes, valAttrValues)
				diags.Append(d...)
				arrayValueObjs[i] = valObj
			}
			arrayValuesList, d = types.ListValue(types.ObjectType{AttrTypes: valueAttrTypes}, arrayValueObjs)
			diags.Append(d...)
		} else {
			arrayValuesList = types.ListNull(types.ObjectType{AttrTypes: valueAttrTypes})
		}

		arrayTypeAttrValues := map[string]attr.Value{
			"type":      types.StringValue(schemaAttr.ArrayType.Type),
			"localized": types.BoolValue(schemaAttr.ArrayType.Localized),
			"values":    arrayValuesList,
		}
		arrayTypeObj, d = types.ObjectValue(arrayTypeAttrTypes, arrayTypeAttrValues)
		diags.Append(d...)
	} else {
		arrayTypeObj = types.ObjectNull(arrayTypeAttrTypes)
	}

	// Build the attribute object
	// Wrap nested attributes tuple in DynamicValue since the type is DynamicType
	attrObjValues := map[string]attr.Value{
		"key":         types.StringValue(schemaAttr.Key),
		"name":        nameMap,
		"description": descMap,
		"type":        types.StringValue(schemaAttr.Type),
		"metadata":    metadataObj,
		"values":      valuesList,
		"attributes":  types.DynamicValue(nestedAttrsTuple),
		"array_type":  arrayTypeObj,
	}
	obj, d := types.ObjectValue(getAttributeObjectType(), attrObjValues)
	diags.Append(d...)

	return obj, diags
}

// convertDynamicToAttributes converts a dynamic Terraform value to API SchemaAttributes
func convertDynamicToAttributes(ctx context.Context, dynamic types.Dynamic) ([]SchemaAttribute, diag.Diagnostics) {
	var diags diag.Diagnostics

	if dynamic.IsNull() || dynamic.IsUnknown() {
		return nil, diags
	}

	// The dynamic value should be a list
	underlyingValue := dynamic.UnderlyingValue()
	if underlyingValue == nil {
		return nil, diags
	}

	// Try to convert to a list or tuple
	var attrList []attr.Value

	switch v := underlyingValue.(type) {
	case basetypes.ListValue:
		attrList = v.Elements()
	case basetypes.TupleValue:
		attrList = v.Elements()
	default:
		// Try JSON parsing as fallback
		jsonBytes, err := json.Marshal(underlyingValue)
		if err != nil {
			diags.AddError("Conversion Error", fmt.Sprintf("Unable to convert attributes: %s", err))
			return nil, diags
		}
		var attributes []SchemaAttribute
		if err := json.Unmarshal(jsonBytes, &attributes); err != nil {
			diags.AddError("Conversion Error", fmt.Sprintf("Unable to parse attributes: %s", err))
			return nil, diags
		}
		return attributes, diags
	}

	attributes := make([]SchemaAttribute, len(attrList))
	for i, attrVal := range attrList {
		attr, d := convertObjectToAttribute(ctx, attrVal)
		diags.Append(d...)
		attributes[i] = attr
	}

	return attributes, diags
}

// convertObjectToAttribute converts a Terraform object value to a SchemaAttribute
func convertObjectToAttribute(ctx context.Context, val attr.Value) (SchemaAttribute, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result SchemaAttribute

	// Handle different value types
	var objMap map[string]attr.Value

	switch v := val.(type) {
	case basetypes.ObjectValue:
		objMap = v.Attributes()
	case basetypes.DynamicValue:
		if underlying, ok := v.UnderlyingValue().(basetypes.ObjectValue); ok {
			objMap = underlying.Attributes()
		} else {
			diags.AddError("Conversion Error", "Expected object value in attributes list")
			return result, diags
		}
	default:
		diags.AddError("Conversion Error", fmt.Sprintf("Unexpected value type in attributes list: %T", val))
		return result, diags
	}

	// Extract key
	if keyVal, ok := objMap["key"]; ok && !keyVal.IsNull() {
		if strVal, ok := keyVal.(basetypes.StringValue); ok {
			result.Key = strVal.ValueString()
		}
	}

	// Extract name (can be MapValue or ObjectValue depending on context)
	if nameVal, ok := objMap["name"]; ok && !nameVal.IsNull() {
		result.Name = extractStringMap(nameVal)
	}

	// Extract description (can be MapValue or ObjectValue depending on context)
	if descVal, ok := objMap["description"]; ok && !descVal.IsNull() {
		result.Description = extractStringMap(descVal)
	}

	// Extract type
	if typeVal, ok := objMap["type"]; ok && !typeVal.IsNull() {
		if strVal, ok := typeVal.(basetypes.StringValue); ok {
			result.Type = strVal.ValueString()
		}
	}

	// Extract metadata
	if metaVal, ok := objMap["metadata"]; ok && !metaVal.IsNull() {
		if objVal, ok := metaVal.(basetypes.ObjectValue); ok {
			metaAttrs := objVal.Attributes()
			result.Metadata = &SchemaAttributeMetadata{}
			if v, ok := metaAttrs["read_only"].(basetypes.BoolValue); ok {
				result.Metadata.ReadOnly = v.ValueBool()
			}
			if v, ok := metaAttrs["localized"].(basetypes.BoolValue); ok {
				result.Metadata.Localized = v.ValueBool()
			}
			if v, ok := metaAttrs["required"].(basetypes.BoolValue); ok {
				result.Metadata.Required = v.ValueBool()
			}
			if v, ok := metaAttrs["nullable"].(basetypes.BoolValue); ok {
				result.Metadata.Nullable = v.ValueBool()
			}
		}
	}

	// Extract values
	if valuesVal, ok := objMap["values"]; ok && !valuesVal.IsNull() {
		var valuesList []attr.Value
		switch v := valuesVal.(type) {
		case basetypes.ListValue:
			valuesList = v.Elements()
		case basetypes.TupleValue:
			valuesList = v.Elements()
		}
		if len(valuesList) > 0 {
			result.Values = make([]SchemaAttributeValue, len(valuesList))
			for i, valItem := range valuesList {
				if objVal, ok := valItem.(basetypes.ObjectValue); ok {
					if v, ok := objVal.Attributes()["value"].(basetypes.StringValue); ok {
						result.Values[i] = SchemaAttributeValue{Value: v.ValueString()}
					}
				}
			}
		}
	}

	// Extract nested attributes (recursive!)
	if attrsVal, ok := objMap["attributes"]; ok && !attrsVal.IsNull() {
		var nestedList []attr.Value
		switch v := attrsVal.(type) {
		case basetypes.ListValue:
			nestedList = v.Elements()
		case basetypes.TupleValue:
			nestedList = v.Elements()
		}
		if len(nestedList) > 0 {
			result.Attributes = make([]SchemaAttribute, len(nestedList))
			for i, nestedVal := range nestedList {
				nestedAttr, d := convertObjectToAttribute(ctx, nestedVal)
				diags.Append(d...)
				result.Attributes[i] = nestedAttr
			}
		}
	}

	// Extract array_type
	if arrayTypeVal, ok := objMap["array_type"]; ok && !arrayTypeVal.IsNull() {
		if objVal, ok := arrayTypeVal.(basetypes.ObjectValue); ok {
			arrayAttrs := objVal.Attributes()
			result.ArrayType = &SchemaArrayType{}
			if v, ok := arrayAttrs["type"].(basetypes.StringValue); ok {
				result.ArrayType.Type = v.ValueString()
			}
			if v, ok := arrayAttrs["localized"].(basetypes.BoolValue); ok {
				result.ArrayType.Localized = v.ValueBool()
			}
			if valuesVal, ok := arrayAttrs["values"]; ok && !valuesVal.IsNull() {
				var valuesList []attr.Value
				switch v := valuesVal.(type) {
				case basetypes.ListValue:
					valuesList = v.Elements()
				case basetypes.TupleValue:
					valuesList = v.Elements()
				}
				if len(valuesList) > 0 {
					result.ArrayType.Values = make([]SchemaAttributeValue, len(valuesList))
					for i, valItem := range valuesList {
						if objVal, ok := valItem.(basetypes.ObjectValue); ok {
							if v, ok := objVal.Attributes()["value"].(basetypes.StringValue); ok {
								result.ArrayType.Values[i] = SchemaAttributeValue{Value: v.ValueString()}
							}
						}
					}
				}
			}
		}
	}

	return result, diags
}

// getAttributeObjectType returns the attr.Type map for a schema attribute object
// This is used for constructing typed values from the recursive structure
func getAttributeObjectType() map[string]attr.Type {
	// Note: We use a simplified type here because the recursive nature
	// means we can't fully define the nested type statically.
	// The dynamic attribute allows this to work at runtime.
	return map[string]attr.Type{
		"key":         types.StringType,
		"name":        types.MapType{ElemType: types.StringType},
		"description": types.MapType{ElemType: types.StringType},
		"type":        types.StringType,
		"metadata": types.ObjectType{AttrTypes: map[string]attr.Type{
			"read_only":  types.BoolType,
			"localized":  types.BoolType,
			"required":   types.BoolType,
			"nullable":   types.BoolType,
		}},
		"values": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
			"value": types.StringType,
		}}},
		"attributes": types.DynamicType, // Dynamic to allow tuple/list for recursive nesting
		"array_type": types.ObjectType{AttrTypes: map[string]attr.Type{
			"type":      types.StringType,
			"localized": types.BoolType,
			"values": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
				"value": types.StringType,
			}}},
		}},
	}
}

// extractStringMap extracts a map[string]string from an attr.Value
// Handles both MapValue and ObjectValue since Terraform may represent
// { en = "value" } as either type depending on the schema context
func extractStringMap(val attr.Value) map[string]string {
	result := make(map[string]string)

	switch v := val.(type) {
	case basetypes.MapValue:
		for k, elem := range v.Elements() {
			if strVal, ok := elem.(basetypes.StringValue); ok {
				result[k] = strVal.ValueString()
			}
		}
	case basetypes.ObjectValue:
		for k, elem := range v.Attributes() {
			if strVal, ok := elem.(basetypes.StringValue); ok {
				result[k] = strVal.ValueString()
			}
		}
	}

	return result
}
