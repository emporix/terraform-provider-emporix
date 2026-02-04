---
page_title: "emporix_delivery_time Resource - terraform-provider-emporix"
subcategory: ""
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
        time_from = "10:00"
        time_to   = "14:00"
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
        time_from = "09:00"
        time_to   = "12:00"
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
        time_from = "13:00"
        time_to   = "17:00"
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
        time_from = "08:00"
        time_to   = "20:00"
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

### Specific Date Delivery (One-Time Event)

For special dates like holidays or events:

```terraform
resource "emporix_delivery_time" "christmas_delivery" {
  name               = "christmas-2024"
  site_code          = "main"
  is_delivery_day    = true
  zone_id            = emporix_shipping_zone.downtown.id
  time_zone_id       = "Europe/Warsaw"
  delivery_day_shift = 0

  day = {
    date = "2024-12-25T11:00:00.000Z"  # In Europe/Warsaw it is 12:00 at this date
  }

  slots = [
    {
      shipping_method = emporix_shipping_method.express.id
      capacity        = 20

      delivery_time_range = {
        time_from = "09:00"
        time_to   = "12:00"
      }

      cut_off_time = {
        time                = "2024-12-24T18:00:00.000Z"
        delivery_cycle_name = "christmas"
      }
    }
  ]

  depends_on = [
    emporix_shipping_zone.downtown,
    emporix_shipping_method.express
  ]
}
```

### Date Range Delivery (Temporary Schedule)

For temporary delivery schedules like summer hours:

```terraform
resource "emporix_delivery_time" "summer_extended" {
  name               = "summer-extended-hours"
  site_code          = "main"
  is_delivery_day    = true
  zone_id            = emporix_shipping_zone.downtown.id
  time_zone_id       = "Europe/Warsaw"
  delivery_day_shift = 0

  day = {
    date_from = "2024-06-01T10:00:00.000Z"  # In Europe/Warsaw it is 12:00 at this date
    date_to   = "2024-08-31T10:00:00.000Z"  # In Europe/Warsaw it is 12:00 at this date
  }

  slots = [
    {
      shipping_method = emporix_shipping_method.standard.id
      capacity        = 100

      delivery_time_range = {
        time_from = "08:00"
        time_to   = "20:00"  # Extended hours during summer
      }

      cut_off_time = {
        time                = "2024-06-01T06:00:00.000Z"
        delivery_cycle_name = "summer"
      }
    }
  ]

  depends_on = [
    emporix_shipping_zone.downtown,
    emporix_shipping_method.standard
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

**IMPORTANT**: The API requires ONLY ONE scheduling scenario per delivery time. You must use one of:
- `weekday` alone (recurring weekly)
- `date` alone (one-time)
- `date_from` + `date_to` together (temporary period)

**You cannot combine these options.** If multiple fields are provided, the provider uses this priority: date > date range > weekday.

Optional:

- `weekday` (String) Day of the week for recurring weekly delivery. Valid values: MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY, SATURDAY, SUNDAY.
- `date` (String) Specific date for one-time delivery in ISO 8601 format. The time component must be `12:00:00.000Z` (according to time_zone_id). Example: "2024-12-25T12:00:00.000Z"
- `date_from` (String) Start date for delivery date range in ISO 8601 format. The time component must be `12:00:00.000Z` (according to time_zone_id). Example: "2024-06-01T12:00:00.000Z". Use with `date_to` for temporary schedules.
- `date_to` (String) End date for delivery date range in ISO 8601 format. The time component must be `12:00:00.000Z` (according to time_zone_id). Example: "2024-08-31T12:00:00.000Z". Use with `date_from` for temporary schedules.

**Valid Scenarios**:

| Field(s) | Scenario | Example |
|----------|----------|---------|
| `weekday` only | Recurring weekly delivery | Every Monday |
| `date` only | One-time special delivery | Christmas Day, Black Friday |
| `date_from` + `date_to` | Temporary schedule | Summer hours, holiday season |

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

- `time_from` (String) Start time in HH:MM format (e.g., "10:00")
- `time_to` (String) End time in HH:MM format (e.g., "14:00")

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