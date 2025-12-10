package provider

import (
	"context"
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

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PaymentModeResource{}
var _ resource.ResourceWithImportState = &PaymentModeResource{}

func NewPaymentModeResource() resource.Resource {
	return &PaymentModeResource{}
}

// PaymentModeResource defines the resource implementation.
type PaymentModeResource struct {
	client *EmporixClient
}

// PaymentModeResourceModel describes the resource data model.
type PaymentModeResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Code            types.String `tfsdk:"code"`
	Active          types.Bool   `tfsdk:"active"`
	PaymentProvider types.String `tfsdk:"payment_provider"`
	Configuration   types.Map    `tfsdk:"configuration"`
}

func (r *PaymentModeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_paymentmode"
}

func (r *PaymentModeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Payment mode resource for configuring payment methods in Emporix",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the payment mode (UUID)",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"code": schema.StringAttribute{
				MarkdownDescription: "Code of the payment mode (unique identifier)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether the payment mode is active. Defaults to true.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"payment_provider": schema.StringAttribute{
				MarkdownDescription: "Payment provider type. Valid values: INVOICE, CASH_ON_DELIVERY, SPREEDLY, SPREEDLY_SAFERPAY, UNZER",
				Required:            true,
			},
			"configuration": schema.MapAttribute{
				MarkdownDescription: "Map of configuration values for the payment gateway. Required keys depend on the provider type.",
				ElementType:         types.StringType,
				Optional:            true,
			},
		},
	}
}

func (r *PaymentModeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PaymentModeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PaymentModeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	paymentMode := &PaymentMode{
		Code:     plan.Code.ValueString(),
		Active:   plan.Active.ValueBool(),
		Provider: plan.PaymentProvider.ValueString(),
	}

	// Handle configuration map
	if !plan.Configuration.IsNull() && !plan.Configuration.IsUnknown() {
		configMap := make(map[string]string)
		diags := plan.Configuration.ElementsAs(ctx, &configMap, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		paymentMode.Configuration = configMap
	}

	tflog.Debug(ctx, "Creating payment mode", map[string]interface{}{
		"code": paymentMode.Code,
	})

	// Create payment mode
	createdMode, err := r.client.CreatePaymentMode(ctx, paymentMode)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create payment mode, got error: %s", err))
		return
	}

	// Update plan with created values
	plan.ID = types.StringValue(createdMode.ID)
	plan.Code = types.StringValue(createdMode.Code)
	plan.Active = types.BoolValue(createdMode.Active)
	plan.PaymentProvider = types.StringValue(createdMode.Provider)

	if len(createdMode.Configuration) > 0 {
		configMap, diags := types.MapValueFrom(ctx, types.StringType, createdMode.Configuration)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Configuration = configMap
	} else {
		plan.Configuration = types.MapNull(types.StringType)
	}

	tflog.Debug(ctx, "Payment mode created", map[string]interface{}{
		"id":   createdMode.ID,
		"code": createdMode.Code,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PaymentModeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PaymentModeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading payment mode", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	// Get payment mode from API
	paymentMode, err := r.client.GetPaymentMode(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read payment mode, got error: %s", err))
		return
	}

	if paymentMode == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state
	state.ID = types.StringValue(paymentMode.ID)
	state.Code = types.StringValue(paymentMode.Code)
	state.Active = types.BoolValue(paymentMode.Active)
	state.PaymentProvider = types.StringValue(paymentMode.Provider)

	if len(paymentMode.Configuration) > 0 {
		configMap, diags := types.MapValueFrom(ctx, types.StringType, paymentMode.Configuration)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Configuration = configMap
	} else {
		state.Configuration = types.MapNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PaymentModeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PaymentModeResourceModel
	var state PaymentModeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating payment mode", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	// Build update request
	updateData := &PaymentModeUpdate{
		Active: plan.Active.ValueBool(),
	}

	// Handle configuration map
	if !plan.Configuration.IsNull() && !plan.Configuration.IsUnknown() {
		configMap := make(map[string]string)
		diags := plan.Configuration.ElementsAs(ctx, &configMap, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateData.Configuration = configMap
	}

	// Update payment mode
	updatedMode, err := r.client.UpdatePaymentMode(ctx, state.ID.ValueString(), updateData)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update payment mode, got error: %s", err))
		return
	}

	// Update state
	plan.ID = types.StringValue(updatedMode.ID)
	plan.Code = types.StringValue(updatedMode.Code)
	plan.Active = types.BoolValue(updatedMode.Active)
	plan.PaymentProvider = types.StringValue(updatedMode.Provider)

	if len(updatedMode.Configuration) > 0 {
		configMap, diags := types.MapValueFrom(ctx, types.StringType, updatedMode.Configuration)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Configuration = configMap
	} else {
		plan.Configuration = types.MapNull(types.StringType)
	}

	tflog.Debug(ctx, "Payment mode updated", map[string]interface{}{
		"id": updatedMode.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PaymentModeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PaymentModeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting payment mode", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	err := r.client.DeletePaymentMode(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete payment mode, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "Payment mode deleted", map[string]interface{}{
		"id": state.ID.ValueString(),
	})
}

func (r *PaymentModeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
