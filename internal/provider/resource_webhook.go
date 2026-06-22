package provider

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &WebhookResource{}
var _ resource.ResourceWithImportState = &WebhookResource{}

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
				MarkdownDescription: "Webhook provider type. Accepted values are case-insensitive and dashes are converted to underscores. Canonical format is UPPERCASE with underscores: 'http' (normalized to 'HTTP'), 'svix' (normalized to 'SVIX'), 'svix_shared' or 'svix-shared' (normalized to 'SVIX_SHARED').",
				Required:            true,
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
			},
			"headers": schema.MapAttribute{
				MarkdownDescription: "HTTP headers to include in webhook requests. Keys and values are strings.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Webhook configuration version (managed by API for optimistic concurrency).",
				Computed:            true,
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

	// Pre-check: List existing webhooks to diagnose 409 conflicts and enforce active constraint
	// This helps identify if the resource exists with a different case or is soft-deleted
	existingWebhooks, listErr := r.client.ListWebhooks(ctx)
	if listErr == nil {
		// API requires at least one active webhook. If plan wants to create an inactive webhook
		// and no other active webhooks exist, force active=true.
		if plan.Active.ValueBool() == false {
			anyActive := false
			for _, wh := range existingWebhooks {
				if wh.Active {
					anyActive = true
					break
				}
			}
			if !anyActive {
				plan.Active = types.BoolValue(true)
			}
		}
	}

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

	if !plan.Active.Equal(state.Active) && plan.Active.ValueBool() == false && state.Active.ValueBool() == true {
		existingWebhooks, listErr := r.client.ListWebhooks(ctx)
		if listErr == nil {
			otherActiveCount := 0
			for _, wh := range existingWebhooks {
				if wh.Active && wh.Code != plan.Code.ValueString() {
					otherActiveCount++
				}
			}
			if otherActiveCount == 0 {
				// Read current state to update version/other computed fields
				current, err := r.client.GetWebhook(ctx, plan.Code.ValueString())
				if err != nil {
					resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read current webhook state, got error: %s", err))
					return
				}
				result := webhookToModel(current)
				preserveTopLevelFields(&result, &state)
				result.Provider = types.StringValue(userProviderValue)
				resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
				return
			}
		}
	}

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
	preserveTopLevelFields(&result, &state)

	if len(result.EventsConfiguration) == 0 && len(plan.EventsConfiguration) > 0 {
		result.EventsConfiguration = plan.EventsConfiguration
	} else {
		mergeEventsFromPlan(&result, &plan)
	}

	if len(result.EventsConfiguration) == 0 && len(plan.EventsConfiguration) == 0 {
		result.EventsConfiguration = nil
	}

	result.Provider = types.StringValue(userProviderValue)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *WebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WebhookResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteWebhook(ctx, state.Code.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete webhook configuration, got error: %s", err))
		return
	}
}

func (r *WebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("code"), req, resp)
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

	if normalizedProvider == "SVIX" || normalizedProvider == "SVIX_SHARED" {
		if !model.SecretKeyString.IsNull() {
			config.ApiKey = model.SecretKeyString.ValueString()
		}
	} else {
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
		patches = append(patches, WebhookConfigPartialUpdates{
			Op:    "UPSERT",
			Path:  configPrefix + "/secretKey",
			Value: plan.SecretKeyString.ValueString(),
		})
	}

	if !reflect.DeepEqual(plan.Headers, state.Headers) {
		patches = append(patches, WebhookConfigPartialUpdates{
			Op:    "UPSERT",
			Path:  configPrefix + "/headers",
			Value: buildHeaderFieldValueMapFromModel(plan.Headers),
		})
	}

	if !reflect.DeepEqual(plan.EventsConfiguration, state.EventsConfiguration) {
		patches = append(patches, WebhookConfigPartialUpdates{
			Op:    "UPSERT",
			Path:  configPrefix + "/eventsConfiguration",
			Value: buildEventConfigNestedFromModel(plan.EventsConfiguration),
		})
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
