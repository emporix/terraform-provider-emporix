# emporix_webhook (Resource)

Manages a webhook subscription configuration in Emporix. Supports `SVIX_SHARED` (default Emporix Svix server), `SVIX` (your own Svix server), and `HTTP` (direct HTTP POST).

**Notes:**
- At least one config must stay active per tenant; the API rejects creating/updating a webhook that would leave zero active.
- Only one config per `provider_type` is allowed per tenant, even if the existing one is `active = false`.
- `HTTP` only: `destination_url` must respond successfully to `HEAD`/`OPTIONS` (a placeholder/unreachable URL fails at `apply`).
- `SVIX`/`SVIX_SHARED` don't use `destination_url` - rejected by the API if set.

## Example Usage

### HTTP Webhook (Direct POST)

```terraform
resource "emporix_webhook" "order_webhook" {
  code          = "orderWebhook"
  provider_type = "HTTP"
  # Must be a real, reachable URL
  destination_url = "<REACHABLE_URL>"
  active        = true

  secret_key = "my-secret-signing-key"

  headers = {
    X-Api-Key     = "api-key-12345"
    X-Request-ID  = "{{uuid}}"
    Custom-Header = "custom-value"
  }
}
```

### SVIX Provider (Your Own Svix Server)

```terraform
resource "emporix_webhook" "svix_webhook" {
  code          = "mySvixWebhook"
  provider_type = "SVIX"
  active        = false

  # Svix application API key for signing
  secret_key = "<secret>"
}
```

### SVIX_SHARED Provider (Emporix Default)

```terraform
resource "emporix_webhook" "svix_shared_webhook" {
  code          = "svixSharedWebhook"
  provider_type = "SVIX_SHARED"
  active        = true
}
```

### Webhook with Event-Specific Configuration

```terraform
resource "emporix_webhook" "multi_event_webhook" {
  code          = "multiEventWebhook"
  provider_type = "HTTP"
  destination_url = "<REACHABLE_URL>"
  active        = true

  secret_key = "default-secret-key"

  events_configuration = [
    {
      event_type      = "order.created"
      destination_url = "<REACHABLE_URL>/orders"
      secret_key      = "orders-secret-key"
      headers = {
        X-Event-Group = "orders"
      }
    },
    {
      event_type = "customer.created"
      secret_key = "customers-secret-key"
      headers = {
        X-Event-Group = "customers"
      }
    },
    {
      event_type      = "product.updated"
      destination_url = "<REACHABLE_URL>/products"
    }
  ]
}
```

### Multiple Targets for the Same Event Type

```terraform
resource "emporix_webhook" "multi_target_webhook" {
  code          = "multiTargetWebhook"
  provider_type = "HTTP"
  destination_url = "<REACHABLE_URL>"
  active        = true

  events_configuration = [
    {
      event_type      = "product.created"
      name            = "products -> catalog sync"
      destination_url = "<REACHABLE_URL>?target=catalog-sync"
    },
    {
      event_type      = "product.created"
      name            = "premium products -> merchandising review"
      destination_url = "<REACHABLE_URL>?target=merchandising-review"
      filter          = "$[?(@.code == 'PREMIUM-001')]"
      excluded_fields = ["internalNotes"]
    }
  ]
}
```

## Schema

### Required

- `code` (String) Webhook code (unique identifier for this configuration). Cannot be changed after creation. Triggers resource replacement if modified.
- `provider_type` (String) Webhook provider type. Accepted values are case-insensitive and dashes are converted to underscores for API requests. Valid values: `SVIX_SHARED`, `SVIX`, `HTTP`. Triggers resource replacement if modified.

### Optional

- `active` (Boolean) Whether this webhook configuration is active. Only one configuration per tenant can be active at a time. The API requires at least one active webhook, so if this is the last active webhook, deactivating it will be prevented. Defaults to `false`.
- `destination_url` (String) Destination URL where event should be sent. `HTTP`-only, and must be reachable. Not used by `SVIX`/`SVIX_SHARED`.
- `secret_key` (String, Sensitive) `HTTP`: Optional secret key which could be used to sign the message (HMAC SHA-256). `SVIX`: API Key for connecting to SVIX - required in practice and must be a real Svix account key (Emporix authenticates against Svix's API with it; a placeholder fails with a 500). Not accepted by `SVIX_SHARED`.
- `headers` (Map of String) HTTP headers to include in webhook requests. Keys and values are strings. `HTTP`-only - rejected by the API for `SVIX`/`SVIX_SHARED`.
- `events_configuration` (Block List) Event-specific configuration. Allows different handling for different event types. `HTTP`-only - rejected by the API for `SVIX`/`SVIX_SHARED`. (see [below for nested schema](#nestedblockfor-events_configuration))

### Read-Only

- `secret_key_exists` (Boolean) Whether a secret key exists for this webhook (read-only, computed by API). Useful for Svix provider to know if signing is configured.
- `version` (Number) Webhook configuration version (managed by API for optimistic concurrency).

<a id="nestedblockfor-events_configuration"></a>
### Nested Schema for `events_configuration`

Required:

- `event_type` (String) Unique identifier of the event. Multiple entries may share the same `event_type`.

Optional:

- `destination_url` (String) Destination URL where the event should be sent. Has higher priority than `destination_url` on the root level - each event can have a separate destination URL. If empty, uses the parent `destination_url`.
- `secret_key` (String, Sensitive) Secret key used to sign the message for this entry. Has higher priority than `secret_key` on the root level - each event can have a separate secret key.
- `headers` (Map of String) Key-value pairs decorating the outgoing HTTP POST request as headers for this entry (size limit `10`). Has higher priority than `headers` on the root level - each event can have separate headers.
- `filter` (String) Optional Jayway JsonPath predicate evaluated against the event payload. When omitted or empty, the entry matches every event of the given `event_type`. Invalid expressions are rejected by the API.
- `excluded_fields` (List of String) Optional per-entry field exclusion list; only non-blank top-level field names are allowed. Omit or leave null to inherit the event-subscription `excludedFields`. An empty list overrides the subscription exclusions with no exclusions for this target.
- `active` (Boolean) Per-endpoint activation switch. When `false`, events for this endpoint are dropped without filter evaluation, delivery, or retries; other endpoints are not affected. Distinct from `subscribed` below, which controls the tenant-wide event subscription. Defaults to `true`.
- `name` (String) Optional user-facing label for this entry (e.g. "ERP integration"). Purely descriptive - it has no impact on delivery. Maximum 255 characters.
- `subscribed` (Boolean) Whether the tenant is actually subscribed to this event type, controlling actual message delivery separately from the URL/headers overrides above. Defaults to `true`. Set to `false` to keep an event's configuration (destination URL, headers, secret key) in place while temporarily disabling delivery, without having to remove the whole `events_configuration` entry.

Read-Only:

- `id` (String) Stable server-generated identifier of this event configuration entry. Omitted on create (client-supplied ids are rejected). Immutable once assigned.

## Provider Types

The webhook resource supports three provider types, each with different configuration requirements:

| Provider Type | Description | Required Fields | Secret Key Purpose |
|---|---|---|---|
| `HTTP` | Direct HTTP POST to destination URL | `destination_url` | HMAC signing key for request authentication (optional) |
| `SVIX` | Your own Svix server instance | `secret_key` | Svix application API key (`apiKey`, **required** - a real Svix account key, not a placeholder) |
| `SVIX_SHARED` | Emporix default Svix server | None | Not accepted - rejected if set |

Svix endpoints (for `SVIX`) are managed on the Svix side, not through this resource.

## Import

Import is supported using the webhook `code` as the identifier:

```bash
# Import by webhook code
terraform import emporix_webhook.order_webhook orderWebhook

# Import with explicit code argument
terraform import emporix_webhook.my_webhook myWebhookCode
```

In Terraform configuration:

```terraform
resource "emporix_webhook" "imported" {
  code          = "orderWebhook"
  provider_type = "HTTP"
  destination_url = "<REACHABLE_URL>" # must match the imported webhook's actual, reachable URL
  active        = true
}

# Import the resource
# $ terraform import emporix_webhook.imported orderWebhook
```

## Key Concepts

### Active Constraint

The Emporix API requires at least one active webhook configuration per tenant.

1. **Create**: the API rejects creating an inactive webhook if it would be the tenant's only one. If you're creating multiple webhooks from scratch and want some inactive, add `depends_on = [emporix_webhook.<some_active_one>]` to the inactive ones so an active webhook is guaranteed to exist first.
2. **Update**: If you try to deactivate the last active webhook, the update is blocked and the state is preserved.
3. **Delete**: `terraform destroy` always deletes with `force=true`, so an active webhook is removed without needing to deactivate it first.

### JSON Patch Updates

Updates to webhook configurations use JSON Patch (RFC 6902) operations. The provider constructs patch operations only for fields that have changed, minimizing the risk of race conditions. The API `version` field is managed automatically.

### Sensitive Value Preservation

Secret keys and headers are carefully preserved during read operations:
- If the API doesn't return sensitive values, the provider falls back to the planned or state values.
- The `secret_key` attribute is marked as `Sensitive` in the schema to prevent exposure in logs.

### Event Subscription Management

When you add events to `events_configuration`, they are automatically subscribed on the Emporix API side:
- Adding an event automatically issues a `SUBSCRIBE` request to the API
- Removing an event from `events_configuration` automatically issues an `UNSUBSCRIBE` request, stopping that event type from being delivered
- Removing the entire `events_configuration` block automatically unsubscribes all previously subscribed event types
- Subscription status is tracked and synchronized with the API during each `terraform plan` and `terraform apply`

This ensures your Terraform configuration always reflects the actual subscription state on the API, preventing drift between declared events and actual event delivery.

The nested `subscribed` attribute exposes this status directly and lets you control it intentionally:
- It defaults to `true`, so any event listed in `events_configuration` is subscribed unless stated otherwise.
- Set `subscribed = false` on an event to unsubscribe it while keeping its `destination_url`, `headers`, and `secret_key` overrides configured in Terraform. This is different from removing the event from `events_configuration` entirely, which discards that configuration.
- The attribute is also `Computed`, so it reflects the real subscription status read back from the API (e.g., if it was changed outside of Terraform), and will show up as drift on the next `plan`/`apply` if it doesn't match your configuration.

### Multi-Target Updates

Changes are sent as per-entry PATCH operations addressed by `id` (`eventsConfigurationEntry`/`eventsConfigurationEntry/{id}`), not a whole-list replace - this is what lets multiple entries share the same `event_type` without one update clobbering another.

Each plan entry is matched to its existing `id` by content first (so inserting, removing, or reordering entries never misattributes one entry's data onto another's id), falling back to `event_type` + position only for entries that are genuinely new or edited. A pure reorder costs zero API calls. The one case this can't resolve automatically is editing *and* reordering two or more entries sharing an `event_type` in the same apply - there's no way to tell which new content belongs to which existing entry.

## API Reference

- **List Webhooks**: `GET /webhook/{tenant}/config`
- **Create Webhook**: `POST /webhook/{tenant}/config`
- **Get Webhook**: `GET /webhook/{tenant}/config/{code}`
- **Update Webhook**: `PATCH /webhook/{tenant}/config/{code}` (JSON Patch)
- **Delete Webhook**: `DELETE /webhook/{tenant}/config/{code}?force=true`
