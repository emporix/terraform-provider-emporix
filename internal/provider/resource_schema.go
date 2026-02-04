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
	ID         types.String `tfsdk:"id"`
	Name       types.Map    `tfsdk:"name"`
	Types      types.List   `tfsdk:"types"`
	Attributes types.List   `tfsdk:"attributes"`
	SchemaUrl  types.String `tfsdk:"schema_url"`
}

// SchemaAttributeModel describes the attribute data model
type SchemaAttributeModel struct {
	Key         types.String `tfsdk:"key"`
	Name        types.Map    `tfsdk:"name"`
	Description types.Map    `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Metadata    types.Object `tfsdk:"metadata"`
	Values      types.List   `tfsdk:"values"`
	Attributes  types.List   `tfsdk:"attributes"`
	ArrayType   types.Object `tfsdk:"array_type"`
}

// SchemaAttributeMetadataModel describes the attribute metadata model
type SchemaAttributeMetadataModel struct {
	ReadOnly  types.Bool `tfsdk:"read_only"`
	Localized types.Bool `tfsdk:"localized"`
	Required  types.Bool `tfsdk:"required"`
	Nullable  types.Bool `tfsdk:"nullable"`
}

// SchemaAttributeValueModel describes the attribute value model
type SchemaAttributeValueModel struct {
	Value types.String `tfsdk:"value"`
}

// NestedSchemaAttributeModel describes nested attribute model (simplified, no further nesting)
type NestedSchemaAttributeModel struct {
	Key         types.String `tfsdk:"key"`
	Name        types.Map    `tfsdk:"name"`
	Description types.Map    `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Metadata    types.Object `tfsdk:"metadata"`
}

// SchemaArrayTypeModel describes the array type model
type SchemaArrayTypeModel struct {
	Type      types.String `tfsdk:"type"`
	Localized types.Bool   `tfsdk:"localized"`
	Values    types.List   `tfsdk:"values"`
}

func (r *SchemaResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schema"
}

func (r *SchemaResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a schema in Emporix. " +
			"Schemas define the structure and validation rules for various entity types in the system. " +
			"The schema ID is immutable and cannot be changed after creation.",

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
			"attributes": schema.ListNestedAttribute{
				MarkdownDescription: "List of schema attributes defining the structure.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							MarkdownDescription: "Unique attribute identifier.",
							Required:            true,
						},
						"name": schema.MapAttribute{
							MarkdownDescription: "Attribute name as a map of language code to name.",
							ElementType:         types.StringType,
							Required:            true,
						},
						"description": schema.MapAttribute{
							MarkdownDescription: "Attribute description as a map of language code to description.",
							ElementType:         types.StringType,
							Optional:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Attribute type. Valid values: TEXT, NUMBER, DECIMAL, BOOLEAN, DATE, TIME, DATE_TIME, ENUM, ARRAY, OBJECT, REFERENCE.",
							Required:            true,
						},
						"metadata": schema.SingleNestedAttribute{
							MarkdownDescription: "Attribute metadata.",
							Required:            true,
							Attributes: map[string]schema.Attribute{
								"read_only": schema.BoolAttribute{
									MarkdownDescription: "Whether the attribute is read-only.",
									Required:            true,
								},
								"localized": schema.BoolAttribute{
									MarkdownDescription: "Whether the attribute is localized.",
									Required:            true,
								},
								"required": schema.BoolAttribute{
									MarkdownDescription: "Whether the attribute is required.",
									Required:            true,
								},
								"nullable": schema.BoolAttribute{
									MarkdownDescription: "Whether the attribute can be null.",
									Required:            true,
								},
							},
						},
						"values": schema.ListNestedAttribute{
							MarkdownDescription: "List of allowed values for ENUM or REFERENCE types.",
							Optional:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"value": schema.StringAttribute{
										MarkdownDescription: "Allowed value for ENUM or REFERENCE type.",
										Required:            true,
									},
								},
							},
						},
						"attributes": schema.ListNestedAttribute{
							MarkdownDescription: "Nested attributes for OBJECT type.",
							Optional:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"key": schema.StringAttribute{
										MarkdownDescription: "Unique attribute identifier.",
										Required:            true,
									},
									"name": schema.MapAttribute{
										MarkdownDescription: "Attribute name as a map of language code to name.",
										ElementType:         types.StringType,
										Required:            true,
									},
									"description": schema.MapAttribute{
										MarkdownDescription: "Attribute description as a map of language code to description.",
										ElementType:         types.StringType,
										Optional:            true,
									},
									"type": schema.StringAttribute{
										MarkdownDescription: "Attribute type.",
										Required:            true,
									},
									"metadata": schema.SingleNestedAttribute{
										MarkdownDescription: "Attribute metadata.",
										Required:            true,
										Attributes: map[string]schema.Attribute{
											"read_only": schema.BoolAttribute{
												MarkdownDescription: "Whether the attribute is read-only.",
												Required:            true,
											},
											"localized": schema.BoolAttribute{
												MarkdownDescription: "Whether the attribute is localized.",
												Required:            true,
											},
											"required": schema.BoolAttribute{
												MarkdownDescription: "Whether the attribute is required.",
												Required:            true,
											},
											"nullable": schema.BoolAttribute{
												MarkdownDescription: "Whether the attribute can be null.",
												Required:            true,
											},
										},
									},
								},
							},
						},
						"array_type": schema.SingleNestedAttribute{
							MarkdownDescription: "Array type configuration for ARRAY attributes.",
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									MarkdownDescription: "Element type for the array.",
									Required:            true,
								},
								"localized": schema.BoolAttribute{
									MarkdownDescription: "Whether array elements are localized.",
									Optional:            true,
								},
								"values": schema.ListNestedAttribute{
									MarkdownDescription: "List of allowed values for ENUM array elements.",
									Optional:            true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"value": schema.StringAttribute{
												MarkdownDescription: "Allowed value for ENUM array element.",
												Required:            true,
											},
										},
									},
								},
							},
						},
					},
				},
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

	// Parse attributes
	attributes, diags := convertAttributesFromModel(ctx, data.Attributes)
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

	// Map API response to model
	mapSchemaToModel(ctx, schema, &data, &resp.Diagnostics)

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

	// Map API response to model
	mapSchemaToModel(ctx, schema, &data, &resp.Diagnostics)

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

	// Parse attributes
	attributes, diags := convertAttributesFromModel(ctx, data.Attributes)
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

	// Map updated response to model
	mapSchemaToModel(ctx, schema, &data, &resp.Diagnostics)

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

	// Convert attributes
	attrList, d := convertAttributesToModel(ctx, schema.Attributes)
	diags.Append(d...)
	data.Attributes = attrList

	// Set schema_url from metadata.url
	if schema.Metadata != nil && schema.Metadata.URL != "" {
		data.SchemaUrl = types.StringValue(schema.Metadata.URL)
	} else {
		data.SchemaUrl = types.StringNull()
	}
}

// convertAttributesToModel converts API SchemaAttributes to Terraform model
func convertAttributesToModel(ctx context.Context, attributes []SchemaAttribute) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(attributes) == 0 {
		emptyList, d := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: getSchemaAttributeAttrTypes(),
		}, []SchemaAttributeModel{})
		diags.Append(d...)
		return emptyList, diags
	}

	attrModels := make([]SchemaAttributeModel, len(attributes))
	for i, attr := range attributes {
		// Convert name
		nameMap, d := types.MapValueFrom(ctx, types.StringType, attr.Name)
		diags.Append(d...)

		// Convert description
		var descMap types.Map
		if len(attr.Description) > 0 {
			descMap, d = types.MapValueFrom(ctx, types.StringType, attr.Description)
			diags.Append(d...)
		} else {
			descMap = types.MapNull(types.StringType)
		}

		// Convert metadata
		var metadataObj types.Object
		if attr.Metadata != nil {
			metadataModel := SchemaAttributeMetadataModel{
				ReadOnly:  types.BoolValue(attr.Metadata.ReadOnly),
				Localized: types.BoolValue(attr.Metadata.Localized),
				Required:  types.BoolValue(attr.Metadata.Required),
				Nullable:  types.BoolValue(attr.Metadata.Nullable),
			}
			metadataObj, d = types.ObjectValueFrom(ctx, getSchemaAttributeMetadataAttrTypes(), metadataModel)
			diags.Append(d...)
		} else {
			metadataObj = types.ObjectNull(getSchemaAttributeMetadataAttrTypes())
		}

		// Convert values
		var valuesList types.List
		if len(attr.Values) > 0 {
			valueModels := make([]SchemaAttributeValueModel, len(attr.Values))
			for j, val := range attr.Values {
				valueModels[j] = SchemaAttributeValueModel{
					Value: types.StringValue(val.Value),
				}
			}
			valuesList, d = types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: getSchemaAttributeValueAttrTypes(),
			}, valueModels)
			diags.Append(d...)
		} else {
			valuesList = types.ListNull(types.ObjectType{
				AttrTypes: getSchemaAttributeValueAttrTypes(),
			})
		}

		// Convert nested attributes (for OBJECT type)
		var nestedAttrsList types.List
		if len(attr.Attributes) > 0 {
			nestedModels := make([]NestedSchemaAttributeModel, len(attr.Attributes))
			for j, nestedAttr := range attr.Attributes {
				nestedNameMap, d := types.MapValueFrom(ctx, types.StringType, nestedAttr.Name)
				diags.Append(d...)

				var nestedDescMap types.Map
				if len(nestedAttr.Description) > 0 {
					nestedDescMap, d = types.MapValueFrom(ctx, types.StringType, nestedAttr.Description)
					diags.Append(d...)
				} else {
					nestedDescMap = types.MapNull(types.StringType)
				}

				var nestedMetadataObj types.Object
				if nestedAttr.Metadata != nil {
					nestedMetadataModel := SchemaAttributeMetadataModel{
						ReadOnly:  types.BoolValue(nestedAttr.Metadata.ReadOnly),
						Localized: types.BoolValue(nestedAttr.Metadata.Localized),
						Required:  types.BoolValue(nestedAttr.Metadata.Required),
						Nullable:  types.BoolValue(nestedAttr.Metadata.Nullable),
					}
					nestedMetadataObj, d = types.ObjectValueFrom(ctx, getSchemaAttributeMetadataAttrTypes(), nestedMetadataModel)
					diags.Append(d...)
				} else {
					nestedMetadataObj = types.ObjectNull(getSchemaAttributeMetadataAttrTypes())
				}

				nestedModels[j] = NestedSchemaAttributeModel{
					Key:         types.StringValue(nestedAttr.Key),
					Name:        nestedNameMap,
					Description: nestedDescMap,
					Type:        types.StringValue(nestedAttr.Type),
					Metadata:    nestedMetadataObj,
				}
			}
			nestedAttrsList, d = types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: getNestedSchemaAttributeAttrTypes(),
			}, nestedModels)
			diags.Append(d...)
		} else {
			nestedAttrsList = types.ListNull(types.ObjectType{
				AttrTypes: getNestedSchemaAttributeAttrTypes(),
			})
		}

		// Convert array_type
		var arrayTypeObj types.Object
		if attr.ArrayType != nil {
			var arrayValuesList types.List
			if len(attr.ArrayType.Values) > 0 {
				arrayValueModels := make([]SchemaAttributeValueModel, len(attr.ArrayType.Values))
				for j, val := range attr.ArrayType.Values {
					arrayValueModels[j] = SchemaAttributeValueModel{
						Value: types.StringValue(val.Value),
					}
				}
				arrayValuesList, d = types.ListValueFrom(ctx, types.ObjectType{
					AttrTypes: getSchemaAttributeValueAttrTypes(),
				}, arrayValueModels)
				diags.Append(d...)
			} else {
				arrayValuesList = types.ListNull(types.ObjectType{
					AttrTypes: getSchemaAttributeValueAttrTypes(),
				})
			}

			arrayTypeModel := SchemaArrayTypeModel{
				Type:      types.StringValue(attr.ArrayType.Type),
				Localized: types.BoolValue(attr.ArrayType.Localized),
				Values:    arrayValuesList,
			}
			arrayTypeObj, d = types.ObjectValueFrom(ctx, getSchemaArrayTypeAttrTypes(), arrayTypeModel)
			diags.Append(d...)
		} else {
			arrayTypeObj = types.ObjectNull(getSchemaArrayTypeAttrTypes())
		}

		attrModels[i] = SchemaAttributeModel{
			Key:         types.StringValue(attr.Key),
			Name:        nameMap,
			Description: descMap,
			Type:        types.StringValue(attr.Type),
			Metadata:    metadataObj,
			Values:      valuesList,
			Attributes:  nestedAttrsList,
			ArrayType:   arrayTypeObj,
		}
	}

	attrList, d := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: getSchemaAttributeAttrTypes(),
	}, attrModels)
	diags.Append(d...)

	return attrList, diags
}

// convertAttributesFromModel converts Terraform model to API SchemaAttributes
func convertAttributesFromModel(ctx context.Context, attributesList types.List) ([]SchemaAttribute, diag.Diagnostics) {
	var diags diag.Diagnostics

	var attrModels []SchemaAttributeModel
	diags.Append(attributesList.ElementsAs(ctx, &attrModels, false)...)
	if diags.HasError() {
		return nil, diags
	}

	attributes := make([]SchemaAttribute, len(attrModels))
	for i, attrModel := range attrModels {
		// Convert name
		var nameMap map[string]string
		diags.Append(attrModel.Name.ElementsAs(ctx, &nameMap, false)...)

		// Convert description
		var descMap map[string]string
		if !attrModel.Description.IsNull() {
			diags.Append(attrModel.Description.ElementsAs(ctx, &descMap, false)...)
		}

		// Convert metadata
		var metadataModel SchemaAttributeMetadataModel
		diags.Append(attrModel.Metadata.As(ctx, &metadataModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
		metadata := &SchemaAttributeMetadata{
			ReadOnly:  metadataModel.ReadOnly.ValueBool(),
			Localized: metadataModel.Localized.ValueBool(),
			Required:  metadataModel.Required.ValueBool(),
			Nullable:  metadataModel.Nullable.ValueBool(),
		}

		// Convert values
		var values []SchemaAttributeValue
		if !attrModel.Values.IsNull() {
			var valueModels []SchemaAttributeValueModel
			diags.Append(attrModel.Values.ElementsAs(ctx, &valueModels, false)...)
			values = make([]SchemaAttributeValue, len(valueModels))
			for j, valModel := range valueModels {
				values[j] = SchemaAttributeValue{
					Value: valModel.Value.ValueString(),
				}
			}
		}

		// Convert nested attributes
		var nestedAttrs []SchemaAttribute
		if !attrModel.Attributes.IsNull() {
			var nestedModels []NestedSchemaAttributeModel
			diags.Append(attrModel.Attributes.ElementsAs(ctx, &nestedModels, false)...)
			nestedAttrs = make([]SchemaAttribute, len(nestedModels))
			for j, nestedModel := range nestedModels {
				var nestedNameMap map[string]string
				diags.Append(nestedModel.Name.ElementsAs(ctx, &nestedNameMap, false)...)

				var nestedDescMap map[string]string
				if !nestedModel.Description.IsNull() {
					diags.Append(nestedModel.Description.ElementsAs(ctx, &nestedDescMap, false)...)
				}

				var nestedMetadataModel SchemaAttributeMetadataModel
				diags.Append(nestedModel.Metadata.As(ctx, &nestedMetadataModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
				nestedMetadata := &SchemaAttributeMetadata{
					ReadOnly:  nestedMetadataModel.ReadOnly.ValueBool(),
					Localized: nestedMetadataModel.Localized.ValueBool(),
					Required:  nestedMetadataModel.Required.ValueBool(),
					Nullable:  nestedMetadataModel.Nullable.ValueBool(),
				}

				nestedAttrs[j] = SchemaAttribute{
					Key:         nestedModel.Key.ValueString(),
					Name:        nestedNameMap,
					Description: nestedDescMap,
					Type:        nestedModel.Type.ValueString(),
					Metadata:    nestedMetadata,
				}
			}
		}

		// Convert array_type
		var arrayType *SchemaArrayType
		if !attrModel.ArrayType.IsNull() {
			var arrayTypeModel SchemaArrayTypeModel
			diags.Append(attrModel.ArrayType.As(ctx, &arrayTypeModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)

			arrayType = &SchemaArrayType{
				Type:      arrayTypeModel.Type.ValueString(),
				Localized: arrayTypeModel.Localized.ValueBool(),
			}

			if !arrayTypeModel.Values.IsNull() {
				var arrayValueModels []SchemaAttributeValueModel
				diags.Append(arrayTypeModel.Values.ElementsAs(ctx, &arrayValueModels, false)...)
				arrayType.Values = make([]SchemaAttributeValue, len(arrayValueModels))
				for j, valModel := range arrayValueModels {
					arrayType.Values[j] = SchemaAttributeValue{
						Value: valModel.Value.ValueString(),
					}
				}
			}
		}

		attributes[i] = SchemaAttribute{
			Key:         attrModel.Key.ValueString(),
			Name:        nameMap,
			Description: descMap,
			Type:        attrModel.Type.ValueString(),
			Metadata:    metadata,
			Values:      values,
			Attributes:  nestedAttrs,
			ArrayType:   arrayType,
		}
	}

	return attributes, diags
}

// Helper functions to define attribute types for nested objects
func getSchemaAttributeAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"key":         types.StringType,
		"name":        types.MapType{ElemType: types.StringType},
		"description": types.MapType{ElemType: types.StringType},
		"type":        types.StringType,
		"metadata":    types.ObjectType{AttrTypes: getSchemaAttributeMetadataAttrTypes()},
		"values":      types.ListType{ElemType: types.ObjectType{AttrTypes: getSchemaAttributeValueAttrTypes()}},
		"attributes":  types.ListType{ElemType: types.ObjectType{AttrTypes: getNestedSchemaAttributeAttrTypes()}},
		"array_type":  types.ObjectType{AttrTypes: getSchemaArrayTypeAttrTypes()},
	}
}

func getNestedSchemaAttributeAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"key":         types.StringType,
		"name":        types.MapType{ElemType: types.StringType},
		"description": types.MapType{ElemType: types.StringType},
		"type":        types.StringType,
		"metadata":    types.ObjectType{AttrTypes: getSchemaAttributeMetadataAttrTypes()},
	}
}

func getSchemaAttributeMetadataAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"read_only":  types.BoolType,
		"localized":  types.BoolType,
		"required":   types.BoolType,
		"nullable":   types.BoolType,
	}
}

func getSchemaAttributeValueAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"value": types.StringType,
	}
}

func getSchemaArrayTypeAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":      types.StringType,
		"localized": types.BoolType,
		"values":    types.ListType{ElemType: types.ObjectType{AttrTypes: getSchemaAttributeValueAttrTypes()}},
	}
}
