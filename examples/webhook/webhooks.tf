# Webhook Subscription Examples

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

# =============================================================================
# Example 1: Simple HTTP Webhook
# =============================================================================
# A basic webhook using direct HTTP POST provider.
# Events are sent to the destination URL with optional secret key for HMAC signing.

# resource "emporix_webhook" "order_created" {
#   code          = "myOrderCreated"
#   provider_type = "http"
#   destination_url = "https://webhook.site/987195e8-a2a2-462e-bb38-d18e153f6888"
#   active        = true
#   secret_key = "my-secret-signing-key"
# }

# =============================================================================
# Example 2: SVIX Shared Provider (Default Emporix Svix Server)
# =============================================================================
# Using Emporix's built-in Svix server. No destination URL or secret key needed.

# resource "emporix_webhook" "svix-shared_webhook" {
#   code     = "svixSharedWebhook"
#   provider_type = "svix-shared"
#   active   = true
# }

# =============================================================================
# Example 3: SVIX Provider (Your Own Svix Server)
# =============================================================================
# Using your own Svix server instance. Requires destination URL to your Svix app.

# resource "emporix_webhook" "svix_webhook" {
#   code          = "mySvixWebhook"
#   provider_type = "svix"
#   destination_url = "https://webhook.site/987195e8-a2a2-462e-bb38-d18e153f6888"
#   active        = true

#   # Svix application secret key for signing
#   secret_key = "SARFJ353DSTGSd3w3cZXX"
# }

# =============================================================================
# Example 4: HTTP Webhook with Custom Headers
# =============================================================================
# Add custom HTTP headers to webhook requests for authentication or tracing.

resource "emporix_webhook" "order_webhook_with_headers" {
  code          = "orderWebhookWithHeaders2"
  provider_type = "http"
  destination_url = "https://webhook.site/987195e8-a2a2-462e-bb38-d18e153f6888"
  active        = true

  secret_key = "my-secret-key"

  headers = {
    X-Api-Key    = "api-key-1234"
    X-Request-ID = "{{uuid}}"
    Custom-Header = "custom-value"
  }
}

# =============================================================================
# Example 5: Webhook with Event-Specific Configuration
# =============================================================================
# Define different destinations and settings for different event types.

# resource "emporix_webhook" "multi_event_webhook" {
#   code          = "multiEventWebhook"
#   provider_type = "http"
#   destination_url = "https://webhook.site/987195e8-a2a2-462e-bb38-d18e153f6888"
#   active        = true

#   secret_key = "default-secret-key"

#   # Event-specific overrides
#   events_configuration = [
#     {
#       event_type      = "order.created"
#       destination_url = "https://webhook.site/987195e8-a2a2-462e-bb38-d18e153f6888"
#       secret_key      = "orders-secret-key"
#       headers = {
#         X-Event-Group = "orders"
#       }
#     },
#     {
#       event_type = "customer.created"
#       secret_key = "customers-secret-key"
#       destination_url = "https://webhook.site/987195e8-a2a2-462e-bb38-d18e153f6888"
#       headers = {
#         X-Event-Group = "customers"
#       }
#     },
#     {
#       event_type      = "product.updated"
#       destination_url = "https://webhook.site/987195e8-a2a2-462e-bb38-d18e153f6888"
#     }
#   ]
# }

# # =============================================================================
# # Example 6: Inactive Webhook (Disabled)
# # =============================================================================
# # Create a webhook configuration without activating it.

# resource "emporix_webhook" "inactive_webhook" {
#   code          = "inactiveWebhook"
#   provider_type = "http"
#   destination_url = "https://webhook.site/987195e8-a2a2-462e-bb38-d18e153f6888"
#   active        = false
# }

# =============================================================================
# Example 7: Multiple Webhooks via for_each
# =============================================================================
# Create multiple similar webhooks using for_each.

locals {
  environment_webhooks = {
    "orders_dev" = {
      destination_url = "https://dev.example.com/webhooks/orders"
      active          = false
    }
    "orders_staging" = {
      destination_url = "https://staging.example.com/webhooks/orders"
      active          = true
    }
    "orders_prod" = {
      destination_url = "https://prod.example.com/webhooks/orders"
      active          = true
    }
  }
}

# resource "emporix_webhook" "environment_webhooks" {
#   for_each = local.environment_webhooks

#   code          = each.key
#   provider_type = "http"
#   destination_url = each.value.destination_url
#   active        = each.value.active

#   secret_key = "environment-secret-key"
# }

# =============================================================================
# Example 8: Import Existing Webhook
# =============================================================================
# To import an existing webhook configuration, use:
# terraform import emporix_webhook.order_webhook_with_headers orderWebhookWithHeaders
