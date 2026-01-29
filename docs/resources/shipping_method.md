---
page_title: "emporix_shipping_method Resource - terraform-provider-emporix"
subcategory: "Delivery & Shipping"
description: |-
  Manages shipping methods for a shipping zone.
---

# emporix_shipping_method (Resource)

Manages shipping methods for a shipping zone. Shipping methods define delivery options (e.g., standard, express) with associated costs, rules, and restrictions.

## What are Shipping Methods?

Shipping methods represent the delivery options available to customers for a specific shipping zone. They define:

- **Delivery types** (standard, express, economy, overnight, etc.)
- **Shipping costs** with tier-based pricing
- **Order value requirements** (min/max thresholds)
- **Tax handling** (via tax codes)
- **Availability** (active/inactive status)

## Example Usage

### Basic Shipping Method

First, create a shipping zone:

```terraform
resource "emporix_shipping_zone" "us_domestic" {
  id   = "us-domestic"
  site = "main"

  name = {
    en = "US Domestic"
  }

  ship_to = [
    { country = "US" }
  ]
}
```

Then create the shipping method:

```terraform
resource "emporix_shipping_method" "standard" {
  id      = "standard-shipping"
  site    = "main"
  zone_id = emporix_shipping_zone.us_domestic.id

  name = {
    en = "Standard Shipping"
    de = "Standardversand"
  }

  active = true

  fees = [
    {
      min_order_value = {
        amount   = 0
        currency = "USD"
      }
      cost = {
        amount   = 5.99
        currency = "USD"
      }
    }
  ]

  shipping_tax_code = "SHIPPING-STANDARD"

  depends_on = [emporix_shipping_zone.us_domestic]
}
```

### Tiered Pricing

```terraform
resource "emporix_shipping_method" "express" {
  id      = "express-shipping"
  site    = "main"
  zone_id = emporix_shipping_zone.us_domestic.id

  name = {
    en = "Express Shipping"
  }

  active = true

  fees = [
    # $15 for orders under $50
    {
      min_order_value = {
        amount   = 0
        currency = "USD"
      }
      cost = {
        amount   = 15.00
        currency = "USD"
      }
    },
    # $10 for orders $50-$99.99
    {
      min_order_value = {
        amount   = 50
        currency = "USD"
      }
      cost = {
        amount   = 10.00
        currency = "USD"
      }
    },
    # Free for orders $100+
    {
      min_order_value = {
        amount   = 100
        currency = "USD"
      }
      cost = {
        amount   = 0
        currency = "USD"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.us_domestic]
}
```

### With Maximum Order Value

```terraform
resource "emporix_shipping_method" "economy" {
  id      = "economy-shipping"
  site    = "main"
  zone_id = emporix_shipping_zone.us_domestic.id

  name = {
    en = "Economy Shipping"
  }

  active = true

  # Only available for orders up to $200
  max_order_value = {
    amount   = 200
    currency = "USD"
  }

  fees = [
    {
      min_order_value = {
        amount   = 0
        currency = "USD"
      }
      cost = {
        amount   = 3.99
        currency = "USD"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.us_domestic]
}
```

### International Shipping

```terraform
resource "emporix_shipping_zone" "international" {
  id   = "international"
  site = "main"

  name = {
    en = "International"
  }

  ship_to = [
    { country = "CA" },
    { country = "MX" },
    { country = "GB" }
  ]
}

resource "emporix_shipping_method" "intl_standard" {
  id      = "international-standard"
  site    = "main"
  zone_id = emporix_shipping_zone.international.id

  name = {
    en = "International Standard"
  }

  active = true

  fees = [
    {
      min_order_value = {
        amount   = 0
        currency = "USD"
      }
      cost = {
        amount   = 25.00
        currency = "USD"
      }
    },
    # Free shipping for orders over $150
    {
      min_order_value = {
        amount   = 150
        currency = "USD"
      }
      cost = {
        amount   = 0
        currency = "USD"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.international]
}
```

## Schema

### Required

- `id` (String) Shipping method identifier. Must be unique within the zone. Changing this forces a new resource to be created.
- `site` (String) Site code (typically 'main' for single-shop tenants). Changing this forces a new resource to be created.
- `zone_id` (String) Shipping zone ID this method belongs to. Must reference an existing shipping zone. Changing this forces a new resource to be created.
- `name` (Map of String) Localized names for the shipping method (e.g., {"en": "Standard Shipping", "de": "Standardversand"}).
- `fees` (Block List, Min: 1) Shipping fee tiers based on order value. See [fees](#fees) below.

### Optional

- `active` (Boolean) Whether the shipping method is active. Defaults to `true`.
- `max_order_value` (Block) Maximum order value for this shipping method. Orders above this value cannot use this method. See [max_order_value](#max_order_value) below.
- `shipping_tax_code` (String) Tax code for shipping fees.
- `shipping_group_id` (String) Shipping group ID to associate with this method.

### Nested Schema for `fees`

Required block list. Each fee tier defines pricing for a range of order values.

- `min_order_value` (Block, Required) Minimum order value for this fee tier. See [monetary_amount](#monetary_amount) below.
- `cost` (Block, Required) Shipping cost for this tier. See [monetary_amount](#monetary_amount) below.
- `shipping_group_id` (String, Optional) Optional shipping group ID for this specific fee tier.

### Nested Schema for `max_order_value`

Optional block. When set, orders above this value cannot use this shipping method.

- `amount` (Number, Required) Amount value.
- `currency` (String, Required) Currency code (e.g., 'USD', 'EUR', 'GBP').

### Nested Schema for `monetary_amount`

Used in `min_order_value` and `cost` blocks.

- `amount` (Number, Required) Amount value.
- `currency` (String, Required) Currency code (e.g., 'USD', 'EUR', 'GBP') [Full list of ISO 4217 codes](https://en.wikipedia.org/wiki/ISO_4217).

## Import

Shipping methods can be imported using the format `site:zone_id:method_id`:

```shell
terraform import emporix_shipping_method.standard main:zone-us:standard-shipping
```

Format: `site:zone_id:method_id`

Where:
- `site` - Site code (typically "main")
- `zone_id` - Shipping zone ID
- `method_id` - Shipping method ID

## Required OAuth Scopes

The following OAuth scopes are required:

- `shipping.shipping_manage` - For create, update, and delete operations
- `shipping.shipping_read` - For read operations

## API Documentation

For more information about the Emporix Shipping Methods API, see:
- [Shipping Methods API](https://developer.emporix.io/api-references/api-guides/delivery-and-shipping/shipping-1/api-reference/shipping-methods)