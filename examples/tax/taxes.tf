# Example Terraform configuration for Emporix Tax Configurations

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

# Example 1: Basic tax configuration (Poland)
resource "emporix_tax" "poland" {
  country_code = "PL"

  tax_classes = [
    {
      code = "EXAMPLE_STANDARD"
      name = {
        en = "Standard VAT Rate"
        pl = "Stawka podstawowa VAT"
      }
      rate       = 0.23
      is_default = true
      order      = 1
      description = {
        en = "Standard 23% VAT rate"
      }
    },
    {
      code = "EXAMPLE_REDUCED"
      name = {
        en = "Reduced VAT Rate"
        pl = "Stawka obniżona VAT"
      }
      rate  = 0.08
      order = 2
    }
  ]
}

# Example 2: Multiple tax classes (Germany)
resource "emporix_tax" "germany" {
  country_code = "DE"

  tax_classes = [
    {
      code = "EXAMPLE_STANDARD"
      name = {
        en = "Standard VAT"
        de = "Regelsteuersatz"
      }
      rate       = 0.19
      is_default = true
      order      = 1
    },
    {
      code = "EXAMPLE_REDUCED"
      name = {
        en = "Reduced VAT"
        de = "Ermäßigter Steuersatz"
      }
      rate  = 0.07
      order = 2
    },
    {
      code = "EXAMPLE_ZERO"
      name = {
        en = "Zero Rate"
        de = "Nullsteuersatz"
      }
      rate  = 0.0
      order = 3
    }
  ]
}

# Example 3: United States (simple sales tax)
resource "emporix_tax" "us" {
  country_code = "US"

  tax_classes = [
    {
      code = "EXAMPLE_STANDARD"
      name = {
        en = "Standard Sales Tax"
      }
      rate       = 0.07
      is_default = true
      order      = 1
    }
  ]
}

# Example 4: United Kingdom (with descriptions)
resource "emporix_tax" "uk" {
  country_code = "GB"

  tax_classes = [
    {
      code = "EXAMPLE_STANDARD"
      name = {
        en = "Standard VAT"
      }
      rate       = 0.20
      is_default = true
      order      = 1
      description = {
        en = "Standard 20% VAT for most goods and services"
      }
    },
    {
      code = "EXAMPLE_REDUCED"
      name = {
        en = "Reduced VAT"
      }
      rate  = 0.05
      order = 2
      description = {
        en = "Reduced 5% VAT for specific goods"
      }
    },
    {
      code = "EXAMPLE_ZERO"
      name = {
        en = "Zero-rated"
      }
      rate  = 0.0
      order = 3
      description = {
        en = "Zero-rated goods and services"
      }
    }
  ]
}

# Example 5: France (multiple rates)
resource "emporix_tax" "france" {
  country_code = "FR"

  tax_classes = [
    {
      code = "EXAMPLE_STANDARD"
      name = {
        en = "Standard VAT"
        fr = "TVA normale"
      }
      rate       = 0.20
      is_default = true
      order      = 1
    },
    {
      code = "EXAMPLE_INTERMEDIATE"
      name = {
        en = "Intermediate Rate"
        fr = "Taux intermédiaire"
      }
      rate  = 0.10
      order = 2
    },
    {
      code = "EXAMPLE_REDUCED"
      name = {
        en = "Reduced Rate"
        fr = "Taux réduit"
      }
      rate  = 0.055
      order = 3
    }
  ]
}

# Example 6: Dynamic configuration for multiple EU countries
locals {
  eu_countries = {
    IT = { name = "Italy", standard = 0.22, reduced = 0.10 }
    ES = { name = "Spain", standard = 0.21, reduced = 0.10 }
    NL = { name = "Netherlands", standard = 0.21, reduced = 0.09 }
    BE = { name = "Belgium", standard = 0.21, reduced = 0.06 }
    AT = { name = "Austria", standard = 0.20, reduced = 0.10 }
  }
}

resource "emporix_tax" "eu" {
  for_each = local.eu_countries

  country_code = each.key

  tax_classes = [
    {
      code = "EXAMPLE_STANDARD"
      name = {
        en = "Standard VAT"
      }
      rate       = each.value.standard
      is_default = true
      order      = 1
    },
    {
      code = "EXAMPLE_REDUCED"
      name = {
        en = "Reduced VAT"
      }
      rate  = each.value.reduced
      order = 2
    }
  ]
}

# Outputs
output "configured_countries" {
  description = "List of countries with tax configurations"
  value = concat(
    [emporix_tax.poland.country_code],
    [emporix_tax.germany.country_code],
    [emporix_tax.us.country_code],
    [emporix_tax.uk.country_code],
    [emporix_tax.france.country_code],
    [for tax in emporix_tax.eu : tax.country_code]
  )
}

output "tax_details" {
  description = "Tax configuration details for specific countries"
  value = {
    poland = {
      country_code = emporix_tax.poland.country_code
      tax_classes  = emporix_tax.poland.tax_classes
    }
    germany = {
      country_code = emporix_tax.germany.country_code
      tax_classes  = emporix_tax.germany.tax_classes
    }
    uk = {
      country_code = emporix_tax.uk.country_code
      tax_classes  = emporix_tax.uk.tax_classes
    }
  }
}

output "eu_tax_rates" {
  description = "EU country tax rates"
  value = {
    for country, config in emporix_tax.eu :
    country => {
      country_code  = config.country_code
      standard_rate = config.tax_classes[0].rate
      reduced_rate  = config.tax_classes[1].rate
    }
  }
}

output "standard_rates" {
  description = "Standard tax rates for all configured countries"
  value = {
    PL = emporix_tax.poland.tax_classes[0].rate
    DE = emporix_tax.germany.tax_classes[0].rate
    US = emporix_tax.us.tax_classes[0].rate
    GB = emporix_tax.uk.tax_classes[0].rate
    FR = emporix_tax.france.tax_classes[0].rate
  }
}
