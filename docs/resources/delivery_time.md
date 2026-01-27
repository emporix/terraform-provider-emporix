---
page_title: "emporix_delivery_time Resource - terraform-provider-emporix"
subcategory: "Delivery & Shipping"
description: |-
  Manages delivery time configurations for shipping zones.
---

# emporix_delivery_time (Resource)

Manages delivery time configurations that define when deliveries can be made for specific shipping zones. Delivery times consist of time slots with capacity limits, shipping methods, and cut-off times.

## Prerequisites

Before creating delivery times, you must have:
1. **Shipping Zone** - The zone where deliveries will be made
2. **Shipping Method** - The method(s) that will be used for delivery

## Example Usage

### Basic Delivery Time with Single Slot

```terraform
# First, create a shipping zone
resource "emporix_shipping_zone" "downtown" {
  id   = "zone-downtown"
  site = "main"

  name = {
    en = "Downtown Zone"
  }

  ship_to = [
    { country = "US" }
  ]
}

# Second, create a shipping method
resource "emporix_shipping_method" "standard" {
  id      = "standard-shipping"
  site    = "main"
  zone_id = emporix_shipping_zone.downtown.id

  name = {
    en = "Standard Shipping"
  }

  active = true

  fees = [
    {
      min_order_value = {
        amount   = 0
        currency = "USD"
      }
      cost = {
        amount   = 9.99
        currency = "USD"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.downtown]
}

# Finally, create the delivery time
resource "emporix_delivery_time" "friday" {
  name               = "friday-delivery"
  site_code          = "main"
  is_delivery_day    = true
  zone_id            = emporix_shipping_zone.downtown.id
  time_zone_id       = "America/New_York"
  delivery_day_shift = 0

  day = {
    weekday = "FRIDAY"
  }

  slots = [
    {
      shipping_method = emporix_shipping_method.standard.id
      capacity        = 50

      delivery_time_range = {
        time_from = "10:00:00"
        time_to   = "14:00:00"
      }

      cut_off_time = {
        time                = "2023-06-12T06:00:00.000Z"
        delivery_cycle_name = "morning"
      }
    }
  ]

  depends_on = [
    emporix_shipping_zone.downtown,
    emporix_shipping_method.standard
  ]
}
```

### Multiple Slots with Different Shipping Methods

```terraform
resource "emporix_shipping_zone" "metro" {
  id   = "zone-metro"
  site = "main"

  name = {
    en = "Metro Area"
  }

  ship_to = [
    { country = "US" }
  ]
}

resource "emporix_shipping_method" "standard" {
  id      = "standard-shipping"
  site    = "main"
  zone_id = emporix_shipping_zone.metro.id

  name = {
    en = "Standard Shipping"
  }

  active = true

  fees = [
    {
      min_order_value = {
        amount   = 0
        currency = "USD"
      }
      cost = {
        amount   = 9.99
        currency = "USD"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.metro]
}

resource "emporix_shipping_method" "express" {
  id      = "express-shipping"
  site    = "main"
  zone_id = emporix_shipping_zone.metro.id

  name = {
    en = "Express Shipping"
  }

  active = true

  fees = [
    {
      min_order_value = {
        amount   = 0
        currency = "USD"
      }
      cost = {
        amount   = 19.99
        currency = "USD"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.metro]
}

resource "emporix_delivery_time" "saturday" {
  name               = "saturday-delivery"
  site_code          = "main"
  is_delivery_day    = true
  zone_id            = emporix_shipping_zone.metro.id
  time_zone_id       = "America/New_York"
  delivery_day_shift = 0

  day = {
    weekday = "SATURDAY"
  }

  slots = [
    # Morning slot with standard shipping
    {
      shipping_method = emporix_shipping_method.standard.id
      capacity        = 50

      delivery_time_range = {
        time_from = "09:00:00"
        time_to   = "12:00:00"
      }

      cut_off_time = {
        time                = "2023-06-13T05:00:00.000Z"
        delivery_cycle_name = "morning"
      }
    },
    # Afternoon slot with express shipping
    {
      shipping_method = emporix_shipping_method.express.id
      capacity        = 30

      delivery_time_range = {
        time_from = "13:00:00"
        time_to   = "17:00:00"
      }

      cut_off_time = {
        time                = "2023-06-13T09:00:00.000Z"
        delivery_cycle_name = "afternoon"
      }
    }
  ]

  depends_on = [
    emporix_shipping_zone.metro,
    emporix_shipping_method.standard,
    emporix_shipping_method.express
  ]
}
```

### Next-Day Delivery

```terraform
# Next-day delivery example
resource "emporix_shipping_zone" "default" {
  id   = "zone-default"
  site = "main"

  name = {
    en = "Default Zone"
  }

  ship_to = [
    { country = "US" }
  ]
}

resource "emporix_shipping_method" "overnight" {
  id      = "overnight-shipping"
  site    = "main"
  zone_id = emporix_shipping_zone.default.id

  name = {
    en = "Overnight Shipping"
  }

  active = true

  fees = [
    {
      min_order_value = {
        amount   = 0
        currency = "USD"
      }
      cost = {
        amount   = 29.99
        currency = "USD"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.default]
}

resource "emporix_delivery_time" "weekday_overnight" {
  name               = "weekday-overnight"
  site_code          = "main"
  is_delivery_day    = true
  zone_id            = emporix_shipping_zone.default.id
  time_zone_id       = "America/New_York"
  delivery_day_shift = 1  # Next day delivery

  day = {
    weekday = "MONDAY"
  }

  slots = [
    {
      shipping_method = emporix_shipping_method.overnight.id
      capacity        = 100

      delivery_time_range = {
        time_from = "08:00:00"
        time_to   = "20:00:00"
      }

      cut_off_time = {
        time                = "2023-06-11T15:00:00.000Z"
        delivery_cycle_name = "overnight"
      }
    }
  ]

  depends_on = [
    emporix_shipping_zone.default,
    emporix_shipping_method.overnight
  ]
}
```

## Schema

### Required

- `name` (String) Delivery time configuration name
- `site_code` (String) Site code (typically 'main')
- `is_delivery_day` (Boolean) Whether this is a delivery day
- `time_zone_id` (String) Time zone ID (e.g., 'Europe/Warsaw', 'America/New_York')
- `day` (Block) Day configuration (see [day](#day) below)
- `slots` (Block List, Min: 1) Delivery time slots (see [slots](#slots) below)

### Optional

- `zone_id` (String) Shipping zone ID. Required unless `is_for_all_zones` is true.
- `is_for_all_zones` (Boolean) Whether this applies to all zones. Defaults to false.
- `delivery_day_shift` (Number) Number of days to shift delivery. Defaults to 0.

### Read-Only

- `id` (String) The ID of this resource (computed by API)

### Nested Schema for `day`

Required:

- `weekday` (String) Day of the week. Valid values: MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY, SATURDAY, SUNDAY.

### Nested Schema for `slots`

Required block list. Each slot defines a delivery window.

Required:

- `shipping_method` (String) **Shipping method ID**. Must reference an existing `emporix_shipping_method` resource.
- `capacity` (Number) Number of deliveries that can be made in this slot
- `delivery_time_range` (Block) Time range for delivery (see below)

Optional:

- `cut_off_time` (Block) Cut-off time for orders (see below)

#### Nested Schema for `delivery_time_range`

Required:

- `time_from` (String) Start time in HH:MM:SS format (e.g., "10:00:00")
- `time_to` (String) End time in HH:MM:SS format (e.g., "14:00:00")

#### Nested Schema for `cut_off_time`

Optional:

- `time` (String) ISO 8601 datetime (e.g., "2023-06-12T06:00:00.000Z")
- `delivery_cycle_name` (String) Name of the delivery cycle

## Import

Delivery times can be imported using their ID:

```shell
terraform import emporix_delivery_time.friday abc123
```

Format: `id`

## Required OAuth Scopes

- `shipping.shipping_manage` - For create, update, and delete operations
- `shipping.shipping_read` - For read operations

## Dependencies

### Critical: Shipping Zone and Shipping Method Required

Delivery times **require** both a shipping zone and at least one shipping method to exist first:

```terraform
# 1. Create shipping zone
resource "emporix_shipping_zone" "zone" {
  id   = "my-zone"
  site = "main"
  # ...
}

# 2. Create shipping method for that zone
resource "emporix_shipping_method" "method" {
  id      = "my-method"
  site    = "main"
  zone_id = emporix_shipping_zone.zone.id  # References zone
  # ...
}

# 3. Create delivery time referencing both
resource "emporix_delivery_time" "delivery" {
  zone_id = emporix_shipping_zone.zone.id  # References zone

  slots = [
    {
      shipping_method = emporix_shipping_method.method.id  # References method
      # ...
    }
  ]

  depends_on = [
    emporix_shipping_zone.zone,
    emporix_shipping_method.method
  ]
}
```

**Why this matters**:
- Shipping methods must be created before delivery times
- Each slot references a specific shipping method by ID
- Without proper dependencies, Terraform may try to create resources in the wrong order

## Important Notes

### Time Format

All times in `delivery_time_range` must use **HH:MM:SS** format:

✅ Correct: `"10:00:00"`
❌ Wrong: `"10:00"`, `"10am"`, `"10:00:00.000Z"`

### Cut-off Time Logic

**CRITICAL**: Cut-off times must be **BEFORE** the delivery window starts, not after.

Example for morning delivery (10:00-12:00):
- ✅ Cut-off: `06:00` same day (4 hours before delivery)
- ❌ Cut-off: `18:00` same day (after delivery - WRONG!)

### Zone ID vs All Zones

Choose one approach (these are mutually exclusive):

**Option 1: Specific Zone** (most common and REQUIRED for delivery days)
```terraform
resource "emporix_delivery_time" "friday" {
  zone_id = emporix_shipping_zone.downtown.id
  # Do NOT set is_for_all_zones (defaults to false)

  day = {
    weekday = "FRIDAY"
  }
  # ...
}
```

**Option 2: All Zones** (ONLY for non-day-specific delivery times)
```terraform
resource "emporix_delivery_time" "general" {
  is_for_all_zones = true
  # Do NOT set zone_id
  # Do NOT set day (no specific weekday)
  # ...
}
```

**⚠️ CRITICAL RESTRICTIONS**:

1. **Never set both `zone_id` and `is_for_all_zones = true` together:**
```
'zoneId' cannot be provided when 'isForAllZones' flag is active.
```

2. **CANNOT use `is_for_all_zones = true` with delivery days:**
```
'isForAllZones' flag cannot be active for delivery days.
errorCode: DELIVERY-TIME-400013
```

**Why?** Delivery days are zone-specific because different zones may have different delivery schedules. Each zone needs its own delivery day configuration.

**Best Practice**:
- For delivery times with specific weekdays (FRIDAY, SATURDAY, etc.): **MUST use `zone_id`**
- For general delivery times without weekdays: Can use `is_for_all_zones`
- Always omit the option you're not using (don't set it to false)

### Weekday Values

Valid weekday values (case-sensitive):
- MONDAY
- TUESDAY
- WEDNESDAY
- THURSDAY
- FRIDAY
- SATURDAY
- SUNDAY