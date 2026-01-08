# Example Terraform configuration for Emporix Currencies

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

# Example 1: Currency with single language
resource "emporix_currency" "usd" {
  code = "USD"
  name = {
    en = "US Dollar"
  }
}

# Example 2: Currency with multiple translations
resource "emporix_currency" "eur" {
  code = "EUR"
  name = {
    en = "Euro"
    de = "Euro"
    fr = "Euro"
  }
}

# Example 3: Currency with custom translations
resource "emporix_currency" "gbp" {
  code = "GBP"
  name = {
    en = "British Pound Sterling"
    de = "Britisches Pfund Sterling"
    fr = "Livre Sterling"
  }
}

# Example 4: Multiple currencies for European market
resource "emporix_currency" "pln" {
  code = "PLN"
  name = {
    en = "Polish Zloty"
  }
}

resource "emporix_currency" "czk" {
  code = "CZK"
  name = {
    en = "Czech Koruna"
  }
}

resource "emporix_currency" "huf" {
  code = "HUF"
  name = {
    en = "Hungarian Forint"
  }
}

# Example 5: Using for_each for multiple currencies
locals {
  asian_currencies = {
    JPY = "Japanese Yen"
    CNY = "Chinese Yuan"
    KRW = "South Korean Won"
    SGD = "Singapore Dollar"
    THB = "Thai Baht"
  }
}

resource "emporix_currency" "asian" {
  for_each = local.asian_currencies
  
  code = each.key
  name = {
    en = each.value
  }
}

# Outputs
output "major_currencies" {
  description = "Major currency codes"
  value = {
    usd = emporix_currency.usd.code
    eur = emporix_currency.eur.code
    gbp = emporix_currency.gbp.code
  }
}

output "currency_names_english" {
  description = "Currency names in English"
  value = {
    usd = lookup(emporix_currency.usd.name, "en", "")
    eur = lookup(emporix_currency.eur.name, "en", "")
    gbp = lookup(emporix_currency.gbp.name, "en", "")
  }
}

output "european_currencies" {
  description = "European currency codes"
  value = {
    pln = emporix_currency.pln.code
    czk = emporix_currency.czk.code
    huf = emporix_currency.huf.code
  }
}

output "asian_currencies" {
  description = "Asian currency codes"
  value = [for curr in emporix_currency.asian : curr.code]
}

output "all_currency_names" {
  description = "All currencies with their names"
  value = merge(
    {
      for key, curr in {
        usd = emporix_currency.usd
        eur = emporix_currency.eur
        gbp = emporix_currency.gbp
        pln = emporix_currency.pln
        czk = emporix_currency.czk
        huf = emporix_currency.huf
      } : key => {
        code    = curr.code
        name_en = lookup(curr.name, "en", "")
      }
    },
    {
      for key, curr in emporix_currency.asian : key => {
        code    = curr.code
        name_en = lookup(curr.name, "en", "")
      }
    }
  )
}
