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
# provider "emporix" {
#   tenant       = var.emporix_tenant       # Or use EMPORIX_TENANT env var
#   access_token = var.emporix_access_token # Or use EMPORIX_ACCESS_TOKEN env var
# }

# # Variables
# variable "emporix_tenant" {
#   description = "Emporix tenant name"
#   type        = string
#   sensitive   = false
# }

# variable "emporix_access_token" {
#   description = "Emporix OAuth2 access token"
#   type        = string
#   sensitive   = true
# }
provider "emporix" {
  tenant       = "tenant"     # Or use EMPORIX_TENANT env var
  client_id = "xxx"
  client_secret = "xxx"
  api_url      = "https://api-develop.emporix.io" # Optional, defaults to https://api.emporix.io
}

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
      street_number = "1501"
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

  cart_calculation_scale = 2

  # Mixins - unified format with schema URL and data in single objects
  mixins = [
    {
      name       = "test3"
      schema_url = "https://res.cloudinary.com/saas-ag/raw/upload/schemata2/ppmdev/test3_v1.json"
      fields = jsonencode({
        field3 = "value3"
      })
    }
  ]

  ship_to_countries = ["US"]

  home_base = {
    address = {
      country = "US"
      city    = "New York"
      zip_code = "AAA"
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
