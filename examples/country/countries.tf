# Example Terraform configuration for Emporix Countries

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
  
  # Use client credentials
  client_id     = var.emporix_client_id
  client_secret = var.emporix_client_secret
  scope         = "tenant=${var.emporix_tenant} country.country_read country.country_manage"
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

# IMPORTANT: Countries must be imported before they can be managed
# Run these commands first:
#   terraform import emporix_country.usa US
#   terraform import emporix_country.canada CA
#   terraform import emporix_country.uk GB
#   terraform import emporix_country.germany DE

# Example 1: Active country (United States)
resource "emporix_country" "usa" {
  code   = "US"
  active = true
}

# Example 2: Active country (Canada)
resource "emporix_country" "canada" {
  code   = "CA"
  active = true
}

# Example 3: Active country (United Kingdom)
resource "emporix_country" "uk" {
  code   = "GB"
  active = true
}

# Example 4: Inactive country (Germany)
resource "emporix_country" "germany" {
  code   = "DE"
  active = false
}

# Outputs
output "active_countries" {
  description = "List of active country codes"
  value = [
    for country in [
      emporix_country.usa,
      emporix_country.canada,
      emporix_country.uk,
      emporix_country.germany,
    ] : country.code if country.active
  ]
}

output "country_names" {
  description = "Country codes with their English names"
  value = {
    usa     = lookup(emporix_country.usa.name, "en", "United States")
    canada  = lookup(emporix_country.canada.name, "en", "Canada")
    uk      = lookup(emporix_country.uk.name, "en", "United Kingdom")
    germany = lookup(emporix_country.germany.name, "en", "Germany")
  }
}

output "country_regions" {
  description = "Countries and their regions"
  value = {
    usa     = emporix_country.usa.regions
    canada  = emporix_country.canada.regions
    uk      = emporix_country.uk.regions
    germany = emporix_country.germany.regions
  }
}
