# Webhook Subscription Examples
# Only one config per provider_type is allowed per tenant.

terraform {
  required_providers {
    emporix = {
      source  = "emporix/emporix"
      version = "~> 0.1"
    }
  }
}

provider "emporix" {
  tenant  = var.emporix_tenant
  api_url = var.emporix_api_url

  client_id     = var.emporix_client_id
  client_secret = var.emporix_client_secret
}

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

# Example 1: SVIX Provider (Your Own Svix Server)
resource "emporix_webhook" "svix_webhook" {
  code          = "mySvixWebhook"
  provider_type = "SVIX"
  active        = false

  secret_key = "<secret>"

  # Ensures svix_shared_webhook (active = true) is created first, since the API rejects
  # creating an inactive webhook while the tenant has none active yet.
  depends_on = [emporix_webhook.svix_shared_webhook]
}

# Example 2: SVIX_SHARED Provider (Emporix's Managed Svix) - accepts no configuration
resource "emporix_webhook" "svix_shared_webhook" {
  code          = "svixSharedWebhook"
  provider_type = "SVIX_SHARED"
  active        = true
}

# Example 3: HTTP Webhook - per-event header overrides plus multi-target routing
variable "http_webhook_events_configuration" {
  description = "Event-specific configuration for http_webhook, supplied via a variable."
  type = list(object({
    event_type      = string
    headers         = optional(map(string), {})
    destination_url = optional(string)
    filter          = optional(string)
    excluded_fields = optional(list(string))
    name            = optional(string)
  }))
  default = [
    {
      event_type = "order.created"
      headers = {
        X-Event-Group = "orders"
      }
    },
    {
      event_type = "customer.created"
      headers = {
        X-Event-Group = "customers"
      }
    },
    {
      event_type = "customer.updated"
      headers = {
        X-Event-Group = "customers"
      }
    },
    {
      event_type = "product.updated"
      headers    = {}
    },
    {
      event_type      = "product.created"
      name            = "products -> catalog sync"
      destination_url = "https://example.com?target=catalog-sync"
    },
    {
      event_type      = "product.created"
      name            = "premium products -> merchandising review"
      destination_url = "https://example.com?target=merchandising-review"
      filter          = "$[?(@.code == 'PREMIUM')]"
      excluded_fields = ["internalNotes"]
    }
  ]
}

resource "emporix_webhook" "http_webhook" {
  code            = "httpWebhook"
  provider_type   = "HTTP"
  destination_url = "https://example.com"
  active          = false

  secret_key = "default-secret-key"

  events_configuration = var.http_webhook_events_configuration

  # See svix_webhook's depends_on comment above.
  depends_on = [emporix_webhook.svix_shared_webhook]
}

# Outputs - Display current webhook configurations after terraform apply

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
      name            = "http_webhook"
      code            = emporix_webhook.http_webhook.code
      active          = emporix_webhook.http_webhook.active
      provider_type   = emporix_webhook.http_webhook.provider_type
      destination_url = emporix_webhook.http_webhook.destination_url
      version         = emporix_webhook.http_webhook.version
      headers         = emporix_webhook.http_webhook.headers
    },
    {
      name          = "svix_shared_webhook"
      code          = emporix_webhook.svix_shared_webhook.code
      active        = emporix_webhook.svix_shared_webhook.active
      provider_type = emporix_webhook.svix_shared_webhook.provider_type
      version       = emporix_webhook.svix_shared_webhook.version
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
    "http_webhook" = {
      code          = emporix_webhook.http_webhook.code
      active        = emporix_webhook.http_webhook.active
      provider_type = emporix_webhook.http_webhook.provider_type
    }
    "svix_shared_webhook" = {
      code          = emporix_webhook.svix_shared_webhook.code
      active        = emporix_webhook.svix_shared_webhook.active
      provider_type = emporix_webhook.svix_shared_webhook.provider_type
    }
  }
}
