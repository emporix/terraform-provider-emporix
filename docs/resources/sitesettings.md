---
page_title: "emporix_sitesettings Resource - terraform-provider-emporix"
subcategory: ""
description: |-
  Manages Emporix Site Settings.
---

# emporix_sitesettings (Resource)

Manages site configuration in Emporix, including languages, currencies, shipping countries, and home base settings.

## Example Usage

### Basic Site

```hcl
resource "emporix_sitesettings" "us_site" {
  code             = "us-main"
  name             = "United States Main Site"
  active           = true
  default_language = "en"
  languages        = ["en", "es"]
  currency         = "USD"
  ship_to_countries = ["US"]
  
  home_base = {
    address = {
      country  = "US"
      zip_code = "10001"
      city     = "New York"
    }
  }
}
```

### Site with Home Base

```hcl
resource "emporix_sitesettings" "eu_site" {
  code             = "eu-main"
  name             = "European Main Site"
  active           = true
  default_language = "en"
  languages        = ["en", "de", "fr"]
  currency         = "EUR"
  available_currencies = ["EUR", "GBP", "CHF"]
  ship_to_countries = ["DE", "FR", "IT", "ES"]
  
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
}
```

### Complete Configuration

```hcl
resource "emporix_sitesettings" "full_site" {
  code                          = "complete"
  name                          = "Complete Site Example"
  active                        = true
  default                       = false
  includes_tax                  = true
  default_language              = "en"
  languages                     = ["en", "de", "fr", "es"]
  currency                      = "EUR"
  available_currencies          = ["EUR", "USD", "GBP"]
  ship_to_countries             = ["US", "GB", "DE", "FR"]
  tax_calculation_address_type  = "SHIPPING_ADDRESS"
  decimal_points                = 2
  
  home_base = {
    address = {
      country       = "US"
      zip_code      = "10036"
      city          = "New York"
      street        = "Broadway"
      street_number = "1500"
      state         = "NY"
    }
    location = {
      latitude  = 40.7568658044745
      longitude = -73.9858713458565
    }
  }
  
  assisted_buying = {
    storefront_url = "https://example.com/store"
  }
  
  mixins = [
    {
      name       = "productSettings"
      schema_url = "https://example.com/mixins/product.json"
      fields = jsonencode({
        defaultCategory = "electronics"
        taxRate         = 0.19
      })
    },
    {
      name       = "customSettings"
      schema_url = "https://example.com/mixins/custom.json"
      fields = jsonencode({
        brandColor = "#FF5733"
        theme      = "modern"
      })
    }
  ]
}
```

### Site with Mixins

```hcl
resource "emporix_sitesettings" "site_with_mixins" {
  code              = "us-custom"
  name              = "US Site with Custom Fields"
  active            = true
  default_language  = "en"
  languages         = ["en"]
  currency          = "USD"
  ship_to_countries = ["US"]
  
  mixins = [
    {
      name       = "customFields"
      schema_url = "https://api.example.com/schemas/custom-fields_v1.json"
      fields = jsonencode({
        brandColor     = "#FF5733"
        customMessage  = "Welcome to our store"
        enableFeatureX = true
      })
    },
    {
      name       = "seoSettings"
      schema_url = "https://api.example.com/schemas/seo_v2.json"
      fields = jsonencode({
        metaTitle       = "Best Online Store"
        metaDescription = "Shop the best products"
        canonicalUrl    = "https://example.com"
      })
    }
  ]
}
```

## Schema

### Required

- `code` (String) Site unique identifier. Cannot be changed after creation.
- `name` (String) Site display name.
- `active` (Boolean) Flag indicating whether the site is active.
- `default_language` (String) Site's default language, compliant with ISO 639-1 standard (e.g., "en", "de").
- `languages` (List of String) Languages supported by the site. Must be compliant with ISO 639-1 standard.
- `currency` (String) Currency used by the site, compliant with ISO 4217 standard (e.g., "USD", "EUR").
- `ship_to_countries` (List of String) Codes of countries to which the site ships products. Must be compliant with ISO 3166-1 alpha-2 standard.
- `home_base` (Object) Home base configuration for the site. See [Home Base](#nested-schema-for-home_base) below.

### Optional

- `default` (Boolean) Flag indicating whether the site is the tenant default site. Defaults to `false`.
- `includes_tax` (Boolean) Indicates whether prices should be returned in gross (true) or net (false).
- `available_currencies` (List of String) List of currencies supported by the site (ISO 4217).
- `tax_calculation_address_type` (String) Specifies whether tax calculation is based on customer billing address or shipping address. Valid values: `BILLING_ADDRESS`, `SHIPPING_ADDRESS`. Defaults to `BILLING_ADDRESS`.
- `decimal_points` (Number) Number of decimal points used in cart calculations. Must be zero or positive. Defaults to `2`.
- `cart_calculation_scale` (Number) Scale for cart calculations. Defaults to `2`.
- `assisted_buying` (Object) Assisted buying configuration. See [Assisted Buying](#nested-schema-for-assisted_buying) below.
- `mixins` (List of Object) Custom mixin configurations. See [Mixins](#nested-schema-for-mixins) below.

### Read-Only

- `id` (String) The ID of this resource (same as `code`).

## Nested Schema for `home_base`

Required:

- `address` (Object) Address configuration. See [Address](#nested-schema-for-home_baseaddress) below.

Optional:

- `location` (Object) Geographic location. See [Location](#nested-schema-for-home_baselocation) below.

### Nested Schema for `home_base.address`

Required:

- `country` (String) Country code (ISO 3166-1 alpha-2).
- `zip_code` (String) ZIP/Postal code.
- `city` (String) City name.

Optional:

- `street` (String) Street name.
- `street_number` (String) Street number.
- `state` (String) State or province.

### Nested Schema for `home_base.location`

Optional:

- `latitude` (Number) Latitude coordinate.
- `longitude` (Number) Longitude coordinate.

## Nested Schema for `assisted_buying`

Optional:

- `storefront_url` (String) URL of the storefront for assisted buying.

## Nested Schema for `mixins`

Each mixin object has the following structure:

Required:

- `name` (String) Unique name for the mixin within the site.
- `schema_url` (String) URL to the JSON schema that defines the mixin's structure.
- `fields` (String) Mixin data as a JSON string. Use `jsonencode()` to convert a map to JSON.

Example:

```hcl
mixins = [
  {
    name       = "customFields"
    schema_url = "https://example.com/schemas/custom_v1.json"
    fields = jsonencode({
      brandColor = "#FF5733"
      theme      = "modern"
      enabled    = true
    })
  }
]
```

## Import

Site settings can be imported using the site code:

```shell
terraform import emporix_sitesettings.example site-code
```

For example:

```shell
terraform import emporix_sitesettings.us_site us-main
```
