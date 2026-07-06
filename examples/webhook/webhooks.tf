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
# Example 1: SVIX Provider (Your Own Svix Server)
# =============================================================================
# Using your own Svix server instance. Requires destination URL to your Svix app.

resource "emporix_webhook" "svix_webhook" {
  code          = "mySvixWebhook"
  provider_type = "svix"
  active        = false

  # Svix application secret key for signing
  secret_key = "<secret>"
}

# =============================================================================
# Example 2: Webhook with Event-Specific Configuration
# =============================================================================
# Define different destinations and settings for different event types.

resource "emporix_webhook" "multi_event_webhook" {
  code          = "multiEventWebhook"
  provider_type = "http"
  destination_url = "<URL>"
  active        = true

  secret_key = "default-secret-key"

  # Event-specific overrides
  events_configuration = [
    {
      event_type      = "order.created"
      destination_url = "<URL>"
      secret_key      = "orders-secret-key"
      headers = {
        X-Event-Group = "orders"
      }
    },
    {
      event_type = "customer.created"
      secret_key = "customers-secret-key"
      destination_url = "<URL>"
      headers = {
        X-Event-Group = "customers"
      }
    },
    {
      event_type      = "product.updated"
      destination_url = "<URL>"
    }
  ]
}

# =============================================================================
# Outputs - Display current webhook configurations after terraform apply
# =============================================================================

output "webhook_configurations" {
  description = "List of all configured webhooks with their key settings"
  value = [
    {
      name            = "svix_webhook"
      code            = emporix_webhook.svix_webhook.code
      active          = emporix_webhook.svix_webhook.active
      provider_type   = emporix_webhook.svix_webhook.provider_type
      destination_url = emporix_webhook.svix_webhook.destination_url
      version         = emporix_webhook.svix_webhook.version
    },
    {
      name            = "multi_event_webhook"
      code            = emporix_webhook.multi_event_webhook.code
      active          = emporix_webhook.multi_event_webhook.active
      provider_type   = emporix_webhook.multi_event_webhook.provider_type
      destination_url = emporix_webhook.multi_event_webhook.destination_url
      version         = emporix_webhook.multi_event_webhook.version
      headers         = emporix_webhook.multi_event_webhook.headers
    }
  ]
}

output "webhook_summary" {
  description = "Summary of webhook configurations (code, active status, provider type)"
  value = {
    "svix_webhook" = {
      code          = emporix_webhook.svix_webhook.code
      active        = emporix_webhook.svix_webhook.active
      provider_type = emporix_webhook.svix_webhook.provider_type
    }
    "multi_event_webhook" = {
      code          = emporix_webhook.multi_event_webhook.code
      active        = emporix_webhook.multi_event_webhook.active
      provider_type = emporix_webhook.multi_event_webhook.provider_type
    }
  }
}
