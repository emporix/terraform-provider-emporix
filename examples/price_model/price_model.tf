# Example Terraform configuration for Emporix Price Models

terraform {
  required_providers {
    emporix = {
      source  = "emporix/emporix"
      version = "~> 0.1"
    }
  }
}

# Configure the Emporix provider
# Recommended: Use a Custom API Key with only the required scopes
# See: https://developer.emporix.io/ce/getting-started/developer-portal/manage-apikeys#custom-api-keys
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

# Example 1: Basic price model - flat price regardless of ordered quantity
resource "emporix_price_model" "basic" {
  id           = "standard-pricing"
  includes_tax = true
  default      = true

  name = {
    en = "Standard Pricing"
    de = "Standardpreis"
  }

  tier_definition = {
    tier_type = "BASIC"
    tiers = [
      {
        min_quantity = {
          quantity  = 0
          unit_code = "pc"
        }
      }
    ]
  }

  measurement_unit = {
    quantity  = 1
    unit_code = "pc"
  }
}

# Example 2: Volume pricing - lower unit price the more you buy in total
resource "emporix_price_model" "volume" {
  id           = "volume-pricing"
  includes_tax = true

  name = {
    en = "Volume Pricing"
  }

  description = {
    en = "Lower unit price the more you buy"
  }

  tier_definition = {
    tier_type = "VOLUME"
    tiers = [
      {
        min_quantity = {
          quantity  = 0
          unit_code = "pc"
        }
      },
      {
        min_quantity = {
          quantity  = 10
          unit_code = "pc"
        }
      },
      {
        min_quantity = {
          quantity  = 50
          unit_code = "pc"
        }
      }
    ]
  }

  measurement_unit = {
    quantity  = 1
    unit_code = "pc"
  }
}

# Example 3: Tiered pricing - price calculated per tier range the total quantity falls into
resource "emporix_price_model" "tiered" {
  id           = "tiered-pricing"
  includes_tax = false

  name = {
    en = "Tiered Pricing"
  }

  tier_definition = {
    tier_type = "TIERED"
    tiers = [
      {
        min_quantity = {
          quantity  = 0
          unit_code = "kg"
        }
      },
      {
        min_quantity = {
          quantity  = 100
          unit_code = "kg"
        }
      }
    ]
  }

  measurement_unit = {
    quantity  = 1
    unit_code = "kg"
  }
}

# Outputs
output "basic_price_model_id" {
  description = "ID of the basic price model"
  value       = emporix_price_model.basic.id
}

output "volume_price_model_tiers" {
  description = "Tiers (including API-assigned tier IDs) of the volume price model"
  value       = emporix_price_model.volume.tier_definition.tiers
}
