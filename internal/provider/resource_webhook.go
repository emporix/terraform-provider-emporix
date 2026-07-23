package provider

import (
	"context"
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

func NewWebhookResource() resource.Resource {
	return &WebhookResource{}
}

type WebhookResource struct {
	client *EmporixClient
}

type EventConfigModel struct {
	EventType      types.String            `tfsdk:"event_type"`
	DestinationUrl types.String            `tfsdk:"destination_url"`
	SecretKey      types.String            `tfsdk:"secret_key"`
	Headers        map[string]types.String `tfsdk:"headers"`
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
						"event_type": schema.StringAttribute{
							MarkdownDescription: "The Emporix event type (e.g., 'order.created', 'customer.registered').",
							Required:            true,
						},
						"destination_url": schema.StringAttribute{
							MarkdownDescription: "Override destination URL for this specific event type. If empty, uses the parent destination_url.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								eventDestinationUrlDefaultModifier{},
							},
						},
						"secret_key": schema.StringAttribute{
							MarkdownDescription: "Override secret key for this specific event type. Omitted from state for Svix_SHARED provider.",
							Optional:            true,
							Sensitive:           true,
						},
						"headers": schema.MapAttribute{
							MarkdownDescription: "HTTP headers to include for this specific event type.",
							Optional:            true,
							ElementType:         types.StringType,
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

	// Lock per-tenant mutex to prevent race conditions when creating webhooks.
	// This ensures that when creating multiple webhooks, the operations are serialized
	// so the API never sees a state with zero active webhooks.
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

	webhook, err := r.client.CreateWebhook(ctx, createReq)
	if err != nil {
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

func (r *WebhookResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var providerType types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("provider_type"), &providerType)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if normalizeProvider(providerType.ValueString()) != "HTTP" {
		return
	}

	var destinationUrl types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("destination_url"), &destinationUrl)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var eventsConfiguration types.List
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("events_configuration"), &eventsConfiguration)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if eventsConfiguration.IsUnknown() {
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

// Provider-specific configuration:
// - HTTP/SVIX_SHARED: destinationUrl, secretKey, headers, eventsConfiguration
// - SVIX: apiKey only
func buildNestedConfigFromModel(model WebhookResourceModel, providerType string) *NestedConfigCreate {
	normalizedProvider := normalizeProvider(providerType)

	config := &NestedConfigCreate{}

	switch normalizedProvider {
	case "SVIX_SHARED":
		if !model.SecretKeyString.IsNull() {
			config.ApiKey = model.SecretKeyString.ValueString()
		}
	case "SVIX":
		if !model.DestinationUrl.IsNull() {
			config.DestinationUrl = model.DestinationUrl.ValueString()
		}

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
		event := EventConfig{
			EventType: m.EventType.ValueString(),
		}
		if !m.DestinationUrl.IsNull() {
			event.DestinationUrl = m.DestinationUrl.ValueString()
		}
		if !m.SecretKey.IsNull() {
			event.SecretKey = m.SecretKey.ValueString()
		}
		if len(m.Headers) > 0 {
			event.Headers = buildHeaderFieldValueMapFromModel(m.Headers)
		}
		events = append(events, event)
	}
	return events
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
					EventType:      types.StringValue(event.EventType),
					DestinationUrl: stringToNull(event.DestinationUrl),
					SecretKey:      stringToNull(event.SecretKey),
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

	if !plan.DestinationUrl.Equal(state.DestinationUrl) {
		patches = append(patches, WebhookConfigPartialUpdates{
			Op:    "UPSERT",
			Path:  configPrefix + "/destinationUrl",
			Value: plan.DestinationUrl.ValueString(),
		})
	}

	if !plan.SecretKeyString.Equal(state.SecretKeyString) {
		secretKeyPath := configPrefix + "/secretKey"
		if provider == "SVIX" || provider == "SVIX_SHARED" {
			secretKeyPath = configPrefix + "/apiKey"
		}
		patches = append(patches, WebhookConfigPartialUpdates{
			Op:    "UPSERT",
			Path:  secretKeyPath,
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

	planEvents := buildEventConfigNestedFromModel(plan.EventsConfiguration)
	stateEvents := buildEventConfigNestedFromModel(state.EventsConfiguration)
	if !reflect.DeepEqual(planEvents, stateEvents) {
		eventsPath := configPrefix + "/eventsConfiguration"
		if len(plan.EventsConfiguration) == 0 {
			patches = append(patches, WebhookConfigPartialUpdates{
				Op:   "REMOVE",
				Path: eventsPath,
			})
		} else {
			patches = append(patches, WebhookConfigPartialUpdates{
				Op:    "UPSERT",
				Path:  eventsPath,
				Value: planEvents,
			})
		}
	}

	return patches
}

func preserveTopLevelFields(result, state *WebhookResourceModel) {
	if result.DestinationUrl.IsNull() && !state.DestinationUrl.IsNull() {
		result.DestinationUrl = state.DestinationUrl
	}
	if result.SecretKeyString.IsNull() && !state.SecretKeyString.IsNull() {
		result.SecretKeyString = state.SecretKeyString
	}
	if len(result.Headers) == 0 && len(state.Headers) > 0 {
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
	if len(result.Headers) == 0 && len(plan.Headers) > 0 {
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

func mergeEventsFromSource(result *[]EventConfigModel, source []EventConfigModel) {
	sourceMap := make(map[string]EventConfigModel, len(source))
	for _, srcEvent := range source {
		sourceMap[srcEvent.EventType.ValueString()] = srcEvent
	}

	for i := range *result {
		eventType := (*result)[i].EventType.ValueString()
		if srcEvent, ok := sourceMap[eventType]; ok {
			if (*result)[i].SecretKey.IsNull() && !srcEvent.SecretKey.IsNull() {
				(*result)[i].SecretKey = srcEvent.SecretKey
			}
			if len((*result)[i].Headers) == 0 && len(srcEvent.Headers) > 0 {
				(*result)[i].Headers = srcEvent.Headers
			}
			if (*result)[i].DestinationUrl.IsNull() && !srcEvent.DestinationUrl.IsNull() {
				(*result)[i].DestinationUrl = srcEvent.DestinationUrl
			}
			if (*result)[i].Subscribed.IsNull() && !srcEvent.Subscribed.IsNull() {
				(*result)[i].Subscribed = srcEvent.Subscribed
			}
		}
	}
}

func reorderEventsToMatch(result *[]EventConfigModel, reference []EventConfigModel) {
	if len(reference) == 0 || len(*result) == 0 {
		return
	}
	resultByType := make(map[string]EventConfigModel, len(*result))
	for _, e := range *result {
		resultByType[e.EventType.ValueString()] = e
	}
	reordered := make([]EventConfigModel, 0, len(*result))
	used := make(map[string]struct{}, len(reference))
	for _, refEvent := range reference {
		refType := refEvent.EventType.ValueString()
		if e, ok := resultByType[refType]; ok {
			reordered = append(reordered, e)
			used[refType] = struct{}{}
		}
	}
	// Append any remaining events to avoid dropping API-returned data.
	for _, e := range *result {
		t := e.EventType.ValueString()
		if _, ok := used[t]; !ok {
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
