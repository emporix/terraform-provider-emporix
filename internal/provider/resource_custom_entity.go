package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CustomEntityResource{}
var _ resource.ResourceWithImportState = &CustomEntityResource{}
var _ resource.ResourceWithValidateConfig = &CustomEntityResource{}

func NewCustomEntityResource() resource.Resource {
	return &CustomEntityResource{}
}

// CustomEntityResource defines the resource implementation.
type CustomEntityResource struct {
	client *EmporixClient
}

// CustomEntityResourceModel describes the resource data model.
type CustomEntityResourceModel struct {
	Type       types.String `tfsdk:"type"`
	ID         types.String `tfsdk:"id"`
	Name       types.Map    `tfsdk:"name"`
	Owner      types.Object `tfsdk:"owner"`
	Mixins     types.String `tfsdk:"mixins"`
	Media      types.List   `tfsdk:"media"`
	CreatedAt  types.String `tfsdk:"created_at"`
	ModifiedAt types.String `tfsdk:"modified_at"`
	Version    types.Int64  `tfsdk:"version"`
}

// CustomEntityOwnerModel describes the nested "owner" attribute.
type CustomEntityOwnerModel struct {
	Type          types.String `tfsdk:"type"`
	UserID        types.String `tfsdk:"user_id"`
	LegalEntityID types.String `tfsdk:"legal_entity_id"`
}

func (CustomEntityOwnerModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":            types.StringType,
		"user_id":         types.StringType,
		"legal_entity_id": types.StringType,
	}
}

func (r *CustomEntityResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_entity"
}

func (r *CustomEntityResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a custom entity instance in Emporix. " +
			"Custom entity instances are data records that live under a custom schema type, managed via the `emporix_custom_entity_type` resource (a distinct resource from `emporix_schema`). " +
			"The `type` argument must match the `id` of such an existing `emporix_custom_entity_type` resource (e.g. `emporix_custom_entity_type.document.id`). " +
			"You may optionally also define an `emporix_schema` resource to enforce a validated structure for instances - reference this type by setting the schema's `types` to the custom type's own id (e.g. `types = [emporix_custom_entity_type.document.id]`), not the generic `CUSTOM_ENTITY` literal. This is not required for basic usage. " +
			"Managing this resource requires the `schema.custominstance_manage` OAuth scope (or the per-type `custom.<lowercase-type>_manage` scope, note the type is lowercased in the scope name) on the tenant's API client; reads require `schema.custominstance_read` or `custom.<lowercase-type>_read`. " +
			"Both `type` and `owner` are immutable after creation: changing either forces replacement. " +
			"Import using the format `type:id` (e.g. `DOCUMENT:doc-123`).",

		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				MarkdownDescription: "The custom schema type this instance belongs to - the `id` of an existing `emporix_custom_entity_type` resource (e.g. \"DOCUMENT\"). Case-sensitive; used as a URL path segment. Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Custom entity instance identifier. If not provided, the API will generate one automatically. Cannot be changed after creation.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.MapAttribute{
				MarkdownDescription: "Display name as a map of language code to name (e.g., {\"en\": \"My Document\"}). Provide at least one language translation.",
				ElementType:         types.StringType,
				Required:            true,
			},
			"owner": schema.SingleNestedAttribute{
				MarkdownDescription: "Ownership of this instance. Once set, ownership cannot be changed; modifying it forces replacement.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Owner type. Valid values when setting an owner: `EMPLOYEE`, `CUSTOMER`. (`SERVICE` can appear when reading back an instance whose owner was auto-assigned by the API under a `manage_own` scope, but cannot be set explicitly.)",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("EMPLOYEE", "CUSTOMER"),
						},
					},
					"user_id": schema.StringAttribute{
						MarkdownDescription: "ID of the owning user. Required when `owner` is set.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
					"legal_entity_id": schema.StringAttribute{
						MarkdownDescription: "Legal entity ID of the owner. Only valid when `owner.type` is `CUSTOMER`.",
						Optional:            true,
					},
				},
			},
			"mixins": schema.StringAttribute{
				MarkdownDescription: "Arbitrary instance data as a JSON-encoded string (e.g. `jsonencode({...})`). Defaults to an empty object. " +
					"Note that the API stores this as a JSON-encoded string value, not a nested object, so its content is opaque to the API unless validated by an attached `emporix_schema`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("{}"),
			},
			"media": schema.ListAttribute{
				MarkdownDescription: "IDs of media assets assigned to this instance. Read-only here; media is assigned through Emporix's media management APIs, not through this resource.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the instance was created.",
				Computed:            true,
			},
			"modified_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the instance was last modified.",
				Computed:            true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Instance version (managed by the API).",
				Computed:            true,
			},
		},
	}
}

func (r *CustomEntityResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CustomEntityResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var owner types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("owner"), &owner)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if owner.IsNull() || owner.IsUnknown() {
		return
	}

	var ownerModel CustomEntityOwnerModel
	resp.Diagnostics.Append(owner.As(ctx, &ownerModel, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if ownerModel.Type.IsUnknown() {
		return
	}

	legalEntitySet := !ownerModel.LegalEntityID.IsNull() && !ownerModel.LegalEntityID.IsUnknown() && ownerModel.LegalEntityID.ValueString() != ""

	if legalEntitySet && ownerModel.Type.ValueString() != "CUSTOMER" {
		resp.Diagnostics.AddAttributeError(
			path.Root("owner").AtName("legal_entity_id"),
			"Invalid Owner Configuration",
			"owner.legal_entity_id can only be set when owner.type is \"CUSTOMER\".",
		)
	}
}

func (r *CustomEntityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CustomEntityResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entityType := data.Type.ValueString()

	tflog.Debug(ctx, "Creating custom entity", map[string]interface{}{
		"type": entityType,
	})

	nameMap := make(map[string]string)
	resp.Diagnostics.Append(data.Name.ElementsAs(ctx, &nameMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner, diags := customEntityOwnerFromModel(ctx, data.Owner)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mixins := data.Mixins.ValueString()
	if !json.Valid([]byte(mixins)) {
		resp.Diagnostics.AddError("Invalid JSON", "Unable to parse mixins: not valid JSON.")
		return
	}

	createData := &CustomEntityCreate{
		ID:     data.ID.ValueString(),
		Name:   nameMap,
		Owner:  owner,
		Mixins: mixins,
	}

	instance, err := r.client.CreateCustomEntity(ctx, entityType, createData)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create custom entity, got error: %s", err))
		return
	}

	mapCustomEntityToModel(ctx, instance, entityType, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomEntityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CustomEntityResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entityType := data.Type.ValueString()

	tflog.Debug(ctx, "Reading custom entity", map[string]interface{}{
		"type": entityType,
		"id":   data.ID.ValueString(),
	})

	instance, err := r.client.GetCustomEntity(ctx, entityType, data.ID.ValueString())
	if err != nil {
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read custom entity, got error: %s", err))
		return
	}

	mapCustomEntityToModel(ctx, instance, entityType, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomEntityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CustomEntityResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entityType := data.Type.ValueString()

	tflog.Debug(ctx, "Updating custom entity", map[string]interface{}{
		"type": entityType,
		"id":   data.ID.ValueString(),
	})

	nameMap := make(map[string]string)
	resp.Diagnostics.Append(data.Name.ElementsAs(ctx, &nameMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner, diags := customEntityOwnerFromModel(ctx, data.Owner)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mixins := data.Mixins.ValueString()
	if !json.Valid([]byte(mixins)) {
		resp.Diagnostics.AddError("Invalid JSON", "Unable to parse mixins: not valid JSON.")
		return
	}

	updateData := &CustomEntityUpdate{
		ID:     data.ID.ValueString(),
		Name:   nameMap,
		Owner:  owner,
		Mixins: mixins,
	}

	instance, err := r.client.UpdateCustomEntity(ctx, entityType, data.ID.ValueString(), updateData)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update custom entity, got error: %s", err))
		return
	}

	mapCustomEntityToModel(ctx, instance, entityType, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomEntityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CustomEntityResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting custom entity", map[string]interface{}{
		"type": data.Type.ValueString(),
		"id":   data.ID.ValueString(),
	})

	if err := r.client.DeleteCustomEntity(ctx, data.Type.ValueString(), data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete custom entity, got error: %s", err))
		return
	}

	// Custom entity is now deleted and will be removed from Terraform state
}

func (r *CustomEntityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: "type:id" (e.g. "DOCUMENT:doc-123")
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'type:id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// customEntityOwnerFromModel converts the "owner" attribute from plan/config into an API payload.
func customEntityOwnerFromModel(ctx context.Context, owner types.Object) (*CustomEntityOwner, diag.Diagnostics) {
	var diags diag.Diagnostics

	if owner.IsNull() || owner.IsUnknown() {
		return nil, diags
	}

	var ownerModel CustomEntityOwnerModel
	diags.Append(owner.As(ctx, &ownerModel, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}

	return &CustomEntityOwner{
		Type:          ownerModel.Type.ValueString(),
		UserID:        ownerModel.UserID.ValueString(),
		LegalEntityID: ownerModel.LegalEntityID.ValueString(),
	}, diags
}

// mapCustomEntityToModel converts a CustomEntityInstance API response into a CustomEntityResourceModel.
func mapCustomEntityToModel(ctx context.Context, instance *CustomEntityInstance, entityType string, data *CustomEntityResourceModel, diags *diag.Diagnostics) {
	data.ID = types.StringValue(instance.ID)
	data.Type = types.StringValue(entityType)

	if len(instance.Name) > 0 {
		nameMapValue, d := types.MapValueFrom(ctx, types.StringType, instance.Name)
		diags.Append(d...)
		data.Name = nameMapValue
	} else {
		nameMapValue, d := types.MapValueFrom(ctx, types.StringType, map[string]string{})
		diags.Append(d...)
		data.Name = nameMapValue
	}

	if instance.Owner != nil {
		userID := types.StringNull()
		if instance.Owner.UserID != "" {
			userID = types.StringValue(instance.Owner.UserID)
		}
		legalEntityID := types.StringNull()
		if instance.Owner.LegalEntityID != "" {
			legalEntityID = types.StringValue(instance.Owner.LegalEntityID)
		}

		ownerObj, d := types.ObjectValueFrom(ctx, CustomEntityOwnerModel{}.AttributeTypes(), CustomEntityOwnerModel{
			Type:          types.StringValue(instance.Owner.Type),
			UserID:        userID,
			LegalEntityID: legalEntityID,
		})
		diags.Append(d...)
		data.Owner = ownerObj
	} else {
		data.Owner = types.ObjectNull(CustomEntityOwnerModel{}.AttributeTypes())
	}

	if instance.Mixins != "" {
		data.Mixins = types.StringValue(instance.Mixins)
	} else {
		data.Mixins = types.StringValue("{}")
	}

	mediaValue, d := types.ListValueFrom(ctx, types.StringType, instance.Media)
	diags.Append(d...)
	data.Media = mediaValue

	if instance.Metadata != nil {
		data.Version = types.Int64Value(int64(instance.Metadata.Version))

		data.CreatedAt = types.StringNull()
		if instance.Metadata.CreatedAt != "" {
			data.CreatedAt = types.StringValue(instance.Metadata.CreatedAt)
		}

		data.ModifiedAt = types.StringNull()
		if instance.Metadata.ModifiedAt != "" {
			data.ModifiedAt = types.StringValue(instance.Metadata.ModifiedAt)
		}
	} else {
		data.Version = types.Int64Null()
		data.CreatedAt = types.StringNull()
		data.ModifiedAt = types.StringNull()
	}
}
