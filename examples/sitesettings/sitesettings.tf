# Example Terraform configuration for Emporix Site Settings

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
  
  # Option 1: Use client credentials (recommended)
  client_id     = var.emporix_client_id
  client_secret = var.emporix_client_secret
  scope         = "tenant=${var.emporix_tenant} site.site_read site.site_manage"
  
  # Option 2: Use pre-generated token
  # access_token = var.emporix_access_token
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

# Uncomment if using access token method
# variable "emporix_access_token" {
#   description = "Emporix OAuth2 access token"
#   type        = string
#   sensitive   = true
# }

# Example 1: Basic site configuration
resource "emporix_sitesettings" "us_site" {
  code             = "us-main"
  name             = "United States Main Site"
  active           = true
  default          = false
  default_language = "en"
  languages        = ["en", "es"]
  currency         = "USD"
  
  ship_to_countries = ["US"]
  
  home_base = {
    address = {
      country       = "US"
      zip_code      = "10036"
      city          = "New York"
      street        = "Broadway"
      street_number = "1500"
      state         = "NY"
    }
  }
}

# Example 2: European site with multiple currencies
resource "emporix_sitesettings" "eu_site" {
  code                 = "eu-main"
  name                 = "European Main Site"
  active               = true
  default              = false
  includes_tax         = true
  default_language     = "en"
  languages            = ["en", "de", "fr", "es", "it"]
  currency             = "EUR"
  available_currencies = ["EUR", "GBP", "CHF"]
  
  ship_to_countries = [
    "DE", "FR", "IT", "ES", "NL", 
    "BE", "AT", "CH", "PL", "SE"
  ]
  
  tax_calculation_address_type = "SHIPPING_ADDRESS"
  decimal_points               = 2
  
  home_base = {
    address = {
      country       = "DE"
      zip_code      = "10115"
      city          = "Berlin"
      street        = "Unter den Linden"
      street_number = "1"
    }
    location = {
      latitude  = 52.5200
      longitude = 13.4050
    }
  }
  
  assisted_buying = {
    storefront_url = "https://shop.example.com/eu"
  }
}

# Example 3: UK site post-Brexit
resource "emporix_sitesettings" "uk_site" {
  code             = "uk-main"
  name             = "United Kingdom Main Site"
  active           = true
  default          = false
  includes_tax     = true
  default_language = "en"
  languages        = ["en"]
  currency         = "GBP"
  
  ship_to_countries = ["GB"]
  
  home_base = {
    address = {
      country  = "GB"
      zip_code = "SW1A 1AA"
      city     = "London"
      street   = "Westminster"
    }
    location = {
      latitude  = 51.5074
      longitude = -0.1278
    }
  }
}

# Example 4: Advanced site with mixins and metadata
resource "emporix_sitesettings" "advanced_site" {
  code             = "advanced"
  name             = "Advanced Site with Mixins"
  active           = true
  default          = false
  default_language = "en"
  languages        = ["en"]
  currency         = "USD"
  
  ship_to_countries = ["US"]
  
  cart_calculation_scale = 2
  
  # Mixins - unified format with schema URL and data in single objects
  mixins = [
    {
      name       = "customFields"
      schema_url = "https://api.example.com/schemas/custom-fields_v1.json"
      fields = jsonencode({
        brandColor    = "#FF5733"
        customMessage = "Welcome to our store"
        enableFeatureX = true
      })
    },
    {
      name       = "seoSettings"
      schema_url = "https://api.example.com/schemas/seo_v2.json"
      fields = jsonencode({
        metaTitle       = "Best Online Store"
        metaDescription = "Shop the best products online"
        canonicalUrl    = "https://example.com"
      })
    }
  ]
  
  home_base = {
    address = {
      country  = "US"
      zip_code = "10001"
      city     = "New York"
    }
  }
}

# Outputs
output "us_site_code" {
  description = "US site code"
  value       = emporix_sitesettings.us_site.code
}

output "eu_site_code" {
  description = "EU site code"
  value       = emporix_sitesettings.eu_site.code
}

output "uk_site_code" {
  description = "UK site code"
  value       = emporix_sitesettings.uk_site.code
}
