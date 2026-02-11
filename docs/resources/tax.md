---
page_title: "emporix_tax Resource - terraform-provider-emporix"
subcategory: ""
description: |-
  Manages tax configurations for countries.
---

# emporix_tax (Resource)

Manages tax configurations that define tax classes and rates for specific countries. Each country can have multiple tax classes (e.g., standard, reduced, zero) with different tax rates.

**Delete Behavior:** When you remove the resource from Terraform or run `terraform destroy`, the tax configuration is **deleted** from Emporix.

## Prerequisites

Tax configurations are country-specific. Ensure the country codes you use follow the Country Service standards (ISO 3166-1 alpha-2).

## Example Usage

### Basic Tax Configuration (Single Country)

```terraform
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
        en = "Standard 23% VAT rate for most goods and services"
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
```

### Multiple Tax Classes with Different Rates

```terraform
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
```

### Multiple Countries

```terraform
# United States
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

# United Kingdom
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
    },
    {
      code = "EXAMPLE_REDUCED"
      name = {
        en = "Reduced VAT"
      }
      rate  = 0.05
      order = 2
    },
    {
      code = "EXAMPLE_ZERO"
      name = {
        en = "Zero-rated"
      }
      rate  = 0.0
      order = 3
    }
  ]
}

# France
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
```

### Dynamic Tax Configuration

```terraform
locals {
  eu_countries = {
    DE = { standard = 0.19, reduced = 0.07 }
    FR = { standard = 0.20, reduced = 0.055 }
    IT = { standard = 0.22, reduced = 0.10 }
    ES = { standard = 0.21, reduced = 0.10 }
    NL = { standard = 0.21, reduced = 0.09 }
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

output "eu_tax_configs" {
  value = {
    for country, config in emporix_tax.eu :
    country => {
      country_code = config.country_code
      tax_classes  = config.tax_classes
    }
  }
}
```

## Schema

### Required

- `country_code` (String) Country code (e.g., 'US', 'DE', 'GB'). Must follow Country Service standards (ISO 3166-1 alpha-2). Cannot be changed after creation. Changing this forces a new resource to be created.
- `tax_classes` (List of Objects) List of tax classes for this country. At least one tax class is required. Tax classes are sorted by their order value. (see [tax_classes](#tax_classes) below)

### Nested Schema for `tax_classes`

Required nested block list. Each tax class defines a rate category.

**Required:**

- `code` (String) Unique code for this tax class (e.g., 'STANDARD', 'REDUCED', 'ZERO').
- `name` (Map of String) Tax class name as a map of language codes to translated names. Example: {en = "Standard Rate", de = "Normalsteuersatz"}. At least one language is required.
- `rate` (Number) Tax rate as a decimal. Examples: 0.19 for 19%, 0.07 for 7%, 0.0 for 0%.

**Optional:**

- `description` (Map of String) Optional description as a map of language codes to translated descriptions.
- `order` (Number) Display order for this tax class. Tax classes are sorted by this value in ascending order. Lower values appear first.
- `is_default` (Boolean) Whether this is the default tax class for the country. Only one tax class can be default. Defaults to false.

## Import

Tax configurations can be imported using the country code:

```shell
terraform import emporix_tax.poland PL
terraform import emporix_tax.germany DE
```

After importing, Terraform will manage the tax configuration. You'll need to provide the complete `tax_classes` configuration in your Terraform files.

## Required OAuth Scopes

To manage tax configurations, your client_id/secret pair must have:

**Required Scopes:**
- `tax.tax_read` - Required for reading tax configurations
- `tax.tax_manage` - Required for creating, updating, and deleting tax configurations

## Outputs

### country_code

The country code for this tax configuration. This is the unique identifier for the resource.

```terraform
output "poland_tax_country" {
  value = emporix_tax.poland.country_code
  # Output: "PL"
}
```

### tax_classes

The list of tax classes with their configurations. You can access individual tax classes by index:

```terraform
output "standard_rate" {
  value = emporix_tax.poland.tax_classes[0].rate
  # Output: 0.23
}

output "standard_name" {
  value = emporix_tax.poland.tax_classes[0].name["en"]
  # Output: "Standard VAT Rate"
}
```

You can also iterate over tax classes:

```terraform
output "all_rates" {
  value = [for tc in emporix_tax.poland.tax_classes : tc.rate]
  # Output: [0.23, 0.08]
}
```
