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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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
var _ resource.Resource = &CustomEntityInstanceResource{}
var _ resource.ResourceWithImportState = &CustomEntityInstanceResource{}
var _ resource.ResourceWithValidateConfig = &CustomEntityInstanceResource{}

func NewCustomEntityInstanceResource() resource.Resource {
	return &CustomEntityInstanceResource{}
}

// CustomEntityInstanceResource defines the resource implementation.
type CustomEntityInstanceResource struct {
	client *EmporixClient
}

// CustomEntityInstanceResourceModel describes the resource data model.
type CustomEntityInstanceResourceModel struct {
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

func (r *CustomEntityInstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_entity_instance"
}

func (r *CustomEntityInstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a custom entity instance in Emporix. " +
			"Custom entity instances are data records that live under a custom schema type, managed via the `emporix_custom_entity_type` resource (a distinct resource from `emporix_schema`). " +
			"The `type` argument must match the `id` of such an existing `emporix_custom_entity_type` resource. " +
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
				MarkdownDescription: "Ownership of this instance. Cannot be changed after creation.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Type of the owner. Valid values: `EMPLOYEE`, `CUSTOMER`.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("EMPLOYEE", "CUSTOMER"),
						},
					},
					"user_id": schema.StringAttribute{
						MarkdownDescription: "Identifier of the employee or customer associated with the owner. Must be a real, existing user ID on your tenant.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
					"legal_entity_id": schema.StringAttribute{
						MarkdownDescription: "Legal entity identifier. Can be provided only when `type` is `CUSTOMER`.",
						Optional:            true,
					},
				},
			},
			"mixins": schema.StringAttribute{
				MarkdownDescription: "Instance data as a JSON-encoded string (e.g. `jsonencode({...})`). Defaults to an empty object. " +
					"Each field must be nested under a top-level key equal to the `id` of the `emporix_schema` that declares it, e.g. `jsonencode({\"document-fields\" = {note = \"hello\"}})`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("{}"),
			},
			"media": schema.ListAttribute{
				MarkdownDescription: "IDs of media assets assigned to this instance. Read-only here; media is assigned through Emporix's media management APIs, not through this resource.",
				ElementType:         types.StringType,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the instance was created.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"modified_at": schema.StringAttribute{
				// No UseStateForUnknown: the API bumps this on every write, not just when this attribute changes.
				MarkdownDescription: "Timestamp when the instance was last modified.",
				Computed:            true,
			},
			"version": schema.Int64Attribute{
				// No UseStateForUnknown - same reason as modified_at above.
				MarkdownDescription: "Instance version (managed by the API).",
				Computed:            true,
			},
		},
	}
}

func (r *CustomEntityInstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CustomEntityInstanceResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
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
			"owner.legal_entity_id can only be provided when owner.type is \"CUSTOMER\".",
		)
	}
}

func (r *CustomEntityInstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CustomEntityInstanceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entityType := data.Type.ValueString()

	tflog.Debug(ctx, "Creating custom entity instance", map[string]interface{}{
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

	var mixins map[string]interface{}
	if err := json.Unmarshal([]byte(data.Mixins.ValueString()), &mixins); err != nil {
		resp.Diagnostics.AddError("Invalid JSON", fmt.Sprintf("Unable to parse mixins as JSON: %s", err))
		return
	}

	createData := &CustomEntityInstanceCreate{
		ID:     data.ID.ValueString(),
		Name:   nameMap,
		Owner:  owner,
		Mixins: mixins,
	}

	instance, err := r.client.CreateCustomEntityInstance(ctx, entityType, createData)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create custom entity instance, got error: %s", err))
		return
	}

	mapCustomEntityInstanceToModel(ctx, instance, entityType, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomEntityInstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CustomEntityInstanceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entityType := data.Type.ValueString()

	tflog.Debug(ctx, "Reading custom entity instance", map[string]interface{}{
		"type": entityType,
		"id":   data.ID.ValueString(),
	})

	instance, err := r.client.GetCustomEntityInstance(ctx, entityType, data.ID.ValueString())
	if err != nil {
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read custom entity instance, got error: %s", err))
		return
	}

	mapCustomEntityInstanceToModel(ctx, instance, entityType, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomEntityInstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CustomEntityInstanceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CustomEntityInstanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entityType := data.Type.ValueString()

	tflog.Debug(ctx, "Updating custom entity instance", map[string]interface{}{
		"type": entityType,
		"id":   data.ID.ValueString(),
	})

	if data.Name.Equal(state.Name) && data.Owner.Equal(state.Owner) && data.Mixins.Equal(state.Mixins) {
		// Nothing to write - refresh from the API instead of sending a no-op PUT that
		// would needlessly bump version/modified_at.
		instance, err := r.client.GetCustomEntityInstance(ctx, entityType, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read custom entity instance, got error: %s", err))
			return
		}
		mapCustomEntityInstanceToModel(ctx, instance, entityType, &data, &resp.Diagnostics)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

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

	var mixins map[string]interface{}
	if err := json.Unmarshal([]byte(data.Mixins.ValueString()), &mixins); err != nil {
		resp.Diagnostics.AddError("Invalid JSON", fmt.Sprintf("Unable to parse mixins as JSON: %s", err))
		return
	}

	updateData := &CustomEntityInstanceUpdate{
		ID:     data.ID.ValueString(),
		Name:   nameMap,
		Owner:  owner,
		Mixins: mixins,
	}

	instance, err := r.client.UpdateCustomEntityInstance(ctx, entityType, data.ID.ValueString(), updateData)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update custom entity instance, got error: %s", err))
		return
	}

	mapCustomEntityInstanceToModel(ctx, instance, entityType, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomEntityInstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CustomEntityInstanceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting custom entity instance", map[string]interface{}{
		"type": data.Type.ValueString(),
		"id":   data.ID.ValueString(),
	})

	if err := r.client.DeleteCustomEntityInstance(ctx, data.Type.ValueString(), data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete custom entity instance, got error: %s", err))
		return
	}

	// Custom entity instance is now deleted and will be removed from Terraform state
}

func (r *CustomEntityInstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

// mapCustomEntityInstanceToModel converts a CustomEntityInstance API response into a CustomEntityInstanceResourceModel.
func mapCustomEntityInstanceToModel(ctx context.Context, instance *CustomEntityInstance, entityType string, data *CustomEntityInstanceResourceModel, diags *diag.Diagnostics) {
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

	if instance.Mixins != nil {
		mixinsJSON, err := json.Marshal(instance.Mixins)
		if err != nil {
			diags.AddError("JSON Error", fmt.Sprintf("Unable to marshal mixins to JSON: %s", err))
		} else {
			data.Mixins = types.StringValue(string(mixinsJSON))
		}
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
