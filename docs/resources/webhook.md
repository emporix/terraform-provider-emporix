# emporix_webhook (Resource)

Manages a webhook subscription configuration in Emporix. Webhooks support three providers: `SVIX_SHARED` (default Emporix Svix server), `SVIX` (your own Svix server), and `HTTP` (direct HTTP POST). Each webhook configuration defines where events should be sent and how they are authenticated.

**Important:** The Emporix API requires at least one active webhook configuration per tenant. If you attempt to deactivate the last active webhook, the API will enforce `active = true`.
Also, only one configuration of given type is alloved. When you try to add more than one configuration of gicen type, you will get 409 error from API.

## Example Usage

### HTTP Webhook (Direct POST)

```terraform
resource "emporix_webhook" "order_webhook" {
  code          = "orderWebhook"
  provider_type = "HTTP"
  destination_url = "<URL>"
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
  destination_url = "<URL>"
  active        = true

  secret_key = "default-secret-key"

  events_configuration = [
    {
      event_type      = "order.created"
      destination_url = "https://orders.webhook.site/endpoint"
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
      destination_url = "https://products.webhook.site/endpoint"
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
- `destination_url` (String) The URL where webhook events will be sent. Required for `HTTP` and `SVIX` providers.
- `secret_key` (String, Sensitive) Secret key for HMAC message signing when provider is `HTTP` (sent as `secretKey`). For `SVIX`/`SVIX_SHARED` provider, this is the Svix application API key (sent as `apiKey`). Omitted from state for `SVIX_SHARED` provider.
- `headers` (Map of String) HTTP headers to include in webhook requests. Keys and values are strings.
- `events_configuration` (Block List) Event-specific configuration. Allows different handling for different event types. (see [below for nested schema](#nestedblockfor-events_configuration))

### Read-Only

- `secret_key_exists` (Boolean) Whether a secret key exists for this webhook (read-only, computed by API). Useful for Svix provider to know if signing is configured.
- `version` (Number) Webhook configuration version (managed by API for optimistic concurrency).

<a id="nestedblockfor-events_configuration"></a>
### Nested Schema for `events_configuration`

Required:

- `event_type` (String) The Emporix event type (e.g., `order.created`, `customer.registered`).

Optional:

- `destination_url` (String) Override destination URL for this specific event type. If empty, uses the parent `destination_url`.
- `secret_key` (String, Sensitive) Override secret key for this specific event type. Omitted from state for `SVIX_SHARED` provider.
- `headers` (Map of String) HTTP headers to include for this specific event type.

## Provider Types

The webhook resource supports three provider types, each with different configuration requirements:

| Provider Type | Description | Required Fields | Secret Key Purpose |
|---|---|---|---|
| `HTTP` | Direct HTTP POST to destination URL | `destination_url` | HMAC signing key for request authentication |
| `SVIX` | Your own Svix server instance | `destination_url` | Svix application API key (`apiKey`) |
| `SVIX_SHARED` | Emporix default Svix server | None | None |

### Provider Type Behavior

- **HTTP Provider**: Sends webhook events directly via HTTP POST to `destination_url`. Supports custom `headers` and `secret_key` for HMAC signing. The `events_configuration` block is fully supported.
- **SVIX Provider**: Routes events through your own Svix server. Requires `destination_url` pointing to your Svix app and `secret_key` as the API key.
- **SVIX_SHARED Provider**: Uses Emporix's managed Svix infrastructure.

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
  destination_url = "https://webhook.site/endpoint"
  active        = true
}

# Import the resource
# $ terraform import emporix_webhook.imported orderWebhook
```

## Key Concepts

### Active Constraint

The Emporix API requires at least one active webhook configuration per tenant. This constraint is enforced during:

1. **Create**: If no other active webhooks exist and you try to create an inactive webhook, `active` is automatically set to `true`.
2. **Update**: If you try to deactivate the last active webhook, the update is blocked and the state is preserved.

### JSON Patch Updates

Updates to webhook configurations use JSON Patch (RFC 6902) operations. The provider constructs patch operations only for fields that have changed, minimizing the risk of race conditions. The API `version` field is managed automatically.

### Sensitive Value Preservation

Secret keys and headers are carefully preserved during read operations:
- If the API doesn't return sensitive values, the provider falls back to the planned or state values.
- The `secret_key` attribute is marked as `Sensitive` in the schema to prevent exposure in logs.

### Event Configuration Merging

Event-specific configurations (`events_configuration`) are merged intelligently:
- Events from the plan are matched to API responses by `event_type`.
- Original event order from the configuration is preserved.
- Sensitive values (secret keys, headers) from the plan are preserved if the API doesn't return them.

## API Reference

- **List Webhooks**: `GET /webhook/{tenant}/config`
- **Create Webhook**: `POST /webhook/{tenant}/config`
- **Get Webhook**: `GET /webhook/{tenant}/config/{code}`
- **Update Webhook**: `PATCH /webhook/{tenant}/config/{code}` (JSON Patch)
- **Delete Webhook**: `DELETE /webhook/{tenant}/config/{code}?force=true`
