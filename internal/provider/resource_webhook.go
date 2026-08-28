package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &WebhookResource{}
var _ resource.ResourceWithImportState = &WebhookResource{}
var _ resource.ResourceWithValidateConfig = &WebhookResource{}
var _ resource.ResourceWithModifyPlan = &WebhookResource{}

func NewWebhookResource() resource.Resource {
	return &WebhookResource{}
}

type WebhookResource struct {
	client *EmporixClient
}

type EventConfigModel struct {
	Id             types.String            `tfsdk:"id"`
	EventType      types.String            `tfsdk:"event_type"`
	DestinationUrl types.String            `tfsdk:"destination_url"`
	SecretKey      types.String            `tfsdk:"secret_key"`
	Headers        map[string]types.String `tfsdk:"headers"`
	Filter         types.String            `tfsdk:"filter"`
	ExcludedFields []types.String          `tfsdk:"excluded_fields"`
	Active         types.Bool              `tfsdk:"active"`
	Name           types.String            `tfsdk:"name"`
	Subscribed     types.Bool              `tfsdk:"subscribed"`
}

type WebhookResourceModel struct {
	Code                types.String            `tfsdk:"code"`
	Active              types.Bool              `tfsdk:"active"`
	Provider            types.String            `tfsdk:"provider_type"`
	DestinationUrl      types.String            `tfsdk:"destination_url"`
	SecretKey           types.Bool              `tfsdk:"secret_key_exists"`
	SecretKeyString     types.String            `tfsdk:"secret_key"`
	Headers             map[string]types.String `tfsdk:"headers"`
	EventsConfiguration []EventConfigModel      `tfsdk:"events_configuration"`
	Version             types.Int64             `tfsdk:"version"`
}

type eventDestinationUrlDefaultModifier struct{}

func (m eventDestinationUrlDefaultModifier) Description(ctx context.Context) string {
	return "Defaults to the resource-level destination_url when this event does not set its own."
}

func (m eventDestinationUrlDefaultModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m eventDestinationUrlDefaultModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	if !req.ConfigValue.IsNull() && req.ConfigValue.ValueString() != "" {
		return
	}

	var parentDestinationUrl types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("destination_url"), &parentDestinationUrl)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.PlanValue = parentDestinationUrl
}

func (r *WebhookResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *WebhookResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a webhook subscription configuration in Emporix. " +
			"Webhooks support three providers: SVIX_SHARED (default Emporix Svix server), " +
			"SVIX (your own Svix server), and HTTP (direct HTTP POST). " +
			"Each webhook configuration defines where events should be sent and how they are authenticated.",

		Attributes: map[string]schema.Attribute{
			"code": schema.StringAttribute{
				MarkdownDescription: "Webhook code (unique identifier for this configuration). Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether this webhook configuration is active. Only one configuration per tenant can be active at a time. Defaults to false.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"provider_type": schema.StringAttribute{
				MarkdownDescription: "Webhook provider type. Accepted values are case-insensitive and dashes are converted to underscores for API requests. Valid values: `SVIX_SHARED`, `SVIX`, `HTTP`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive("SVIX_SHARED", "SVIX", "HTTP"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destination_url": schema.StringAttribute{
				MarkdownDescription: "The URL where webhook events will be sent.",
				Optional:            true,
			},
			"secret_key": schema.StringAttribute{
				MarkdownDescription: "Secret key for HMAC message signing when provider is 'HTTP' (sent as `secretKey`). For 'SVIX'/'SVIX_SHARED' provider, this is the Svix application API key (sent as `apiKey`). Omitted from state for 'SVIX_SHARED' provider.",
				Optional:            true,
				Sensitive:           true,
			},
			"secret_key_exists": schema.BoolAttribute{
				MarkdownDescription: "Whether a secret key exists for this webhook (read-only, computed by API). Useful for Svix provider to know if signing is configured.",
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"headers": schema.MapAttribute{
				MarkdownDescription: "HTTP headers to include in webhook requests. Keys and values are strings.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Webhook configuration version (managed by API for optimistic concurrency).",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"events_configuration": schema.ListNestedAttribute{
				MarkdownDescription: "Event-specific configuration. Allows different handling for different event types.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Stable server-generated identifier of this event configuration entry. Omitted on create (client-supplied ids are rejected). Immutable once assigned.",
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"event_type": schema.StringAttribute{
							MarkdownDescription: "Unique identifier of the event. Multiple entries may share the same `event_type`.",
							Required:            true,
						},
						"destination_url": schema.StringAttribute{
							MarkdownDescription: "Destination URL where the event should be sent. Has higher priority than `destination_url` on the root level - each event can have a separate destination URL. If empty, uses the parent destination_url.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								eventDestinationUrlDefaultModifier{},
							},
						},
						"secret_key": schema.StringAttribute{
							MarkdownDescription: "Secret key used to sign the message for this entry. Has higher priority than `secret_key` on the root level - each event can have a separate secret key. Omitted from state for SVIX_SHARED provider.",
							Optional:            true,
							Sensitive:           true,
						},
						"headers": schema.MapAttribute{
							MarkdownDescription: "Key-value pairs decorating the outgoing HTTP POST request as headers for this entry (size limit 10). Has higher priority than `headers` on the root level - each event can have separate headers.",
							Optional:            true,
							ElementType:         types.StringType,
						},
						"filter": schema.StringAttribute{
							MarkdownDescription: "Optional Jayway JsonPath predicate evaluated against the event payload. When omitted or empty, the entry matches every event of the given event_type. Invalid expressions are rejected by the API.",
							Optional:            true,
						},
						"excluded_fields": schema.ListAttribute{
							MarkdownDescription: "Optional per-entry field exclusion list; only non-blank top-level field names are allowed. Omit or leave null to inherit the event-subscription excludedFields. An empty list overrides the subscription exclusions with no exclusions for this target.",
							Optional:            true,
							ElementType:         types.StringType,
						},
						"active": schema.BoolAttribute{
							MarkdownDescription: "Per-endpoint activation switch. When false, events for this endpoint are dropped without filter evaluation, delivery, or retries; other endpoints are not affected. Distinct from `subscribed` below, which controls the tenant-wide event subscription. Defaults to true.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(true),
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Optional user-facing label for this entry (e.g. \"ERP integration\"). Purely descriptive - it has no impact on delivery. Maximum 255 characters.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(255),
							},
						},
						"subscribed": schema.BoolAttribute{
							MarkdownDescription: "Indicates whether the tenant is actually subscribed to this event type (controls actual message delivery, separately from the URL/headers overrides above). Defaults to true.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(true),
						},
					},
				},
			},
		},
	}
}

func (r *WebhookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WebhookResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userProviderValue := plan.Provider.ValueString()
	apiProviderValue := normalizeProvider(plan.Provider.ValueString())

	// Serializes webhook creates per tenant so the API never sees zero active webhooks.
	mu := getWebhookMutex(r.client.Tenant)
	mu.Lock()
	defer mu.Unlock()

	nestedConfig := buildNestedConfigFromModel(plan, userProviderValue)
	createReq := &webhookCreateRequest{
		Code:          plan.Code.ValueString(),
		Active:        plan.Active.ValueBool(),
		Provider:      apiProviderValue,
		Configuration: nestedConfig,
	}

	// active is known at plan time, so it must be sent as-is; a create that would leave
	// zero active webhooks is rejected by the API and surfaced below, not worked around.
	webhook, err := r.client.CreateWebhook(ctx, createReq)
	if err != nil {
		if !createReq.Active && strings.Contains(err.Error(), "active config has to be present") {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf(
				"Unable to create webhook configuration: the API requires at least one active webhook on the "+
					"tenant, and this would be the first/only one. Set active = true on this resource, or add "+
					"depends_on = [emporix_webhook.<some_active_one>] so an active webhook is guaranteed to exist "+
					"first. Got error: %s", err))
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create webhook configuration, got error: %s", err))
		return
	}

	webhook, err = r.client.GetWebhook(ctx, webhook.Code)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read created webhook state, got error: %s", err))
		return
	}

	result := webhookToModel(webhook)
	result.Provider = types.StringValue(userProviderValue)
	mergeSensitiveValuesIntoResult(&result, &plan)
	mergeEventsFromPlan(&result, &plan)
	if len(result.EventsConfiguration) == 0 && len(plan.EventsConfiguration) == 0 {
		result.EventsConfiguration = nil
	}

	if updates := buildEventSubscriptionUpdates(plan.EventsConfiguration, nil); len(updates) > 0 {
		if err := r.client.UpdateEventSubscriptions(ctx, updates); err != nil {
			resp.Diagnostics.AddWarning("Event subscriptions not fully applied",
				fmt.Sprintf("Webhook '%s' was created, but failed to set event subscriptions: %s. "+
					"The webhook has been saved to state; re-run apply to reconcile subscriptions.",
					plan.Code.ValueString(), err))
		}
	}
	refreshEventSubscriptions(ctx, r.client, &result, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *WebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WebhookResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	webhook, err := r.client.GetWebhook(ctx, state.Code.ValueString())
	if err != nil {
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read webhook configuration, got error: %s", err))
		return
	}

	result := webhookToModel(webhook)
	preserveTopLevelFields(&result, &state)

	if len(result.EventsConfiguration) == 0 && len(state.EventsConfiguration) > 0 {
		result.EventsConfiguration = state.EventsConfiguration
	} else {
		mergeEventsFromState(&result, &state)
	}

	if len(result.EventsConfiguration) == 0 && len(state.EventsConfiguration) == 0 {
		result.EventsConfiguration = nil
	}

	refreshEventSubscriptions(ctx, r.client, &result, &resp.Diagnostics)

	if !state.Provider.IsNull() {
		result.Provider = state.Provider
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *WebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state WebhookResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userProviderValue := plan.Provider.ValueString()

	mu := getWebhookMutex(r.client.Tenant)
	mu.Lock()
	defer mu.Unlock()

	current, err := r.client.GetWebhook(ctx, plan.Code.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read current webhook state, got error: %s", err))
		return
	}

	patches := buildPatchOperations(current, plan, state)

	// Skip API call if no patches are needed (plan and state are identical).
	// The API rejects PATCH requests with an empty body.
	if len(patches) == 0 {
		// Just refresh state from API
		webhook, err := r.client.GetWebhook(ctx, plan.Code.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read webhook state, got error: %s", err))
			return
		}
		result := webhookToModel(webhook)
		preserveTopLevelFields(&result, &state)

		subscriptionUpdates := buildEventSubscriptionUpdates(plan.EventsConfiguration, state.EventsConfiguration)
		if len(subscriptionUpdates) > 0 {
			if err := r.client.UpdateEventSubscriptions(ctx, subscriptionUpdates); err != nil {
				resp.Diagnostics.AddWarning("UpdateEventSubscriptions failed",
					fmt.Sprintf("Unable to update event subscriptions: %s", err))
			}
		}
		refreshEventSubscriptions(ctx, r.client, &result, &resp.Diagnostics)

		result.Provider = types.StringValue(userProviderValue)
		resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		return
	}

	_, err = r.client.UpdateWebhook(ctx, plan.Code.ValueString(), patches)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update webhook configuration, got error: %s", err))
		return
	}

	webhook, err := r.client.GetWebhook(ctx, plan.Code.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated webhook state, got error: %s", err))
		return
	}

	result := webhookToModel(webhook)
	mergeSensitiveValuesIntoResult(&result, &plan)

	if len(result.EventsConfiguration) == 0 && len(plan.EventsConfiguration) > 0 {
		result.EventsConfiguration = plan.EventsConfiguration
	} else {
		mergeEventsFromPlan(&result, &plan)
	}

	if len(result.EventsConfiguration) == 0 && len(plan.EventsConfiguration) == 0 {
		result.EventsConfiguration = nil
	}

	subscriptionUpdates := buildEventSubscriptionUpdates(plan.EventsConfiguration, state.EventsConfiguration)
	if len(subscriptionUpdates) > 0 {
		if err := r.client.UpdateEventSubscriptions(ctx, subscriptionUpdates); err != nil {
			resp.Diagnostics.AddWarning("UpdateEventSubscriptions failed",
				fmt.Sprintf("Unable to update event subscriptions: %s", err))
		}
	}
	refreshEventSubscriptions(ctx, r.client, &result, &resp.Diagnostics)

	result.Provider = types.StringValue(userProviderValue)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *WebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WebhookResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mu := getWebhookMutex(r.client.Tenant)
	mu.Lock()
	defer mu.Unlock()

	err := r.client.DeleteWebhook(ctx, state.Code.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete webhook configuration, got error: %s", err))
		return
	}
}

func (r *WebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("code"), req, resp)
}

// Resolves events_configuration[*].id via correlateEventEntries instead of the schema's
// position-based UseStateForUnknown, which would otherwise lock in a wrong id before
// Update runs and crash with "provider produced inconsistent result after apply".
func (r *WebhookResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// No prior state to correlate against, and a replace gets fresh ids regardless.
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() || len(resp.RequiresReplace) > 0 {
		return
	}

	var plan, state WebhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(plan.EventsConfiguration) == 0 {
		return
	}

	matched := correlateEventEntries(plan.EventsConfiguration, state.EventsConfiguration)
	for i, idx := range matched {
		if idx != -1 {
			plan.EventsConfiguration[i].Id = state.EventsConfiguration[idx].Id
		} else {
			plan.EventsConfiguration[i].Id = types.StringUnknown()
		}
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

func (r *WebhookResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var providerType types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("provider_type"), &providerType)...)
	if resp.Diagnostics.HasError() {
		return
	}

	provider := normalizeProvider(providerType.ValueString())

	var destinationUrl types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("destination_url"), &destinationUrl)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if provider == "SVIX" || provider == "SVIX_SHARED" {
		if !destinationUrl.IsNull() && !destinationUrl.IsUnknown() && destinationUrl.ValueString() != "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("destination_url"),
				"destination_url not used for this provider",
				fmt.Sprintf("provider_type %q does not use destination_url - Svix endpoints are managed on the Svix "+
					"side, not through this resource, and the Emporix API rejects a config of this type that includes "+
					"it. Remove destination_url.", providerType.ValueString()),
			)
		}

		// apiKey is documented as optional but the API rejects a SVIX config without one.
		var secretKey types.String
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("secret_key"), &secretKey)...)
		if resp.Diagnostics.HasError() {
			return
		}
		secretKeySet := !secretKey.IsNull() && !secretKey.IsUnknown() && secretKey.ValueString() != ""

		if provider == "SVIX_SHARED" && secretKeySet {
			resp.Diagnostics.AddAttributeError(
				path.Root("secret_key"),
				"secret_key not used for this provider",
				"provider_type \"SVIX_SHARED\" does not accept secret_key - its configuration is managed entirely "+
					"by Emporix, and the API rejects a config of this type that includes one. Remove secret_key.",
			)
		}
		if provider == "SVIX" && !secretKeySet {
			resp.Diagnostics.AddAttributeError(
				path.Root("secret_key"),
				"Missing secret_key",
				"provider_type \"SVIX\" requires secret_key (sent as the Svix apiKey) - the Emporix API rejects a "+
					"SVIX config without it, even though it's documented as optional. Note the value must be a real "+
					"API key from your Svix account: Emporix authenticates against Svix's API with it, so a "+
					"placeholder string will fail later with a 500 \"Client 'svix': Unauthorized\" rather than a "+
					"clean validation error.",
			)
		}

		var eventsConfiguration types.List
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("events_configuration"), &eventsConfiguration)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if !eventsConfiguration.IsNull() && !eventsConfiguration.IsUnknown() && len(eventsConfiguration.Elements()) > 0 {
			resp.Diagnostics.AddAttributeError(
				path.Root("events_configuration"),
				"events_configuration not used for this provider",
				fmt.Sprintf("provider_type %q has no eventsConfiguration field - it's HTTP-only. Remove events_configuration.", providerType.ValueString()),
			)
		}

		var headers types.Map
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("headers"), &headers)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if !headers.IsNull() && !headers.IsUnknown() && len(headers.Elements()) > 0 {
			resp.Diagnostics.AddAttributeError(
				path.Root("headers"),
				"headers not used for this provider",
				fmt.Sprintf("provider_type %q has no headers field - it's HTTP-only. Remove headers.", providerType.ValueString()),
			)
		}
		return
	}

	if provider != "HTTP" {
		return
	}

	var eventsConfiguration types.List
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("events_configuration"), &eventsConfiguration)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if eventsConfiguration.IsNull() || eventsConfiguration.IsUnknown() {
		return
	}

	var events []EventConfigModel
	resp.Diagnostics.Append(eventsConfiguration.ElementsAs(ctx, &events, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parentSet := destinationUrl.IsUnknown() || (!destinationUrl.IsNull() && destinationUrl.ValueString() != "")

	for i, event := range events {
		eventSet := event.DestinationUrl.IsUnknown() || (!event.DestinationUrl.IsNull() && event.DestinationUrl.ValueString() != "")
		if !eventSet && !parentSet {
			resp.Diagnostics.AddAttributeError(
				path.Root("events_configuration").AtListIndex(i).AtName("destination_url"),
				"Missing destination_url",
				fmt.Sprintf("events_configuration[%d] (event_type %q) has no destination_url, and the resource-level "+
					"destination_url is also not set. The Emporix API requires a destination URL for every event; "+
					"set one on this event or on the parent destination_url.", i, event.EventType.ValueString()),
			)
		}
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func normalizeProvider(input string) string {
	return strings.ToUpper(strings.ReplaceAll(input, "-", "_"))
}

func stringToNull(s string) types.String {
	if s != "" {
		return types.StringValue(s)
	}
	return types.StringNull()
}

// Config fields differ per provider: HTTP sends destinationUrl/secretKey/headers/
// eventsConfiguration, SVIX sends apiKey only, SVIX_SHARED sends nothing.
func buildNestedConfigFromModel(model WebhookResourceModel, providerType string) *NestedConfigCreate {
	normalizedProvider := normalizeProvider(providerType)

	config := &NestedConfigCreate{}

	switch normalizedProvider {
	case "SVIX_SHARED":
		// Config schema is empty (EmptyConfiguration) - nothing to send.
	case "SVIX":
		// SvixConfig only accepts apiKey; no destinationUrl (endpoints live on Svix's side).
		if !model.SecretKeyString.IsNull() {
			config.ApiKey = model.SecretKeyString.ValueString()
		}
	default:
		if !model.DestinationUrl.IsNull() {
			config.DestinationUrl = model.DestinationUrl.ValueString()
		}

		if !model.SecretKeyString.IsNull() {
			config.SecretKey = model.SecretKeyString.ValueString()
		}

		if len(model.Headers) > 0 {
			config.Headers = buildHeaderFieldValueMapFromModel(model.Headers)
		}

		if len(model.EventsConfiguration) > 0 {
			config.EventsConfiguration = buildEventConfigNestedFromModel(model.EventsConfiguration)
		}
	}

	return config
}

func buildHeaderFieldValueMapFromModel(modelMap map[string]types.String) map[string]HeaderFieldValue {
	if len(modelMap) == 0 {
		return nil
	}

	result := make(map[string]HeaderFieldValue, len(modelMap))
	for k, v := range modelMap {
		if !v.IsNull() {
			result[k] = HeaderFieldValue{Value: v.ValueString()}
		}
	}
	return result
}

func buildEventConfigNestedFromModel(models []EventConfigModel) []EventConfig {
	if len(models) == 0 {
		return nil
	}

	events := make([]EventConfig, 0, len(models))
	for _, m := range models {
		events = append(events, buildOneEventConfigFromModel(m))
	}
	return events
}

// Never sets Id: the API assigns it, and updates address it via the PATCH path, not the payload.
func buildOneEventConfigFromModel(m EventConfigModel) EventConfig {
	event := EventConfig{
		EventType: m.EventType.ValueString(),
	}
	if !m.DestinationUrl.IsNull() && !m.DestinationUrl.IsUnknown() {
		event.DestinationUrl = m.DestinationUrl.ValueString()
	}
	if !m.SecretKey.IsNull() && !m.SecretKey.IsUnknown() {
		event.SecretKey = m.SecretKey.ValueString()
	}
	if len(m.Headers) > 0 {
		event.Headers = buildHeaderFieldValueMapFromModel(m.Headers)
	}
	if !m.Filter.IsNull() && !m.Filter.IsUnknown() {
		event.Filter = m.Filter.ValueString()
	}
	if !m.Name.IsNull() && !m.Name.IsUnknown() {
		event.Name = m.Name.ValueString()
	}
	if !m.Active.IsNull() && !m.Active.IsUnknown() {
		active := m.Active.ValueBool()
		event.Active = &active
	}
	event.ExcludedFields = excludedFieldsFromModel(m.ExcludedFields)
	return event
}

// A pointer lets nil mean "omit" and &[]string{} mean "explicitly clear".
func excludedFieldsFromModel(fields []types.String) *[]string {
	if fields == nil {
		return nil
	}
	values := make([]string, 0, len(fields))
	for _, f := range fields {
		if f.IsNull() || f.IsUnknown() {
			continue
		}
		values = append(values, f.ValueString())
	}
	return &values
}

// excludedFieldsToModel is the read-side inverse of excludedFieldsFromModel.
func excludedFieldsToModel(fields *[]string) []types.String {
	if fields == nil {
		return nil
	}
	values := make([]types.String, 0, len(*fields))
	for _, f := range *fields {
		values = append(values, types.StringValue(f))
	}
	return values
}

func webhookToModel(api *WebhookConfigGet) WebhookResourceModel {
	model := WebhookResourceModel{
		Code:     types.StringValue(api.Code),
		Active:   types.BoolValue(api.Active),
		Provider: types.StringValue(normalizeProvider(api.Provider)),
		Version:  types.Int64Value(int64(api.Version)),
	}

	if api.Configuration != nil {
		config := api.Configuration

		model.DestinationUrl = stringToNull(config.DestinationUrl)

		if config.SecretKeyExists != nil {
			model.SecretKey = types.BoolValue(*config.SecretKeyExists)
		}

		model.SecretKeyString = stringToNull(config.SecretKey)

		if len(config.Headers) > 0 {
			model.Headers = make(map[string]types.String, len(config.Headers))
			for k, v := range config.Headers {
				model.Headers[k] = types.StringValue(v.Value)
			}
		}

		if len(config.EventsConfiguration) > 0 {
			model.EventsConfiguration = make([]EventConfigModel, 0, len(config.EventsConfiguration))
			for _, event := range config.EventsConfiguration {
				eventModel := EventConfigModel{
					Id:             stringToNull(event.Id),
					EventType:      types.StringValue(event.EventType),
					DestinationUrl: stringToNull(event.DestinationUrl),
					SecretKey:      stringToNull(event.SecretKey),
					Filter:         stringToNull(event.Filter),
					Name:           stringToNull(event.Name),
					ExcludedFields: excludedFieldsToModel(event.ExcludedFields),
				}
				if event.Active != nil {
					eventModel.Active = types.BoolValue(*event.Active)
				} else {
					eventModel.Active = types.BoolValue(true)
				}
				if len(event.Headers) > 0 {
					eventModel.Headers = make(map[string]types.String, len(event.Headers))
					for k, v := range event.Headers {
						eventModel.Headers[k] = types.StringValue(v.Value)
					}
				}
				model.EventsConfiguration = append(model.EventsConfiguration, eventModel)
			}
		}
	}

	return model
}

func buildPatchOperations(current *WebhookConfigGet, plan, state WebhookResourceModel) []WebhookConfigPartialUpdates {
	var patches []WebhookConfigPartialUpdates

	provider := strings.ToUpper(strings.ReplaceAll(plan.Provider.ValueString(), "-", "_"))
	var configPrefix string
	switch provider {
	case "SVIX", "SVIX_SHARED":
		configPrefix = "/configuration/svix"
	default:
		configPrefix = "/configuration/http"
	}

	if !plan.Active.Equal(state.Active) {
		patches = append(patches, WebhookConfigPartialUpdates{
			Op:    "UPSERT",
			Path:  "/active",
			Value: plan.Active.ValueBool(),
		})
	}

	// SVIX/SVIX_SHARED have no destinationUrl/headers/eventsConfiguration; SVIX_SHARED has no apiKey either.
	if provider == "SVIX" || provider == "SVIX_SHARED" {
		if provider == "SVIX" && !plan.SecretKeyString.Equal(state.SecretKeyString) {
			patches = append(patches, WebhookConfigPartialUpdates{
				Op:    "UPSERT",
				Path:  configPrefix + "/apiKey",
				Value: plan.SecretKeyString.ValueString(),
			})
		}
		return patches
	}

	if !plan.DestinationUrl.Equal(state.DestinationUrl) {
		patches = append(patches, WebhookConfigPartialUpdates{
			Op:    "UPSERT",
			Path:  configPrefix + "/destinationUrl",
			Value: plan.DestinationUrl.ValueString(),
		})
	}

	if !plan.SecretKeyString.Equal(state.SecretKeyString) {
		patches = append(patches, WebhookConfigPartialUpdates{
			Op:    "UPSERT",
			Path:  configPrefix + "/secretKey",
			Value: plan.SecretKeyString.ValueString(),
		})
	}

	if !reflect.DeepEqual(plan.Headers, state.Headers) {
		headersPath := configPrefix + "/headers"
		if len(plan.Headers) == 0 {
			patches = append(patches, WebhookConfigPartialUpdates{
				Op:   "REMOVE",
				Path: headersPath,
			})
		} else {
			patches = append(patches, WebhookConfigPartialUpdates{
				Op:    "UPSERT",
				Path:  headersPath,
				Value: buildHeaderFieldValueMapFromModel(plan.Headers),
			})
		}
	}

	patches = append(patches, buildEventsConfigurationEntryPatches(
		configPrefix+"/eventsConfigurationEntry", plan.EventsConfiguration, state.EventsConfiguration)...)

	return patches
}

// contentSignature is an entry's API payload (see buildOneEventConfigFromModel) as a
// string, excluding Id and Subscribed - equal signatures mean equal content.
func contentSignature(m EventConfigModel) string {
	b, _ := json.Marshal(buildOneEventConfigFromModel(m))
	return string(b)
}

// correlateEventEntries maps each plan entry to a state index (-1 if none), matching by
// exact content first (order-independent, so reorders/inserts don't misattribute
// content onto the wrong id) and falling back to event_type + first-seen order for
// whatever's left (i.e. genuinely new or edited entries).
func correlateEventEntries(plan, state []EventConfigModel) []int {
	matched := make([]int, len(plan))
	for i := range matched {
		matched[i] = -1
	}
	usedState := make(map[int]bool, len(state))

	contentGroups := make(map[string][]int, len(state))
	for i, s := range state {
		key := contentSignature(s)
		contentGroups[key] = append(contentGroups[key], i)
	}
	for i, p := range plan {
		key := contentSignature(p)
		if queue := contentGroups[key]; len(queue) > 0 {
			matched[i] = queue[0]
			contentGroups[key] = queue[1:]
			usedState[queue[0]] = true
		}
	}

	typeGroups := make(map[string][]int, len(state))
	for i, s := range state {
		if !usedState[i] {
			typeGroups[s.EventType.ValueString()] = append(typeGroups[s.EventType.ValueString()], i)
		}
	}
	for i, p := range plan {
		if matched[i] != -1 {
			continue
		}
		key := p.EventType.ValueString()
		if queue := typeGroups[key]; len(queue) > 0 {
			matched[i] = queue[0]
			typeGroups[key] = queue[1:]
			usedState[queue[0]] = true
		}
	}

	return matched
}

// Correlates plan to state via correlateEventEntries, then emits per-entry PATCH ops.
func buildEventsConfigurationEntryPatches(entryPath string, plan, state []EventConfigModel) []WebhookConfigPartialUpdates {
	var patches []WebhookConfigPartialUpdates

	matched := correlateEventEntries(plan, state)
	matchedState := make(map[int]bool, len(state))
	for _, idx := range matched {
		if idx != -1 {
			matchedState[idx] = true
		}
	}

	for i, p := range plan {
		idx := matched[i]
		if idx == -1 {
			// No matching state entry: a genuinely new entry, not a renamed/moved one.
			patches = append(patches, WebhookConfigPartialUpdates{
				Op:    "UPSERT",
				Path:  entryPath,
				Value: buildOneEventConfigFromModel(p),
			})
			continue
		}
		id := state[idx].Id.ValueString()
		if id == "" {
			// No id to address (e.g. state not yet refreshed after an upgrade) - skip
			// rather than emit a malformed path; the next Read repopulates the real id.
			continue
		}
		if !eventEntryContentEqual(p, state[idx]) {
			patches = append(patches, WebhookConfigPartialUpdates{
				Op:    "UPSERT",
				Path:  entryPath + "/" + id,
				Value: buildOneEventConfigFromModel(p),
			})
		}
	}

	for i, s := range state {
		if matchedState[i] {
			continue
		}
		if id := s.Id.ValueString(); id != "" {
			patches = append(patches, WebhookConfigPartialUpdates{
				Op:   "REMOVE",
				Path: entryPath + "/" + id,
			})
		}
	}

	return patches
}

// Ignores Id (identity, not content) and Subscribed (drives a separate API).
func eventEntryContentEqual(a, b EventConfigModel) bool {
	return reflect.DeepEqual(buildOneEventConfigFromModel(a), buildOneEventConfigFromModel(b))
}

func preserveTopLevelFields(result, state *WebhookResourceModel) {
	if result.DestinationUrl.IsNull() && !state.DestinationUrl.IsNull() {
		result.DestinationUrl = state.DestinationUrl
	}
	if result.SecretKeyString.IsNull() && !state.SecretKeyString.IsNull() {
		result.SecretKeyString = state.SecretKeyString
	}
	if result.Headers == nil && state.Headers != nil {
		result.Headers = state.Headers
	}
}

func mergeSensitiveValuesIntoResult(result, plan *WebhookResourceModel) {
	if result.DestinationUrl.IsNull() && !plan.DestinationUrl.IsNull() {
		result.DestinationUrl = plan.DestinationUrl
	}
	if result.SecretKeyString.IsNull() && !plan.SecretKeyString.IsNull() {
		result.SecretKeyString = plan.SecretKeyString
	}
	if result.Headers == nil && plan.Headers != nil {
		result.Headers = plan.Headers
	}
}

func mergeEventsFromState(result *WebhookResourceModel, state *WebhookResourceModel) {
	mergeEventsFromSource(&result.EventsConfiguration, state.EventsConfiguration)
	reorderEventsToMatch(&result.EventsConfiguration, state.EventsConfiguration)
}

func mergeEventsFromPlan(result *WebhookResourceModel, plan *WebhookResourceModel) {
	mergeEventsFromSource(&result.EventsConfiguration, plan.EventsConfiguration)
	reorderEventsToMatch(&result.EventsConfiguration, plan.EventsConfiguration)
}

// Backfills fields the API omitted (e.g. secrets) from source into result.
func mergeEventsFromSource(result *[]EventConfigModel, source []EventConfigModel) {
	matched := correlateEventEntries(source, *result)

	for i, srcEvent := range source {
		idx := matched[i]
		if idx == -1 {
			continue
		}

		resEvent := &(*result)[idx]
		if resEvent.SecretKey.IsNull() && !srcEvent.SecretKey.IsNull() {
			resEvent.SecretKey = srcEvent.SecretKey
		}
		// headers is Optional (not Computed): an explicit {} in config must survive, not
		// just a non-empty map, or a nil map turns into an inconsistent-result error.
		if resEvent.Headers == nil && srcEvent.Headers != nil {
			resEvent.Headers = srcEvent.Headers
		}
		if resEvent.DestinationUrl.IsNull() && !srcEvent.DestinationUrl.IsNull() {
			resEvent.DestinationUrl = srcEvent.DestinationUrl
		}
		if resEvent.Subscribed.IsNull() && !srcEvent.Subscribed.IsNull() {
			resEvent.Subscribed = srcEvent.Subscribed
		}
	}
}

// Restores reference's order onto result (an API read, whose order the server controls).
func reorderEventsToMatch(result *[]EventConfigModel, reference []EventConfigModel) {
	if len(reference) == 0 || len(*result) == 0 {
		return
	}

	matched := correlateEventEntries(reference, *result)
	usedResultIndex := make(map[int]struct{}, len(*result))
	reordered := make([]EventConfigModel, 0, len(*result))

	for _, idx := range matched {
		if idx == -1 {
			continue
		}
		if _, already := usedResultIndex[idx]; already {
			continue
		}
		reordered = append(reordered, (*result)[idx])
		usedResultIndex[idx] = struct{}{}
	}

	// Append any remaining API-returned entries not already placed, to avoid dropping data.
	for i, e := range *result {
		if _, ok := usedResultIndex[i]; !ok {
			reordered = append(reordered, e)
		}
	}

	*result = reordered
}

func refreshEventSubscriptions(ctx context.Context, client *EmporixClient, result *WebhookResourceModel, diags *diag.Diagnostics) {
	if len(result.EventsConfiguration) == 0 {
		return
	}
	entries, err := client.ListEventSubscriptions(ctx)
	if err != nil {
		diags.AddWarning("ListEventSubscriptions failed",
			fmt.Sprintf("Failed to refresh subscription status: %s", err))
		return
	}
	applyCurrentSubscriptions(result.EventsConfiguration, subscriptionStatusMap(entries))
}

func subscriptionStatusMap(entries []WebhookEventSubscriptionEntry) map[string]string {
	m := make(map[string]string, len(entries))
	for _, e := range entries {
		m[e.Event.Type] = e.Subscription
	}
	return m
}

func applyCurrentSubscriptions(events []EventConfigModel, statuses map[string]string) {
	for i := range events {
		status, ok := statuses[events[i].EventType.ValueString()]
		events[i].Subscribed = types.BoolValue(ok && status == "SUBSCRIBED")
	}
}

func buildEventSubscriptionUpdates(plan, state []EventConfigModel) []WebhookEventSubscriptionUpdate {
	stateByType := make(map[string]EventConfigModel, len(state))
	for _, e := range state {
		stateByType[e.EventType.ValueString()] = e
	}
	planByType := make(map[string]struct{}, len(plan))

	var updates []WebhookEventSubscriptionUpdate
	for _, p := range plan {
		eventType := p.EventType.ValueString()
		planByType[eventType] = struct{}{}
		prev, existed := stateByType[eventType]
		hadSubscribed := existed && (prev.Subscribed.IsNull() || prev.Subscribed.ValueBool())

		var wantSubscribed bool
		switch {
		case p.Subscribed.IsUnknown():
			if existed {
				wantSubscribed = hadSubscribed
			} else {
				wantSubscribed = true
			}
		case p.Subscribed.IsNull():
			wantSubscribed = true
		default:
			wantSubscribed = p.Subscribed.ValueBool()
		}

		if !existed || wantSubscribed != hadSubscribed {
			action := "UNSUBSCRIBE"
			if wantSubscribed {
				action = "SUBSCRIBE"
			}
			updates = append(updates, WebhookEventSubscriptionUpdate{EventType: eventType, Action: action})
		}
	}

	for _, s := range state {
		eventType := s.EventType.ValueString()
		if _, stillPlanned := planByType[eventType]; stillPlanned {
			continue
		}
		hadSubscribed := s.Subscribed.IsNull() || s.Subscribed.ValueBool()
		if hadSubscribed {
			updates = append(updates, WebhookEventSubscriptionUpdate{EventType: eventType, Action: "UNSUBSCRIBE"})
		}
	}

	return updates
}
