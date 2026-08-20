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

# Example 4: Advanced site with mixins
# Mixins reference schemas managed by the emporix_schema resource, so
# Terraform creates the schema first and passes its id/schema_url into
# the site's mixins list.
resource "emporix_schema" "site_branding" {
  id = "site-branding"
  name = {
    en = "Site Branding"
  }
  types = ["SITE"]

  attributes = [
    {
      key = "brandColor"
      name = {
        en = "Brand Color"
      }
      description = {
        en = "Primary brand color used across the storefront, as a hex code"
      }
      type = "TEXT"
      metadata = {
        read_only = false
        localized = false
        required  = false
        nullable  = true
      }
    },
    {
      key = "featuredProductCount"
      name = {
        en = "Featured Product Count"
      }
      description = {
        en = "Number of products highlighted on the site's landing page"
      }
      type = "NUMBER"
      metadata = {
        read_only = false
        localized = false
        required  = false
        nullable  = true
      }
    }
  ]
}

resource "emporix_schema" "site_operations" {
  id = "site-operations"
  name = {
    en = "Site Operations"
  }
  types = ["SITE"]

  attributes = [
    {
      key = "isFlagshipStore"
      name = {
        en = "Is Flagship Store"
      }
      description = {
        en = "Marks this site as the brand's flagship store"
      }
      type = "BOOLEAN"
      metadata = {
        read_only = false
        localized = false
        required  = false
        nullable  = true
      }
    },
    {
      key = "siteManagerName"
      name = {
        en = "Site Manager Name"
      }
      description = {
        en = "Full name of the employee responsible for the site"
      }
      type = "TEXT"
      metadata = {
        read_only = false
        localized = false
        required  = false
        nullable  = true
      }
    }
  ]
}

resource "emporix_sitesettings" "flagship_site" {
  code             = "flagship"
  name             = "Flagship Site with Custom Attributes"
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
      name       = emporix_schema.site_branding.id
      schema_url = emporix_schema.site_branding.schema_url
      fields = jsonencode({
        brandColor           = "#FF5733"
        featuredProductCount = 100
      })
    },
    {
      name       = emporix_schema.site_operations.id
      schema_url = emporix_schema.site_operations.schema_url
      fields = jsonencode({
        isFlagshipStore = true
        siteManagerName = "John Doe"
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

# # Outputs
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