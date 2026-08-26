# Webhook Subscription Examples
# API allows only 1 configuration of given type.
# So you can have 1 http, 1 SVIX and 1 shared SVIX config.
# To test, uncomment desiresd setup.

terraform {
  required_providers {
    emporix = {
      source  = "emporix/emporix"
      version = "~> 0.1"
    }
  }
}

# Configure the Emporix provider
provider "emporix" {
  tenant  = var.emporix_tenant
  api_url = var.emporix_api_url

  # Use client credentials from your Custom API Key
  client_id     = var.emporix_client_id
  client_secret = var.emporix_client_secret
}

# Variables
variable "emporix_tenant" {
  description = "Emporix tenant name"
  type        = string
  sensitive   = false
}

variable "emporix_api_url" {
  description = "Emporix API base URL"
  type        = string
  default     = "https://api.emporix.io"
}

variable "emporix_client_id" {
  description = "Emporix OAuth2 client ID"
  type        = string
  sensitive   = true
}

variable "emporix_client_secret" {
  description = "Emporix OAuth2 client secret"
  type        = string
  sensitive   = true
}

variable "webhook_destination_url" {
  description = "Destination URL for the HTTP webhook examples below. Must be reachable: the API verifies it by sending a HEAD/OPTIONS request and rejects unreachable URLs."
  type        = string
}

# =============================================================================
# Example 1: SVIX Provider (Your Own Svix Server)
# =============================================================================
# Using your own Svix server instance. Requires destination URL to your Svix app.

resource "emporix_webhook" "svix_webhook" {
  code          = "mySvixWebhook"
  provider_type = "SVIX"
  active        = false

  # Svix application secret key for signing
  secret_key = "<secret>"
}

# # =============================================================================
# # Example 2: Webhook with Event-Specific Configuration
# # =============================================================================
# # Define different destinations and settings for different event types.
# #
# # This example drives events_configuration from a variable and varies per-event headers only.
# # If you also need per-event overrides (destination_url, secret_key, subscribed, filter,
# # excluded_fields, active, name), include those attributes (as optional) in the variable's
# # object type as well.

# variable "multi_event_webhook_events_configuration" {
#   description = "Event-specific configuration for multi_event_webhook, supplied via a variable."
#   type = list(object({
#     event_type = string
#     headers    = map(string)
#   }))
#   default = [
#     {
#       event_type = "order.created"
#       headers = {
#         X-Event-Group = "orders"
#       }
#     },
#     {
#       event_type = "customer.created"
#       headers = {
#         X-Event-Group = "customers"
#       }
#     },
#     {
#       event_type = "customer.updated"
#       headers = {
#         X-Event-Group = "customers"
#       }
#     },
#     {
#       event_type = "product.updated"
#       headers    = {}
#     }
#   ]
# }

# resource "emporix_webhook" "multi_event_webhook" {
#   code            = "multiEventWebhook"
#   provider_type   = "HTTP"
#   destination_url = var.webhook_destination_url
#   active          = true

#   secret_key = "default-secret-key"

#   events_configuration = var.multi_event_webhook_events_configuration
# }

# =============================================================================
# Example 3: Multiple Targets for the Same Event Type
# =============================================================================
# events_configuration may contain more than one entry for the same event_type - each is
# routed to its own destination independently, and each may carry its own JsonPath filter
# to only forward matching events.

resource "emporix_webhook" "multi_target_webhook" {
  code            = "multiTargetWebhook"
  provider_type   = "HTTP"
  destination_url = var.webhook_destination_url
  active          = true

  # Both entries reuse webhook_destination_url with distinct query strings so this example
  # is directly appliable: the API verifies every destination_url (including per-entry
  # overrides) is reachable via HEAD/OPTIONS, so a placeholder domain like
  # fulfillment.example.com will be rejected. Point webhook_destination_url at a real
  # endpoint (e.g. a webhook.site URL) and swap these for your actual targets.
  # filter below is illustrative - adjust the field path to match your actual
  # product.created payload shape.
  events_configuration = [
    {
      event_type      = "product.created"
      name            = "products -> catalog sync"
      destination_url = "${var.webhook_destination_url}?target=catalog-sync"
    },
    {
      event_type      = "product.created"
      name            = "premium products -> merchandising review"
      destination_url = "${var.webhook_destination_url}?target=merchandising-review"
      filter          = "$[?(@.code == 'PREMIUM-001')]"
      excluded_fields = ["internalNotes"]
    }
  ]
}

# # =============================================================================
# # Outputs - Display current webhook configurations after terraform apply
# # =============================================================================

# output "webhook_configurations" {
#   description = "List of all configured webhooks with their key settings"
#   value = [
#     {
#       name            = "svix_webhook"
#       code            = emporix_webhook.svix_webhook.code
#       active          = emporix_webhook.svix_webhook.active
#       provider_type   = emporix_webhook.svix_webhook.provider_type
#       destination_url = emporix_webhook.svix_webhook.destination_url
#       version         = emporix_webhook.svix_webhook.version
#     },
#     {
#       name            = "multi_event_webhook"
#       code            = emporix_webhook.multi_event_webhook.code
#       active          = emporix_webhook.multi_event_webhook.active
#       provider_type   = emporix_webhook.multi_event_webhook.provider_type
#       destination_url = emporix_webhook.multi_event_webhook.destination_url
#       version         = emporix_webhook.multi_event_webhook.version
#       headers         = emporix_webhook.multi_event_webhook.headers
#     }
#   ]
# }

# output "webhook_summary" {
#   description = "Summary of webhook configurations (code, active status, provider type)"
#   value = {
#     "svix_webhook" = {
#       code          = emporix_webhook.svix_webhook.code
#       active        = emporix_webhook.svix_webhook.active
#       provider_type = emporix_webhook.svix_webhook.provider_type
#     }
#     "multi_event_webhook" = {
#       code          = emporix_webhook.multi_event_webhook.code
#       active        = emporix_webhook.multi_event_webhook.active
#       provider_type = emporix_webhook.multi_event_webhook.provider_type
#     }
#   }
# }
