package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &TenantConfigurationResource{}
var _ resource.ResourceWithImportState = &TenantConfigurationResource{}

func NewTenantConfigurationResource() resource.Resource {
	return &TenantConfigurationResource{}
}

type TenantConfigurationResource struct {
	client *EmporixClient
}

// TenantConfigurationResourceModel describes the resource data model.
type TenantConfigurationResourceModel struct {
	Key     types.String `tfsdk:"key"`
	Value   types.String `tfsdk:"value"`
	Version types.Int64  `tfsdk:"version"`
	Secured types.Bool   `tfsdk:"secured"`
}

func (r *TenantConfigurationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_configuration"
}

func (r *TenantConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a tenant configuration in Emporix. " +
			"Tenant configurations store key-value pairs where values can be any valid JSON. " +
			"The key is immutable and cannot be changed after creation.",

		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				MarkdownDescription: "Configuration key (unique identifier). Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Configuration value as JSON string. Can be any valid JSON: object, string, array, or boolean.",
				Required:            true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Configuration version (managed by API).",
				Computed:            true,
			},
			"secured": schema.BoolAttribute{
				MarkdownDescription: "Flag indicating whether the configuration should be encrypted. Defaults to false.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *TenantConfigurationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TenantConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TenantConfigurationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating tenant configuration", map[string]interface{}{
		"key": data.Key.ValueString(),
	})

	// Parse value JSON string to interface{}
	var valueInterface interface{}
	if err := json.Unmarshal([]byte(data.Value.ValueString()), &valueInterface); err != nil {
		resp.Diagnostics.AddError("Invalid JSON", fmt.Sprintf("Unable to parse value as JSON: %s", err))
		return
	}

	// Create configuration via API
	configCreate := &TenantConfigurationCreate{
		Key:     data.Key.ValueString(),
		Value:   valueInterface,
		Secured: data.Secured.ValueBool(),
	}

	config, err := r.client.CreateTenantConfiguration(ctx, configCreate)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create tenant configuration, got error: %s", err))
		return
	}

	// Convert value back to JSON string for storage
	valueJSON, err := json.Marshal(config.Value)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal value to JSON: %s", err))
		return
	}

	data.Key = types.StringValue(config.Key)
	data.Value = types.StringValue(string(valueJSON))
	data.Version = types.Int64Value(int64(config.Version))
	data.Secured = types.BoolValue(config.Secured)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TenantConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TenantConfigurationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading tenant configuration", map[string]interface{}{
		"key": data.Key.ValueString(),
	})

	// Get configuration from API
	config, err := r.client.GetTenantConfiguration(ctx, data.Key.ValueString())
	if err != nil {
		// If resource not found, remove from state (drift detection)
		if IsNotFound(err) {
			tflog.Warn(ctx, "Tenant configuration not found, removing from state", map[string]interface{}{
				"key": data.Key.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read tenant configuration, got error: %s", err))
		return
	}

	// Convert value to JSON string
	valueJSON, err := json.Marshal(config.Value)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal value to JSON: %s", err))
		return
	}

	data.Key = types.StringValue(config.Key)
	data.Value = types.StringValue(string(valueJSON))
	data.Version = types.Int64Value(int64(config.Version))
	data.Secured = types.BoolValue(config.Secured)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TenantConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TenantConfigurationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating tenant configuration", map[string]interface{}{
		"key": data.Key.ValueString(),
	})

	// Parse value JSON string to interface{}
	var valueInterface interface{}
	if err := json.Unmarshal([]byte(data.Value.ValueString()), &valueInterface); err != nil {
		resp.Diagnostics.AddError("Invalid JSON", fmt.Sprintf("Unable to parse value as JSON: %s", err))
		return
	}

	// Get current version for optimistic locking
	currentConfig, err := r.client.GetTenantConfiguration(ctx, data.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read current configuration state, got error: %s", err))
		return
	}

	// Prepare update payload
	updateData := &TenantConfigurationUpdate{
		Key:     data.Key.ValueString(),
		Value:   valueInterface,
		Version: currentConfig.Version,
		Secured: data.Secured.ValueBool(),
	}

	// Update configuration via API
	config, err := r.client.UpdateTenantConfiguration(ctx, data.Key.ValueString(), updateData)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update tenant configuration, got error: %s", err))
		return
	}

	// Convert value back to JSON string
	valueJSON, err := json.Marshal(config.Value)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal value to JSON: %s", err))
		return
	}

	data.Key = types.StringValue(config.Key)
	data.Value = types.StringValue(string(valueJSON))
	data.Version = types.Int64Value(int64(config.Version))
	data.Secured = types.BoolValue(config.Secured)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TenantConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TenantConfigurationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting tenant configuration", map[string]interface{}{
		"key": data.Key.ValueString(),
	})

	// Delete configuration via API
	err := r.client.DeleteTenantConfiguration(ctx, data.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete tenant configuration, got error: %s", err))
		return
	}

	// Configuration is now deleted and will be removed from Terraform state
}

func (r *TenantConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by configuration key (e.g., "project_country", "taxConfiguration")
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}
