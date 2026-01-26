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
- `currency` (String, Required) Currency code (e.g., 'USD', 'EUR', 'GBP').

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

## API Endpoints

This resource uses the following API endpoints:

- `POST /shipping/{tenant}/{site}/zones/{zoneId}/methods` - Create shipping method
- `GET /shipping/{tenant}/{site}/zones/{zoneId}/methods/{methodId}` - Get shipping method
- `PUT /shipping/{tenant}/{site}/zones/{zoneId}/methods/{methodId}` - Update shipping method
- `DELETE /shipping/{tenant}/{site}/zones/{zoneId}/methods/{methodId}` - Delete shipping method

## Notes

### Concurrency Control

Shipping method operations are protected by per-tenant mutexes to ensure sequential execution. This prevents conflicts when multiple shipping methods are created, updated, or deleted simultaneously.

**Why this is important**:
- The Emporix API requires sequential operations per tenant
- Concurrent operations can cause 409 Conflict errors
- Mutexes ensure operations complete in a predictable order

**What this means for you**:
- Multiple shipping methods in the same `terraform apply` will be created sequentially
- Operations for different tenants can still run in parallel
- No action needed on your part - this is handled automatically by the provider

### Dependencies

Shipping methods require a shipping zone to be created first. Use the `emporix_shipping_zone` resource and reference its ID:

```terraform
resource "emporix_shipping_zone" "my_zone" {
  id   = "my-zone"
  site = "main"
  name = { en = "My Zone" }
  ship_to = [{ country = "US" }]
}

resource "emporix_shipping_method" "my_method" {
  zone_id = emporix_shipping_zone.my_zone.id
  # ... other configuration
  depends_on = [emporix_shipping_zone.my_zone]
}
```

### Fee Tier Logic

Fee tiers are evaluated based on the cart's total order value:

**Example with 3 tiers:**
```terraform
fees = [
  { min_order_value = { amount = 0 },   cost = { amount = 10 } },  # $0-$49.99
  { min_order_value = { amount = 50 },  cost = { amount = 5 } },   # $50-$99.99
  { min_order_value = { amount = 100 }, cost = { amount = 0 } }    # $100+
]
```

**How it works:**
- Cart total: $30 → Shipping cost: $10
- Cart total: $75 → Shipping cost: $5
- Cart total: $150 → Shipping cost: $0 (free)

**Best Practice**: Always start with `min_order_value.amount = 0` to cover all order values.

### Localized Names

Provide names in multiple languages for better customer experience:

```terraform
name = {
  en = "Standard Shipping"
  de = "Standardversand"
  fr = "Livraison Standard"
  es = "Envío Estándar"
  it = "Spedizione Standard"
}
```

### Currency Codes

Use ISO 4217 currency codes:
- USD (US Dollar)
- EUR (Euro)
- GBP (British Pound)
- CAD (Canadian Dollar)
- AUD (Australian Dollar)
- JPY (Japanese Yen)

[Full list of ISO 4217 codes](https://en.wikipedia.org/wiki/ISO_4217)

### Active vs Inactive

Control method availability without deleting:

```terraform
active = false  # Hidden from customers
active = true   # Available for selection
```

Use `active = false` to temporarily disable a shipping method (e.g., during holidays, carrier issues, etc.).

### Maximum Order Value

Restrict expensive shipping methods to smaller orders:

```terraform
# Economy shipping only for orders under $200
max_order_value = {
  amount   = 200
  currency = "USD"
}
```

Use cases:
- **Weight/size limits**: Restrict methods that can't handle large shipments
- **Insurance limits**: Cap order values for basic shipping options
- **Regional restrictions**: Limit expensive items to premium shipping

## Use Cases

### Free Shipping Threshold

Encourage larger orders with free shipping:

```terraform
resource "emporix_shipping_method" "standard" {
  id      = "standard"
  site    = "main"
  zone_id = emporix_shipping_zone.us.id

  name = { en = "Standard Shipping" }

  fees = [
    {
      min_order_value = { amount = 0, currency = "USD" }
      cost = { amount = 9.99, currency = "USD" }
    },
    {
      min_order_value = { amount = 75, currency = "USD" }
      cost = { amount = 0, currency = "USD" }  # Free!
    }
  ]
}
```

### Express with Volume Discount

Incentivize faster shipping for large orders:

```terraform
resource "emporix_shipping_method" "express" {
  id      = "express"
  site    = "main"
  zone_id = emporix_shipping_zone.us.id

  name = { en = "Express Shipping" }

  fees = [
    {
      min_order_value = { amount = 0, currency = "USD" }
      cost = { amount = 19.99, currency = "USD" }
    },
    {
      min_order_value = { amount = 100, currency = "USD" }
      cost = { amount = 14.99, currency = "USD" }  # $5 discount
    },
    {
      min_order_value = { amount = 200, currency = "USD" }
      cost = { amount = 9.99, currency = "USD" }   # $10 discount
    }
  ]
}
```

### Budget Option with Limits

Low-cost shipping for small orders:

```terraform
resource "emporix_shipping_method" "economy" {
  id      = "economy"
  site    = "main"
  zone_id = emporix_shipping_zone.us.id

  name = { en = "Economy Shipping" }

  max_order_value = {
    amount   = 150
    currency = "USD"
  }

  fees = [
    {
      min_order_value = { amount = 0, currency = "USD" }
      cost = { amount = 2.99, currency = "USD" }
    }
  ]
}
```

### International Tiered Pricing

Different rates for international shipping:

```terraform
resource "emporix_shipping_method" "international" {
  id      = "intl-standard"
  site    = "main"
  zone_id = emporix_shipping_zone.international.id

  name = {
    en = "International Standard"
  }

  fees = [
    {
      min_order_value = { amount = 0, currency = "USD" }
      cost = { amount = 30.00, currency = "USD" }
    },
    {
      min_order_value = { amount = 100, currency = "USD" }
      cost = { amount = 20.00, currency = "USD" }
    },
    {
      min_order_value = { amount = 200, currency = "USD" }
      cost = { amount = 0, currency = "USD" }  # Free
    }
  ]
}
```

## Troubleshooting

### Error: zone not found

**Problem**: Shipping zone doesn't exist when creating the method.

**Solution**: Ensure the zone is created first:
```terraform
resource "emporix_shipping_method" "test" {
  zone_id = emporix_shipping_zone.my_zone.id
  # ...
  depends_on = [emporix_shipping_zone.my_zone]
}
```

### Error: method already exists

**Problem**: A method with this ID already exists in the zone.

**Solution**: Use a different ID or import the existing method:
```bash
terraform import emporix_shipping_method.test main:zone-id:existing-method-id
```

### Error: invalid currency

**Problem**: Currency code is not valid.

**Solution**: Use ISO 4217 currency codes (USD, EUR, GBP, etc.).

### Missing tier for order value

**Problem**: No fee tier covers a specific order value range.

**Solution**: Ensure your first tier starts at 0:
```terraform
fees = [
  {
    min_order_value = { amount = 0, currency = "USD" }  # Always start at 0
    cost = { amount = 10, currency = "USD" }
  }
]
```

## Summary

✅ **Dependencies** - Require shipping zone to exist first
✅ **Fee tiers** - Support multiple pricing levels based on order value
✅ **Max order value** - Optional restriction for method availability
✅ **Localization** - Multi-language name support
✅ **Tax codes** - Optional for proper tax calculation
✅ **Active/Inactive** - Control availability without deletion
✅ **Per-tier groups** - Optional shipping group IDs per fee tier

Shipping methods provide flexible delivery options with sophisticated pricing rules tailored to your business needs.
