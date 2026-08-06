package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CustomEntityTypeResource{}
var _ resource.ResourceWithImportState = &CustomEntityTypeResource{}

func NewCustomEntityTypeResource() resource.Resource {
	return &CustomEntityTypeResource{}
}

// CustomEntityTypeResource defines the resource implementation.
type CustomEntityTypeResource struct {
	client *EmporixClient
}

// CustomEntityTypeResourceModel describes the resource data model.
type CustomEntityTypeResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.Map    `tfsdk:"name"`
	CreatedAt  types.String `tfsdk:"created_at"`
	ModifiedAt types.String `tfsdk:"modified_at"`
	Version    types.Int64  `tfsdk:"version"`
}

func (r *CustomEntityTypeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_entity_type"
}

func (r *CustomEntityTypeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a custom schema type in Emporix - the container that custom entity instances (`emporix_custom_entity`) belong to. " +
			"This is a distinct resource from `emporix_schema`: it registers the *type* (e.g. \"DOCUMENT\") under which instances live at `/schema/{tenant}/custom-entities/{type}/instances`, " +
			"and its creation auto-generates the per-type OAuth scopes `custom.<lowercase-id>_manage`, `custom.<lowercase-id>_manage_own`, `custom.<lowercase-id>_read`, and `custom.<lowercase-id>_read_own`. " +
			"Managing this resource requires the `schema.schema_manage` OAuth scope. " +
			"The `id` is immutable after creation and cannot be `AVAILABILITY` or `LOCATION` (reserved). " +
			"Deletion fails if any `emporix_schema` or `emporix_custom_entity` resources still reference this type.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique code for the custom type. Must start with an uppercase letter and contain only uppercase letters, digits, and underscores (e.g. \"DOCUMENT\"). Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`),
						"must start with an uppercase letter and contain only uppercase letters, digits, and underscores",
					),
					stringvalidator.NoneOf("AVAILABILITY", "LOCATION"),
				},
			},
			"name": schema.MapAttribute{
				MarkdownDescription: "Localized custom type name as a map of language code to name (e.g., {\"en\": \"Document\"}). Provide at least one language translation.",
				ElementType:         types.StringType,
				Required:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the custom type was created.",
				Computed:            true,
			},
			"modified_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the custom type was last modified.",
				Computed:            true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Custom type version, used for optimistic locking on updates (managed by the API).",
				Computed:            true,
			},
		},
	}
}

func (r *CustomEntityTypeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CustomEntityTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CustomEntityTypeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating custom entity type", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	nameMap := make(map[string]string)
	resp.Diagnostics.Append(data.Name.ElementsAs(ctx, &nameMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entityType, err := r.client.CreateCustomEntityType(ctx, &CustomEntityTypeCreate{
		ID:   data.ID.ValueString(),
		Name: nameMap,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create custom entity type, got error: %s", err))
		return
	}

	mapCustomEntityTypeToModel(ctx, entityType, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomEntityTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CustomEntityTypeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading custom entity type", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	entityType, err := r.client.GetCustomEntityType(ctx, data.ID.ValueString())
	if err != nil {
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read custom entity type, got error: %s", err))
		return
	}

	mapCustomEntityTypeToModel(ctx, entityType, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomEntityTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CustomEntityTypeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating custom entity type", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	nameMap := make(map[string]string)
	resp.Diagnostics.Append(data.Name.ElementsAs(ctx, &nameMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch current version for optimistic locking (required by PUT)
	current, err := r.client.GetCustomEntityType(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read current custom entity type state, got error: %s", err))
		return
	}

	updateData := &CustomEntityTypeUpdate{
		Name: nameMap,
	}
	if current.Metadata != nil && current.Metadata.Version > 0 {
		updateData.Metadata = &SchemaMetadataUpdate{Version: current.Metadata.Version}
	}

	entityType, err := r.client.UpdateCustomEntityType(ctx, data.ID.ValueString(), updateData)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update custom entity type, got error: %s", err))
		return
	}

	mapCustomEntityTypeToModel(ctx, entityType, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomEntityTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CustomEntityTypeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting custom entity type", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	if err := r.client.DeleteCustomEntityType(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete custom entity type, got error: %s", err))
		return
	}

	// Custom entity type is now deleted and will be removed from Terraform state
}

func (r *CustomEntityTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapCustomEntityTypeToModel converts a CustomEntityType API response into a CustomEntityTypeResourceModel.
func mapCustomEntityTypeToModel(ctx context.Context, entityType *CustomEntityType, data *CustomEntityTypeResourceModel, diags *diag.Diagnostics) {
	data.ID = types.StringValue(entityType.ID)

	if len(entityType.Name) > 0 {
		nameMapValue, d := types.MapValueFrom(ctx, types.StringType, entityType.Name)
		diags.Append(d...)
		data.Name = nameMapValue
	} else {
		nameMapValue, d := types.MapValueFrom(ctx, types.StringType, map[string]string{})
		diags.Append(d...)
		data.Name = nameMapValue
	}

	if entityType.Metadata != nil {
		data.Version = types.Int64Value(int64(entityType.Metadata.Version))

		data.CreatedAt = types.StringNull()
		if entityType.Metadata.CreatedAt != "" {
			data.CreatedAt = types.StringValue(entityType.Metadata.CreatedAt)
		}

		data.ModifiedAt = types.StringNull()
		if entityType.Metadata.ModifiedAt != "" {
			data.ModifiedAt = types.StringValue(entityType.Metadata.ModifiedAt)
		}
	} else {
		data.Version = types.Int64Null()
		data.CreatedAt = types.StringNull()
		data.ModifiedAt = types.StringNull()
	}
}
