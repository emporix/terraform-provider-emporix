# Tenant Configuration Examples

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
  scope         = "tenant=${var.emporix_tenant} configuration.configuration_view configuration.configuration_manage"
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

# Example 1: Simple String Configuration
resource "emporix_tenant_configuration" "project_country" {
  key   = "test_project_country"
  value = jsonencode("US")
}

# Example 2: Object Configuration - Tax Settings
resource "emporix_tenant_configuration" "tax_config" {
  key = "test_taxConfiguration"
  value = jsonencode({
    taxClassOrder = ["FULL", "HALF", "ZERO"]
    taxClasses = {
      FULL = 19
      HALF = 7
      ZERO = 0
    }
  })
}

# Example 3: Array Configuration - Project Languages
# Note: project_lang value should be a JSON STRING (not a direct array)
# This matches the API expectation: "value": "[{\"id\":\"en\",...}]"
resource "emporix_tenant_configuration" "project_languages" {
  key = "test_project_lang"
  # Double-encode: first to create JSON array string, then to make it a JSON value
  value = jsonencode(jsonencode([
    {
      id       = "en"
      label    = "English"
      default  = true
      required = true
    },
    {
      id       = "de"
      label    = "German"
      default  = false
      required = false
    },
    {
      id       = "fr"
      label    = "French"
      default  = false
      required = false
    }
  ]))
}

# Example 4: Array Configuration - Project Currencies
# Note: project_curr value should also be a JSON STRING
resource "emporix_tenant_configuration" "project_currencies" {
  key = "test_project_curr"
  # Double-encode: first to create JSON array string, then to make it a JSON value
  value = jsonencode(jsonencode([
    {
      id       = "USD"
      label    = "US Dollar"
      default  = true
      required = true
    },
    {
      id       = "EUR"
      label    = "Euro"
      default  = false
      required = false
    },
    {
      id       = "GBP"
      label    = "British Pound"
      default  = false
      required = false
    }
  ]))
}

# Example 4b: Cleaner approach using locals for JSON string values
locals {
  # Build the array structure
  currencies_array = [
    {
      id       = "USD"
      label    = "US Dollar"
      default  = true
      required = true
    },
    {
      id       = "EUR"
      label    = "Euro"
      default  = false
      required = false
    }
  ]
  
  # Convert to JSON string once
  currencies_json_string = jsonencode(local.currencies_array)
}

resource "emporix_tenant_configuration" "project_currencies_clean" {
  key = "test_project_curr_alternative"
  # Then encode the JSON string as a JSON value
  value = jsonencode(local.currencies_json_string)
}


# Example 5: Storefront Configuration
resource "emporix_tenant_configuration" "storefront_host" {
  key   = "test_storefront.host"
  value = jsonencode("example.com")
}

resource "emporix_tenant_configuration" "storefront_page" {
  key   = "test_storefront.htmlPage"
  value = jsonencode("index.html")
}

# Example 6: Email Configuration
resource "emporix_tenant_configuration" "email_from" {
  key   = "test_cust.notification.email.from"
  value = jsonencode("[email protected]")
}

# Example 7: Customer URLs Configuration
resource "emporix_tenant_configuration" "password_reset_url" {
  key   = "test_customer.passwordreset.redirecturl"
  value = jsonencode("https://example.com/reset-password?token=")
}

resource "emporix_tenant_configuration" "email_change_url" {
  key   = "test_customer.changeemail.redirecturl"
  value = jsonencode("https://example.com/confirm-email?token=")
}

resource "emporix_tenant_configuration" "account_deletion_url" {
  key   = "test_customer.deletion.redirecturl"
  value = jsonencode("https://example.com/delete-account?token=")
}

# Example 8: Complex Configuration - Unit Handling
resource "emporix_tenant_configuration" "unit_config" {
  key = "test_unitConf"
  value = jsonencode([
    {
      name        = "UnitConfiguration"
      description = "This object holds the unit configurations and the factors to calculate between different units"
      units = {
        LTR = {
          availableUnitValue = 1
          conversion = {
            MLT = 1000
          }
        }
        GRM = {
          availableUnitValue = 100
          conversion = {
            KGM = 0.001
          }
        }
        KGM = {
          availableUnitValue = 1
          conversion = {
            GRM = 1000
          }
        }
        MLT = {
          availableUnitValue = 100
          conversion = {
            LTR = 0.001
          }
        }
      }
    }
  ])
}

# Example 9: Packaging Configuration
resource "emporix_tenant_configuration" "packaging_config" {
  key = "test_packagingConf"
  value = jsonencode({
    packagingGroupOptions = [
      "1_cooling_product",
      "2_standard",
      "3_fragile"
    ]
    packagingPositionOptions = [
      "1_bottom_robust",
      "2_bottom_insensitive",
      "3_middle_standard",
      "4_top_sensitive"
    ]
  })
}

# Example 10: Secured Configuration (encrypted)
resource "emporix_tenant_configuration" "api_key" {
  key     = "test_external_api_key"
  value   = jsonencode("secret-api-key-value")
  secured = true  # Value will be encrypted
}

# Example 11: Boolean Configuration
resource "emporix_tenant_configuration" "feature_flag" {
  key   = "test_enable_new_checkout"
  value = jsonencode(true)
}

# Example 12: Numeric Configuration
resource "emporix_tenant_configuration" "timeout" {
  key   = "test_api.timeout_seconds"
  value = jsonencode(30)
}

# Example 13: Using for_each with Map
locals {
  simple_configs = {
    "test_app.environment"  = "production"
    "test_app.version"      = "1.0.0"
    "test_app.debug_mode"   = false
    "test_app.max_retries"  = 3
  }
}

resource "emporix_tenant_configuration" "app_configs" {
  for_each = local.simple_configs

  key   = each.key
  value = jsonencode(each.value)
}

# Example 14: Using for_each with Complex Objects
locals {
  endpoint_configs = {
    "test_api_endpoint" = {
      url     = "https://api.example.com"
      timeout = 30
      retries = 3
    }
    "test_cdn_endpoint" = {
      url     = "https://cdn.example.com"
      timeout = 60
      retries = 5
    }
  }
}

resource "emporix_tenant_configuration" "endpoints" {
  for_each = local.endpoint_configs

  key   = each.key
  value = jsonencode(each.value)
}

# Example 15: Environment-Specific Configuration
variable "environment" {
  type        = string
  description = "Environment (dev, staging, prod)"
  default     = "dev"
}

resource "emporix_tenant_configuration" "env_config" {
  key = "test_environment.config"
  value = jsonencode({
    name        = var.environment
    is_prod     = var.environment == "prod"
    api_url     = var.environment == "prod" ? "https://api.example.com" : "https://api-dev.example.com"
    debug_mode  = var.environment != "prod"
  })
}

# Outputs
output "project_country" {
  description = "Project country configuration"
  value       = jsondecode(emporix_tenant_configuration.project_country.value)
}

output "tax_configuration" {
  description = "Tax configuration object"
  value       = jsondecode(emporix_tenant_configuration.tax_config.value)
}

output "project_languages" {
  description = "Configured project languages"
  value       = jsondecode(emporix_tenant_configuration.project_languages.value)
}

output "all_app_configs" {
  description = "All application configurations"
  value = {
    for key, config in emporix_tenant_configuration.app_configs :
    key => {
      value   = jsondecode(config.value)
      version = config.version
    }
  }
}

output "configuration_versions" {
  description = "Versions of all configurations"
  value = {
    country    = emporix_tenant_configuration.project_country.version
    tax        = emporix_tenant_configuration.tax_config.version
    languages  = emporix_tenant_configuration.project_languages.version
  }
}
