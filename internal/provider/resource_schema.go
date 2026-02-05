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

// NestedSchemaAttributeModel describes nested attribute model (supports one more level of nesting)
type NestedSchemaAttributeModel struct {
	Key         types.String `tfsdk:"key"`
	Name        types.Map    `tfsdk:"name"`
	Description types.Map    `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Metadata    types.Object `tfsdk:"metadata"`
	Values      types.List   `tfsdk:"values"`
	Attributes  types.List   `tfsdk:"attributes"`
	ArrayType   types.Object `tfsdk:"array_type"`
}

// DeeplyNestedSchemaAttributeModel describes deeply nested attribute model (3rd level, supports one more level)
type DeeplyNestedSchemaAttributeModel struct {
	Key         types.String `tfsdk:"key"`
	Name        types.Map    `tfsdk:"name"`
	Description types.Map    `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Metadata    types.Object `tfsdk:"metadata"`
	Values      types.List   `tfsdk:"values"`
	Attributes  types.List   `tfsdk:"attributes"`
	ArrayType   types.Object `tfsdk:"array_type"`
}

// Level4SchemaAttributeModel describes 4th level nested attribute model (no further nesting)
type Level4SchemaAttributeModel struct {
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
										MarkdownDescription: "Deeply nested attributes for OBJECT type (third level).",
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
													MarkdownDescription: "Fourth level nested attributes for OBJECT type.",
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

				// Convert nested values (for ENUM or REFERENCE types)
				var nestedValuesList types.List
				if len(nestedAttr.Values) > 0 {
					nestedValueModels := make([]SchemaAttributeValueModel, len(nestedAttr.Values))
					for k, val := range nestedAttr.Values {
						nestedValueModels[k] = SchemaAttributeValueModel{
							Value: types.StringValue(val.Value),
						}
					}
					nestedValuesList, d = types.ListValueFrom(ctx, types.ObjectType{
						AttrTypes: getSchemaAttributeValueAttrTypes(),
					}, nestedValueModels)
					diags.Append(d...)
				} else {
					nestedValuesList = types.ListNull(types.ObjectType{
						AttrTypes: getSchemaAttributeValueAttrTypes(),
					})
				}

				// Convert deeply nested attributes (for OBJECT type within OBJECT)
				var deeplyNestedAttrsList types.List
				if len(nestedAttr.Attributes) > 0 {
					deeplyNestedModels := make([]DeeplyNestedSchemaAttributeModel, len(nestedAttr.Attributes))
					for k, deepAttr := range nestedAttr.Attributes {
						deepNameMap, d := types.MapValueFrom(ctx, types.StringType, deepAttr.Name)
						diags.Append(d...)

						var deepDescMap types.Map
						if len(deepAttr.Description) > 0 {
							deepDescMap, d = types.MapValueFrom(ctx, types.StringType, deepAttr.Description)
							diags.Append(d...)
						} else {
							deepDescMap = types.MapNull(types.StringType)
						}

						var deepMetadataObj types.Object
						if deepAttr.Metadata != nil {
							deepMetadataModel := SchemaAttributeMetadataModel{
								ReadOnly:  types.BoolValue(deepAttr.Metadata.ReadOnly),
								Localized: types.BoolValue(deepAttr.Metadata.Localized),
								Required:  types.BoolValue(deepAttr.Metadata.Required),
								Nullable:  types.BoolValue(deepAttr.Metadata.Nullable),
							}
							deepMetadataObj, d = types.ObjectValueFrom(ctx, getSchemaAttributeMetadataAttrTypes(), deepMetadataModel)
							diags.Append(d...)
						} else {
							deepMetadataObj = types.ObjectNull(getSchemaAttributeMetadataAttrTypes())
						}

						// Convert deep values (for ENUM or REFERENCE types at 3rd level)
						var deepValuesList types.List
						if len(deepAttr.Values) > 0 {
							deepValueModels := make([]SchemaAttributeValueModel, len(deepAttr.Values))
							for l, val := range deepAttr.Values {
								deepValueModels[l] = SchemaAttributeValueModel{
									Value: types.StringValue(val.Value),
								}
							}
							deepValuesList, d = types.ListValueFrom(ctx, types.ObjectType{
								AttrTypes: getSchemaAttributeValueAttrTypes(),
							}, deepValueModels)
							diags.Append(d...)
						} else {
							deepValuesList = types.ListNull(types.ObjectType{
								AttrTypes: getSchemaAttributeValueAttrTypes(),
							})
						}

						// Convert 4th level attributes (for OBJECT type at 3rd level)
						var level4AttrsList types.List
						if len(deepAttr.Attributes) > 0 {
							level4Models := make([]Level4SchemaAttributeModel, len(deepAttr.Attributes))
							for l, l4Attr := range deepAttr.Attributes {
								l4NameMap, d := types.MapValueFrom(ctx, types.StringType, l4Attr.Name)
								diags.Append(d...)

								var l4DescMap types.Map
								if len(l4Attr.Description) > 0 {
									l4DescMap, d = types.MapValueFrom(ctx, types.StringType, l4Attr.Description)
									diags.Append(d...)
								} else {
									l4DescMap = types.MapNull(types.StringType)
								}

								var l4MetadataObj types.Object
								if l4Attr.Metadata != nil {
									l4MetadataModel := SchemaAttributeMetadataModel{
										ReadOnly:  types.BoolValue(l4Attr.Metadata.ReadOnly),
										Localized: types.BoolValue(l4Attr.Metadata.Localized),
										Required:  types.BoolValue(l4Attr.Metadata.Required),
										Nullable:  types.BoolValue(l4Attr.Metadata.Nullable),
									}
									l4MetadataObj, d = types.ObjectValueFrom(ctx, getSchemaAttributeMetadataAttrTypes(), l4MetadataModel)
									diags.Append(d...)
								} else {
									l4MetadataObj = types.ObjectNull(getSchemaAttributeMetadataAttrTypes())
								}

								level4Models[l] = Level4SchemaAttributeModel{
									Key:         types.StringValue(l4Attr.Key),
									Name:        l4NameMap,
									Description: l4DescMap,
									Type:        types.StringValue(l4Attr.Type),
									Metadata:    l4MetadataObj,
								}
							}
							level4AttrsList, d = types.ListValueFrom(ctx, types.ObjectType{
								AttrTypes: getLevel4SchemaAttributeAttrTypes(),
							}, level4Models)
							diags.Append(d...)
						} else {
							level4AttrsList = types.ListNull(types.ObjectType{
								AttrTypes: getLevel4SchemaAttributeAttrTypes(),
							})
						}

						// Convert deep array_type (for ARRAY type at 3rd level)
						var deepArrayTypeObj types.Object
						if deepAttr.ArrayType != nil {
							var deepArrayValuesList types.List
							if len(deepAttr.ArrayType.Values) > 0 {
								deepArrayValueModels := make([]SchemaAttributeValueModel, len(deepAttr.ArrayType.Values))
								for l, val := range deepAttr.ArrayType.Values {
									deepArrayValueModels[l] = SchemaAttributeValueModel{
										Value: types.StringValue(val.Value),
									}
								}
								deepArrayValuesList, d = types.ListValueFrom(ctx, types.ObjectType{
									AttrTypes: getSchemaAttributeValueAttrTypes(),
								}, deepArrayValueModels)
								diags.Append(d...)
							} else {
								deepArrayValuesList = types.ListNull(types.ObjectType{
									AttrTypes: getSchemaAttributeValueAttrTypes(),
								})
							}

							deepArrayTypeModel := SchemaArrayTypeModel{
								Type:      types.StringValue(deepAttr.ArrayType.Type),
								Localized: types.BoolValue(deepAttr.ArrayType.Localized),
								Values:    deepArrayValuesList,
							}
							deepArrayTypeObj, d = types.ObjectValueFrom(ctx, getSchemaArrayTypeAttrTypes(), deepArrayTypeModel)
							diags.Append(d...)
						} else {
							deepArrayTypeObj = types.ObjectNull(getSchemaArrayTypeAttrTypes())
						}

						deeplyNestedModels[k] = DeeplyNestedSchemaAttributeModel{
							Key:         types.StringValue(deepAttr.Key),
							Name:        deepNameMap,
							Description: deepDescMap,
							Type:        types.StringValue(deepAttr.Type),
							Metadata:    deepMetadataObj,
							Values:      deepValuesList,
							Attributes:  level4AttrsList,
							ArrayType:   deepArrayTypeObj,
						}
					}
					deeplyNestedAttrsList, d = types.ListValueFrom(ctx, types.ObjectType{
						AttrTypes: getDeeplyNestedSchemaAttributeAttrTypes(),
					}, deeplyNestedModels)
					diags.Append(d...)
				} else {
					deeplyNestedAttrsList = types.ListNull(types.ObjectType{
						AttrTypes: getDeeplyNestedSchemaAttributeAttrTypes(),
					})
				}

				// Convert nested array_type (for ARRAY type within OBJECT)
				var nestedArrayTypeObj types.Object
				if nestedAttr.ArrayType != nil {
					var nestedArrayValuesList types.List
					if len(nestedAttr.ArrayType.Values) > 0 {
						nestedArrayValueModels := make([]SchemaAttributeValueModel, len(nestedAttr.ArrayType.Values))
						for k, val := range nestedAttr.ArrayType.Values {
							nestedArrayValueModels[k] = SchemaAttributeValueModel{
								Value: types.StringValue(val.Value),
							}
						}
						nestedArrayValuesList, d = types.ListValueFrom(ctx, types.ObjectType{
							AttrTypes: getSchemaAttributeValueAttrTypes(),
						}, nestedArrayValueModels)
						diags.Append(d...)
					} else {
						nestedArrayValuesList = types.ListNull(types.ObjectType{
							AttrTypes: getSchemaAttributeValueAttrTypes(),
						})
					}

					nestedArrayTypeModel := SchemaArrayTypeModel{
						Type:      types.StringValue(nestedAttr.ArrayType.Type),
						Localized: types.BoolValue(nestedAttr.ArrayType.Localized),
						Values:    nestedArrayValuesList,
					}
					nestedArrayTypeObj, d = types.ObjectValueFrom(ctx, getSchemaArrayTypeAttrTypes(), nestedArrayTypeModel)
					diags.Append(d...)
				} else {
					nestedArrayTypeObj = types.ObjectNull(getSchemaArrayTypeAttrTypes())
				}

				nestedModels[j] = NestedSchemaAttributeModel{
					Key:         types.StringValue(nestedAttr.Key),
					Name:        nestedNameMap,
					Description: nestedDescMap,
					Type:        types.StringValue(nestedAttr.Type),
					Metadata:    nestedMetadataObj,
					Values:      nestedValuesList,
					Attributes:  deeplyNestedAttrsList,
					ArrayType:   nestedArrayTypeObj,
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

				// Convert nested values (for ENUM or REFERENCE types)
				var nestedValues []SchemaAttributeValue
				if !nestedModel.Values.IsNull() {
					var nestedValueModels []SchemaAttributeValueModel
					diags.Append(nestedModel.Values.ElementsAs(ctx, &nestedValueModels, false)...)
					nestedValues = make([]SchemaAttributeValue, len(nestedValueModels))
					for k, valModel := range nestedValueModels {
						nestedValues[k] = SchemaAttributeValue{
							Value: valModel.Value.ValueString(),
						}
					}
				}

				// Convert deeply nested attributes (for OBJECT type within OBJECT)
				var deeplyNestedAttrs []SchemaAttribute
				if !nestedModel.Attributes.IsNull() {
					var deepModels []DeeplyNestedSchemaAttributeModel
					diags.Append(nestedModel.Attributes.ElementsAs(ctx, &deepModels, false)...)
					deeplyNestedAttrs = make([]SchemaAttribute, len(deepModels))
					for k, deepModel := range deepModels {
						var deepNameMap map[string]string
						diags.Append(deepModel.Name.ElementsAs(ctx, &deepNameMap, false)...)

						var deepDescMap map[string]string
						if !deepModel.Description.IsNull() {
							diags.Append(deepModel.Description.ElementsAs(ctx, &deepDescMap, false)...)
						}

						var deepMetadataModel SchemaAttributeMetadataModel
						diags.Append(deepModel.Metadata.As(ctx, &deepMetadataModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
						deepMetadata := &SchemaAttributeMetadata{
							ReadOnly:  deepMetadataModel.ReadOnly.ValueBool(),
							Localized: deepMetadataModel.Localized.ValueBool(),
							Required:  deepMetadataModel.Required.ValueBool(),
							Nullable:  deepMetadataModel.Nullable.ValueBool(),
						}

						// Convert deep values (for ENUM or REFERENCE types at 3rd level)
						var deepValues []SchemaAttributeValue
						if !deepModel.Values.IsNull() {
							var deepValueModels []SchemaAttributeValueModel
							diags.Append(deepModel.Values.ElementsAs(ctx, &deepValueModels, false)...)
							deepValues = make([]SchemaAttributeValue, len(deepValueModels))
							for l, valModel := range deepValueModels {
								deepValues[l] = SchemaAttributeValue{
									Value: valModel.Value.ValueString(),
								}
							}
						}

						// Convert 4th level attributes (for OBJECT type at 3rd level)
						var level4Attrs []SchemaAttribute
						if !deepModel.Attributes.IsNull() {
							var l4Models []Level4SchemaAttributeModel
							diags.Append(deepModel.Attributes.ElementsAs(ctx, &l4Models, false)...)
							level4Attrs = make([]SchemaAttribute, len(l4Models))
							for l, l4Model := range l4Models {
								var l4NameMap map[string]string
								diags.Append(l4Model.Name.ElementsAs(ctx, &l4NameMap, false)...)

								var l4DescMap map[string]string
								if !l4Model.Description.IsNull() {
									diags.Append(l4Model.Description.ElementsAs(ctx, &l4DescMap, false)...)
								}

								var l4MetadataModel SchemaAttributeMetadataModel
								diags.Append(l4Model.Metadata.As(ctx, &l4MetadataModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
								l4Metadata := &SchemaAttributeMetadata{
									ReadOnly:  l4MetadataModel.ReadOnly.ValueBool(),
									Localized: l4MetadataModel.Localized.ValueBool(),
									Required:  l4MetadataModel.Required.ValueBool(),
									Nullable:  l4MetadataModel.Nullable.ValueBool(),
								}

								level4Attrs[l] = SchemaAttribute{
									Key:         l4Model.Key.ValueString(),
									Name:        l4NameMap,
									Description: l4DescMap,
									Type:        l4Model.Type.ValueString(),
									Metadata:    l4Metadata,
								}
							}
						}

						// Convert deep array_type (for ARRAY type at 3rd level)
						var deepArrayType *SchemaArrayType
						if !deepModel.ArrayType.IsNull() {
							var deepArrayTypeModel SchemaArrayTypeModel
							diags.Append(deepModel.ArrayType.As(ctx, &deepArrayTypeModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)

							deepArrayType = &SchemaArrayType{
								Type:      deepArrayTypeModel.Type.ValueString(),
								Localized: deepArrayTypeModel.Localized.ValueBool(),
							}

							if !deepArrayTypeModel.Values.IsNull() {
								var deepArrayValueModels []SchemaAttributeValueModel
								diags.Append(deepArrayTypeModel.Values.ElementsAs(ctx, &deepArrayValueModels, false)...)
								deepArrayType.Values = make([]SchemaAttributeValue, len(deepArrayValueModels))
								for l, valModel := range deepArrayValueModels {
									deepArrayType.Values[l] = SchemaAttributeValue{
										Value: valModel.Value.ValueString(),
									}
								}
							}
						}

						deeplyNestedAttrs[k] = SchemaAttribute{
							Key:         deepModel.Key.ValueString(),
							Name:        deepNameMap,
							Description: deepDescMap,
							Type:        deepModel.Type.ValueString(),
							Metadata:    deepMetadata,
							Values:      deepValues,
							Attributes:  level4Attrs,
							ArrayType:   deepArrayType,
						}
					}
				}

				// Convert nested array_type (for ARRAY type within OBJECT)
				var nestedArrayType *SchemaArrayType
				if !nestedModel.ArrayType.IsNull() {
					var nestedArrayTypeModel SchemaArrayTypeModel
					diags.Append(nestedModel.ArrayType.As(ctx, &nestedArrayTypeModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)

					nestedArrayType = &SchemaArrayType{
						Type:      nestedArrayTypeModel.Type.ValueString(),
						Localized: nestedArrayTypeModel.Localized.ValueBool(),
					}

					if !nestedArrayTypeModel.Values.IsNull() {
						var nestedArrayValueModels []SchemaAttributeValueModel
						diags.Append(nestedArrayTypeModel.Values.ElementsAs(ctx, &nestedArrayValueModels, false)...)
						nestedArrayType.Values = make([]SchemaAttributeValue, len(nestedArrayValueModels))
						for k, valModel := range nestedArrayValueModels {
							nestedArrayType.Values[k] = SchemaAttributeValue{
								Value: valModel.Value.ValueString(),
							}
						}
					}
				}

				nestedAttrs[j] = SchemaAttribute{
					Key:         nestedModel.Key.ValueString(),
					Name:        nestedNameMap,
					Description: nestedDescMap,
					Type:        nestedModel.Type.ValueString(),
					Metadata:    nestedMetadata,
					Values:      nestedValues,
					Attributes:  deeplyNestedAttrs,
					ArrayType:   nestedArrayType,
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
		"values":      types.ListType{ElemType: types.ObjectType{AttrTypes: getSchemaAttributeValueAttrTypes()}},
		"attributes":  types.ListType{ElemType: types.ObjectType{AttrTypes: getDeeplyNestedSchemaAttributeAttrTypes()}},
		"array_type":  types.ObjectType{AttrTypes: getSchemaArrayTypeAttrTypes()},
	}
}

func getDeeplyNestedSchemaAttributeAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"key":         types.StringType,
		"name":        types.MapType{ElemType: types.StringType},
		"description": types.MapType{ElemType: types.StringType},
		"type":        types.StringType,
		"metadata":    types.ObjectType{AttrTypes: getSchemaAttributeMetadataAttrTypes()},
		"values":      types.ListType{ElemType: types.ObjectType{AttrTypes: getSchemaAttributeValueAttrTypes()}},
		"attributes":  types.ListType{ElemType: types.ObjectType{AttrTypes: getLevel4SchemaAttributeAttrTypes()}},
		"array_type":  types.ObjectType{AttrTypes: getSchemaArrayTypeAttrTypes()},
	}
}

func getLevel4SchemaAttributeAttrTypes() map[string]attr.Type {
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
