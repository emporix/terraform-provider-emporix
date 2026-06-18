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

# =============================================================================
# Example 1: Simple HTTP Webhook
# =============================================================================
# A basic webhook using direct HTTP POST provider.
# Events are sent to the destination URL with optional secret key for HMAC signing.

# resource "emporix_webhook" "order_created" {
#   code          = "myOrderCreated"
#   provider_type = "http"
#   destination_url = "<URL>"
#   active        = false
#   secret_key = "my-secret-signing-key"
# }

# =============================================================================
# Example 2: SVIX Shared Provider (Default Emporix Svix Server)
# =============================================================================
# Using Emporix's built-in Svix server. No destination URL or secret key needed.

# resource "emporix_webhook" "svix-shared_webhook" {
#   code          = "svixSharedWebhook"
#   provider_type = "svix-shared"
#   active        = true
# }

# =============================================================================
# Example 3: SVIX Provider (Your Own Svix Server)
# =============================================================================
# Using your own Svix server instance. Requires destination URL to your Svix app.

# resource "emporix_webhook" "svix_webhook" {
#   code          = "mySvixWebhook"
#   provider_type = "svix"
#   destination_url = "<URL>"
#   active        = true

#   # Svix application secret key for signing
#   secret_key = "SARFJ353DSTGSd3w3cZXX"
# }

# =============================================================================
# Example 4: HTTP Webhook with Custom Headers
# =============================================================================
# Add custom HTTP headers to webhook requests for authentication or tracing.

# resource "emporix_webhook" "order_webhook_with_headers" {
#   code          = "orderWebhookWithHeaders"
#   provider_type = "http"
#   destination_url = "<URL>"
#   active        = false

#   secret_key = "my-secret-key"

#   headers = {
#     X-Api-Key     = "api-key-12345"
#     X-Request-ID  = "{{uuid}}"
#     Custom-Header = "custom-value"
#   }
# }

# =============================================================================
# Example 5: Webhook with Event-Specific Configuration
# =============================================================================
# Define different destinations and settings for different event types.

# resource "emporix_webhook" "multi_event_webhook" {
#   code          = "multiEventWebhook"
#   provider_type = "http"
#   destination_url = "<URL>"
#   active        = false

#   secret_key = "default-secret-key"

#   # Event-specific overrides
#   events_configuration = [
#     {
#       event_type      = "order.created"
#       destination_url = "<URL>"
#       secret_key      = "orders-secret-key"
#       headers = {
#         X-Event-Group = "orders"
#       }
#     },
#     {
#       event_type = "customer.created"
#       secret_key = "customers-secret-key"
#       destination_url = "<URL>"
#       headers = {
#         X-Event-Group = "customers"
#       }
#     },
#     {
#       event_type      = "product.updated"
#       destination_url = "<URL>"
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
#   destination_url = "<URL>"
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
#
#   code          = each.key
#   provider_type = "http"
#   destination_url = each.value.destination_url
#   active        = each.value.active
#
#   secret_key = "environment-secret-key"
# }

# =============================================================================
# Example 8: Import Existing Webhook
# =============================================================================
# To import an existing webhook configuration, use:
# terraform import emporix_webhook.order_webhook_with_headers orderWebhookWithHeaders

# =============================================================================
# Outputs - Display current webhook configurations after terraform apply
# =============================================================================

# output "webhook_configurations" {
#   description = "List of all configured webhooks with their key settings"
#   value = [
#     {
#       name            = "svix-shared_webhook"
#       code            = emporix_webhook.svix-shared_webhook.code
#       active          = emporix_webhook.svix-shared_webhook.active
#       provider_type   = emporix_webhook.svix-shared_webhook.provider_type
#       destination_url = emporix_webhook.svix-shared_webhook.destination_url
#       version         = emporix_webhook.svix-shared_webhook.version
#     },
#     {
#       name            = "order_webhook_with_headers"
#       code            = emporix_webhook.order_webhook_with_headers.code
#       active          = emporix_webhook.order_webhook_with_headers.active
#       provider_type   = emporix_webhook.order_webhook_with_headers.provider_type
#       destination_url = emporix_webhook.order_webhook_with_headers.destination_url
#       version         = emporix_webhook.order_webhook_with_headers.version
#       headers         = emporix_webhook.order_webhook_with_headers.headers
#     }
#   ]
# }

# output "webhook_summary" {
#   description = "Summary of webhook configurations (code, active status, provider type)"
#   value = {
#     "svix-shared_webhook" = {
#       code          = emporix_webhook.svix-shared_webhook.code
#       active        = emporix_webhook.svix-shared_webhook.active
#       provider_type = emporix_webhook.svix-shared_webhook.provider_type
#     }
#     "order_webhook_with_headers" = {
#       code          = emporix_webhook.order_webhook_with_headers.code
#       active        = emporix_webhook.order_webhook_with_headers.active
#       provider_type = emporix_webhook.order_webhook_with_headers.provider_type
#     }
#   }
# }
