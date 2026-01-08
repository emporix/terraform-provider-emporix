---
page_title: "emporix_currency Resource - terraform-provider-emporix"
subcategory: ""
description: |-
  Currency resource for managing currencies in Emporix.
---

# emporix_currency (Resource)

Manages currencies in Emporix. Currencies must be ISO-4217 compliant 3-letter codes.

**Delete Behavior:** When you remove the resource from Terraform or run `terraform destroy`, the currency is **deleted** from Emporix. This also deletes all exchange rates and prices associated with the currency.

## Example Usage

### Basic Currency (Single Language)

```terraform
resource "emporix_currency" "usd" {
  code = "USD"
  name = {
    en = "US Dollar"
  }
}
```

### Currency with Multiple Translations

```terraform
resource "emporix_currency" "eur" {
  code = "EUR"
  name = {
    en = "Euro"
    de = "Euro"
    fr = "Euro"
    es = "Euro"
  }
}
```

### Multiple Currencies

```terraform
resource "emporix_currency" "usd" {
  code = "USD"
  name = {
    en = "US Dollar"
  }
}

resource "emporix_currency" "eur" {
  code = "EUR"
  name = {
    en = "Euro"
  }
}

resource "emporix_currency" "gbp" {
  code = "GBP"
  name = {
    en = "British Pound"
  }
}
```

### Using with Site Settings

```terraform
resource "emporix_currency" "usd" {
  code = "USD"
  name = {
    en = "US Dollar"
  }
}

resource "emporix_currency" "eur" {
  code = "EUR"
  name = {
    en = "Euro"
  }
}

resource "emporix_sitesettings" "main" {
  code             = "main-site"
  name             = "Main Store"
  active           = true
  default          = true
  default_language = "en"
  languages        = ["en", "de"]
  
  # Reference currencies
  currency = emporix_currency.usd.code
  
  available_currencies = [
    emporix_currency.usd.code,
    emporix_currency.eur.code,
  ]
  
  ship_to_countries = ["US", "GB", "DE"]
  
  home_base = {
    address = {
      zip_code = "10001"
      city     = "New York"
      country  = "US"
    }
  }
}
```

### Dynamic Currency Creation

```terraform
locals {
  currencies = {
    USD = "US Dollar"
    EUR = "Euro"
    GBP = "British Pound"
    JPY = "Japanese Yen"
  }
}

resource "emporix_currency" "currencies" {
  for_each = local.currencies
  
  code = each.key
  name = {
    en = each.value
  }
}
```

### Asian Currencies

```terraform
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

output "created_currencies" {
  value = {
    for code, currency in emporix_currency.asian :
    code => {
      code    = currency.code
      name_en = currency.name["en"]
    }
  }
}
```

## Schema

### Required

- `code` (String) Currency code (3-letter uppercase ISO-4217 code, e.g., USD, EUR, GBP). Cannot be changed after creation. Changing this forces a new resource to be created.
- `name` (Map of String) Currency name as a map of language code to name (e.g., {"en": "US Dollar", "de": "US-Dollar"}). You must provide at least one language translation.

## Import

You can import existing currencies:

```shell
terraform import emporix_currency.usd USD
```

After importing, Terraform will manage the currency. Note that you'll need to provide the `name` field in your configuration after import.

## Required OAuth Scopes

To manage currencies, your client_id/secret pair (used in provider section) must have the following scopes:

**Required Scopes:**
- `currency.currency_read` - Required for reading currency information
- `currency.currency_manage` - Required for creating, updating, and deleting currencies

## ISO-4217 Compliance

Currency codes must be valid ISO-4217 codes. The code must:
- Be exactly 3 letters
- Be uppercase
- Be a valid currency code from the ISO-4217 standard

### Valid Examples
- `USD` - US Dollar
- `EUR` - Euro
- `GBP` - British Pound
- `JPY` - Japanese Yen
- `CHF` - Swiss Franc
- `PLN` - Polish Zloty

### Invalid Examples
- `usd` - ❌ Lowercase (must be USD)
- `USDT` - ❌ Not ISO-4217 (cryptocurrency)
- `BTC` - ❌ Not ISO-4217 (cryptocurrency)
- `US` - ❌ Only 2 letters (must be 3)

See [ISO-4217 Currency Codes](https://www.iso.org/iso-4217-currency-codes.html) for the complete list.

## Common Currency Codes

### Major World Currencies
- `USD` - US Dollar
- `EUR` - Euro
- `GBP` - British Pound Sterling
- `JPY` - Japanese Yen
- `CHF` - Swiss Franc
- `CAD` - Canadian Dollar
- `AUD` - Australian Dollar
- `CNY` - Chinese Yuan

### European Currencies
- `EUR` - Euro (Eurozone)
- `GBP` - British Pound
- `PLN` - Polish Zloty
- `CZK` - Czech Koruna
- `HUF` - Hungarian Forint
- `RON` - Romanian Leu
- `SEK` - Swedish Krona
- `NOK` - Norwegian Krone
- `DKK` - Danish Krone

### Asian Currencies
- `JPY` - Japanese Yen
- `CNY` - Chinese Yuan
- `KRW` - South Korean Won
- `INR` - Indian Rupee
- `SGD` - Singapore Dollar
- `THB` - Thai Baht
- `MYR` - Malaysian Ringgit
- `IDR` - Indonesian Rupiah

### Latin American Currencies
- `MXN` - Mexican Peso
- `BRL` - Brazilian Real
- `ARS` - Argentine Peso
- `CLP` - Chilean Peso
- `COP` - Colombian Peso

## Delete Warning

⚠️ **Important:** Deleting a currency removes:
- The currency itself
- All exchange rates for this currency
- All prices in this currency (asynchronous operation)

Make sure the currency is not actively used in your system before deleting it.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

### name

The `name` attribute contains a map of language codes to translated currency names. You provide this map when creating the currency, and it's stored exactly as you provide it.

```terraform
resource "emporix_currency" "usd" {
  code = "USD"
  name = {
    en = "US Dollar"
    de = "US-Dollar"
    fr = "Dollar américain"
  }
}

output "usd_names" {
  value = emporix_currency.usd.name
}
```

Output:
```
{
  "en" = "US Dollar"
  "de" = "US-Dollar"
  "fr" = "Dollar américain"
}
```

You can access specific languages:

```terraform
output "usd_english_name" {
  value = emporix_currency.usd.name["en"]
  # Output: "US Dollar"
}

output "usd_german_name" {
  value = emporix_currency.usd.name["de"]
  # Output: "US-Dollar"
}
```

## Notes

- Currency codes are immutable. If you need to change a code, you must destroy and recreate the resource.
- The `name` field is required and must contain at least one language translation.
- You can provide translations in as many languages as you need.
- Currencies are system-wide and can be referenced by site settings and other resources.
- Deleting a currency is a destructive operation that removes all related exchange rates and prices.

