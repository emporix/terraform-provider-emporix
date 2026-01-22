# Example Terraform configuration for Emporix Shipping Zones

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

# Example 1: Simple single country zone
resource "emporix_shipping_zone" "germany" {
  id   = "zone-germany"
  site = "main"
  name = {
    en = "Germany"
  }

  ship_to = [
    {
      country = "DE"
    }
  ]
}

# Example 2: Zone with postal code pattern
resource "emporix_shipping_zone" "berlin" {
  id   = "zone-berlin"
  site = "main"
  name = {
    en = "Berlin Area"
  }

  ship_to = [
    {
      country     = "DE"
      postal_code = "10*"
    }
  ]
}

# Example 3: Multi-country zone with translations
resource "emporix_shipping_zone" "eu_zone" {
  id   = "zone-eu"
  site = "main"
  name = {
    en = "European Union"
    de = "Europäische Union"
    fr = "Union Européenne"
  }

  ship_to = [
    { country = "AT" },
    { country = "BE" },
    { country = "FR" },
    { country = "IT" },
    { country = "NL" },
    { country = "ES" }
  ]
}

# Example 4: Regional zone - South Germany (Stuttgart area)
resource "emporix_shipping_zone" "south_germany" {
  id   = "zone-south-de"
  site = "main"
  name = {
    en = "South Germany"
  }

  ship_to = [
    {
      country     = "DE"
      postal_code = "70*"
    }
  ]
}

# Example 5: Express delivery zone for major cities across countries
resource "emporix_shipping_zone" "express" {
  id   = "zone-express"
  site = "main"
  name = {
    en = "Express Delivery"
  }

  ship_to = [
    { country = "DE", postal_code = "10*" },  # Berlin
    { country = "AT", postal_code = "10*" },  # Vienna
    { country = "FR", postal_code = "75*" },  # Paris
    { country = "IT", postal_code = "00*" },  # Rome
  ]
}

# Example 6: Default fallback zone
resource "emporix_shipping_zone" "default" {
  id      = "zone-default"
  site    = "main"
  name    = {
    en = "Default Delivery Zone"
  }
  default = true

  ship_to = [
    { country = "DE" },
    { country = "AT" },
    { country = "CH" }
  ]
}

# Example 7: International zone
resource "emporix_shipping_zone" "international" {
  id   = "zone-international"
  site = "main"
  name = {
    en = "International"
    de = "International"
  }

  ship_to = [
    { country = "US" },
    { country = "GB" },
    { country = "CA" },
    { country = "AU" }
  ]
}

# Example 8: Specific postal code (exact match)
resource "emporix_shipping_zone" "vienna" {
  id   = "zone-vienna"
  site = "main"
  name = {
    en = "Vienna Center"
  }

  ship_to = [
    {
      country     = "AT"
      postal_code = "1010"
    }
  ]
}

# Outputs
output "zone_ids" {
  description = "Shipping zone IDs"
  value = {
    germany       = emporix_shipping_zone.germany.id
    berlin        = emporix_shipping_zone.berlin.id
    eu            = emporix_shipping_zone.eu_zone.id
    south_germany = emporix_shipping_zone.south_germany.id
    express       = emporix_shipping_zone.express.id
    default       = emporix_shipping_zone.default.id
    international = emporix_shipping_zone.international.id
    vienna        = emporix_shipping_zone.vienna.id
  }
}

output "default_zone" {
  description = "Default zone information"
  value = {
    id      = emporix_shipping_zone.default.id
    name    = emporix_shipping_zone.default.name
    default = emporix_shipping_zone.default.default
  }
}

output "zones_with_postal_codes" {
  description = "Zones that use postal code filtering"
  value = [
    emporix_shipping_zone.berlin.id,
    emporix_shipping_zone.south_germany.id,
    emporix_shipping_zone.express.id,
    emporix_shipping_zone.vienna.id
  ]
}
