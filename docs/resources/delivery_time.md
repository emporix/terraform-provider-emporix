---
page_title: "emporix_delivery_time Resource - terraform-provider-emporix"
subcategory: "Delivery & Shipping"
description: |-
  Manages delivery time slot configurations for scheduled deliveries.
---

# emporix_delivery_time (Resource)

Manages delivery time slot configurations for scheduled deliveries. Delivery times define specific days and time windows when deliveries can occur, with capacity management and cut-off times.

## What are Delivery Times?

Delivery times in Emporix are **scheduled delivery slot configurations** used by businesses that deliver on specific days/times (e.g., grocery delivery, furniture delivery). They are not simple delivery duration estimates (like "3-5 days").

**Key Features:**
- Specific weekdays for delivery (e.g., Friday only)
- Time slots with windows (e.g., 10:00-12:00, 14:00-16:00)
- Capacity limits per slot
- Order cut-off times
- Zone-specific or all-zones configuration
- Shipping method associations

## Example Usage

### Basic Delivery Time with Single Slot

First, create a shipping zone:

```terraform
resource "emporix_shipping_zone" "downtown" {
  id   = "zone-downtown"
  site = "main"
  name = {
    en = "Downtown Zone"
    pl = "Strefa Śródmieście"
  }

  ship_to = [
    { country = "PL" }
  ]
}
```

Then create the delivery time:

```terraform
resource "emporix_delivery_time" "friday_morning" {
  name            = "friday-morning-delivery"
  site_code       = "main"
  is_delivery_day = true
  zone_id         = emporix_shipping_zone.downtown.id
  time_zone_id    = "Europe/Warsaw"

  day = {
    weekday = "FRIDAY"
  }

  slots = [
    {
      shipping_method = "standard"
      capacity        = 50

      delivery_time_range = {
        time_from = "10:00:00"
        time_to   = "12:00:00"
      }

      cut_off_time = {
        time                = "2023-06-12T06:00:00.000Z"
        delivery_cycle_name = "morning"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.downtown]
}
```

### Delivery Time with Multiple Slots

```terraform
resource "emporix_shipping_zone" "central" {
  id   = "zone-central"
  site = "main"
  name = {
    en = "Central Zone"
    pl = "Strefa Centralna"
  }

  ship_to = [
    { country = "PL" }
  ]
}

resource "emporix_delivery_time" "saturday_delivery" {
  name            = "saturday-slots"
  site_code       = "main"
  zone_id         = emporix_shipping_zone.central.id
  time_zone_id    = "Europe/Warsaw"
  is_delivery_day = true

  day = {
    weekday = "SATURDAY"
  }

  slots = [
    # Morning slot
    {
      shipping_method = "standard"
      capacity        = 50

      delivery_time_range = {
        time_from = "09:00:00"
        time_to   = "12:00:00"
      }

      cut_off_time = {
        time                = "2023-06-13T06:00:00.000Z"
        delivery_cycle_name = "morning"
      }
    },
    # Afternoon slot
    {
      shipping_method = "standard"
      capacity        = 30

      delivery_time_range = {
        time_from = "14:00:00"
        time_to   = "17:00:00"
      }

      cut_off_time = {
        time                = "2023-06-13T11:00:00.000Z"
        delivery_cycle_name = "afternoon"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.central]
}
```

### All Zones Configuration

```terraform
resource "emporix_delivery_time" "sunday_all_zones" {
  name             = "sunday-delivery"
  site_code        = "main"
  is_delivery_day  = true
  is_for_all_zones = true
  time_zone_id     = "America/New_York"

  day = {
    weekday = "SUNDAY"
  }

  slots = [
    {
      shipping_method = "standard"
      capacity        = 100

      delivery_time_range = {
        time_from = "10:00:00"
        time_to   = "18:00:00"
      }
    }
  ]
}
```

### Next-Day Delivery with Shift

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

resource "emporix_delivery_time" "next_day_express" {
  name               = "next-day-express"
  site_code          = "main"
  zone_id            = emporix_shipping_zone.metro.id
  time_zone_id       = "Europe/London"
  is_delivery_day    = true
  delivery_day_shift = 1 # Next day delivery

  day = {
    weekday = "MONDAY"
  }

  slots = [
    {
      shipping_method = "express"
      capacity        = 25

      delivery_time_range = {
        time_from = "08:00:00"
        time_to   = "20:00:00"
      }

      cut_off_time = {
        time                = "2023-06-11T15:00:00.000Z"
        delivery_cycle_name = "next-day"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.metro]
}
```

## Schema

### Required

- `name` (String) Unique name for the delivery time configuration. Changing this forces a new resource to be created.
- `site_code` (String) Site code. Typically 'main' for single-shop tenants.
- `time_zone_id` (String) Timezone identifier (e.g., 'Europe/Warsaw', 'America/New_York'). Must be a valid IANA timezone.

### Optional

- `is_delivery_day` (Boolean) Whether this is an active delivery day. Defaults to `true`.
- `zone_id` (String) Shipping zone ID this delivery time applies to. Required if `is_for_all_zones` is false. Should reference the ID of an `emporix_shipping_zone` resource.
- `is_for_all_zones` (Boolean) Whether this delivery time applies to all zones. Defaults to `false`. Cannot be used together with `zone_id`.
- `delivery_day_shift` (Number) Number of days to shift delivery from order date. Defaults to `0`. Use `1` for next-day delivery, `2` for two-day delivery, etc.
- `day` (Block) Day configuration for this delivery time. See [day](#day) below.
- `slots` (Block List) Delivery time slots with shipping methods and capacity. See [slots](#slots) below.

### Read-Only

- `id` (String) Unique identifier generated by the API.

### Nested Schema for `day`

Optional single nested block.

- `weekday` (String, Required) Day of the week. Must be one of: `MONDAY`, `TUESDAY`, `WEDNESDAY`, `THURSDAY`, `FRIDAY`, `SATURDAY`, `SUNDAY`.

### Nested Schema for `slots`

Optional list of nested blocks.

- `shipping_method` (String, Required) Shipping method identifier associated with this slot.
- `capacity` (Number, Required) Maximum number of deliveries allowed for this slot.
- `delivery_time_range` (Block, Required) Time range for delivery. See [delivery_time_range](#delivery_time_range) below.
- `cut_off_time` (Block, Optional) Order cutoff time for this slot. See [cut_off_time](#cut_off_time) below.

### Nested Schema for `delivery_time_range`

Required nested block within `slots`.

- `time_from` (String, Required) Start time in HH:MM:SS format (e.g., '10:00:00'). Must match pattern `HH:MM:SS` where HH is 00-23, MM is 00-59, SS is 00-59.
- `time_to` (String, Required) End time in HH:MM:SS format (e.g., '12:00:00'). Must match pattern `HH:MM:SS` where HH is 00-23, MM is 00-59, SS is 00-59.

### Nested Schema for `cut_off_time`

Optional nested block within `slots`.

- `time` (String, Required) Cutoff timestamp in ISO 8601 format (e.g., '2023-06-12T18:00:00.000Z').
- `delivery_cycle_name` (String, Required) Delivery cycle identifier (e.g., 'morning', 'afternoon', 'express').

## Import

Delivery times can be imported using their ID (generated by the API):

```shell
terraform import emporix_delivery_time.example abc123def456
```

Format: `id`

**Note**: The ID is generated by the Emporix API and returned in the response when creating a delivery time. You can find it in your Terraform state file or by querying the API directly.

## Required OAuth Scopes

The following OAuth scopes are required:

- `delivery.delivery_manage` - For create, update, and delete operations
- `delivery.delivery_read` - For read operations

## API Documentation

For more information about the Emporix Delivery Times API, see:
- [Delivery Times Management API](https://developer.emporix.io/api-references/api-guides/delivery-and-shipping/shipping-1/api-reference/delivery-times-management)

## API Endpoints

This resource uses the following API endpoints:

- `POST /shipping/{tenant}/delivery-times` - Create delivery time (returns ID in response)
- `GET /shipping/{tenant}/delivery-times/{id}` - Get delivery time (uses API-generated ID)
- `PUT /shipping/{tenant}/delivery-times/{id}` - Update delivery time (uses API-generated ID)
- `DELETE /shipping/{tenant}/delivery-times/{id}` - Delete delivery time (uses API-generated ID)

**Note**: The `id` in the URL path is the API-generated identifier returned when creating the delivery time, not the user-defined `name` field.

## Notes

### Dependencies

Delivery times require a shipping zone to be created first (unless using `is_for_all_zones = true`). Use the `emporix_shipping_zone` resource and reference its ID:

```terraform
resource "emporix_shipping_zone" "my_zone" {
  id   = "my-zone"
  site = "main"
  name = {
    en = "My Zone"
  }

  ship_to = [
    { country = "PL" }
  ]
}

resource "emporix_delivery_time" "my_time" {
  zone_id = emporix_shipping_zone.my_zone.id
  # ... other configuration
  depends_on = [emporix_shipping_zone.my_zone]
}
```

### ID vs Name

**ID (API-Generated)**:
- The `id` field is automatically generated by the Emporix API when you create a delivery time
- This ID is used internally by the API for all operations (GET, PUT, DELETE)
- It's stored in your Terraform state file automatically
- You'll need this ID for importing existing resources

**Name (User-Defined)**:
- The `name` field is a human-readable identifier that you choose
- It must be unique across all delivery times
- Changing the name forces a new resource to be created (because it's part of the resource definition)
- The name is included in the API request body but the ID is used in the URL path

### Name as Identifier

The `name` field serves as the unique identifier you specify when creating the delivery time. It must be unique across all delivery times in the tenant and cannot be changed without recreating the resource.

### Weekday Values

The `weekday` field accepts these exact values (case-sensitive):
- MONDAY
- TUESDAY
- WEDNESDAY
- THURSDAY
- FRIDAY
- SATURDAY
- SUNDAY

### Cut-off Time Validation

**IMPORTANT**: The cut-off time must be BEFORE the delivery window starts (`time_from`), not after it ends.

**Valid Example**:
```terraform
delivery_time_range = {
  time_from = "10:00:00"  # Delivery starts at 10:00
  time_to   = "12:00:00"  # Delivery ends at 12:00
}

cut_off_time = {
  time = "2023-06-12T06:00:00.000Z"  # ✅ Cut-off at 06:00 (before 10:00)
}
```

**Invalid Example**:
```terraform
delivery_time_range = {
  time_from = "10:00:00"
  time_to   = "12:00:00"
}

cut_off_time = {
  time = "2023-06-12T18:00:00.000Z"  # ❌ ERROR: 18:00 is after 12:00
}
```

**Error Message**:
```
CutOff time cannot be after timeTo
```

**Guidelines**:
- Morning slot (10:00-12:00): Set cut-off to early morning (e.g., 06:00) same day or previous evening
- Afternoon slot (14:00-16:00): Set cut-off to late morning (e.g., 11:00) same day
- All-day slot (09:00-17:00): Set cut-off to previous evening (e.g., 20:00 day before)
- Next-day delivery: Set cut-off to afternoon/evening of the previous day (e.g., 15:00 day before)

### Time Formats

**Delivery Time Range**: Use 24-hour HH:MM:SS format with validation
- Valid: "10:00:00", "14:30:00", "18:00:00"
- Invalid: "10:00" (missing seconds), "10am", "2:30PM"

**Cut-off Time**: Use ISO 8601 format with timezone
- Valid: "2023-06-12T18:00:00.000Z"
- Valid: "2023-06-12T15:00:00-05:00"

### Capacity Management

The `capacity` field limits how many deliveries can be scheduled for a slot. When capacity is reached, customers cannot select that slot for delivery.

### Delivery Day Shift

Controls when delivery occurs relative to order date:
- `0` = Delivery on the configured weekday
- `1` = Next day delivery
- `2` = Two days later, etc.

Example: If `weekday = "FRIDAY"` and `delivery_day_shift = 1`, orders placed Monday-Thursday will be delivered on Friday (next occurrence), but orders on Friday will be delivered Saturday.

### Zone Configuration

Two mutually exclusive options:

**Zone-Specific**:
```terraform
zone_id          = emporix_shipping_zone.my_zone.id
is_for_all_zones = false  # or omit (defaults to false)
```

**All Zones**:
```terraform
is_for_all_zones = true
# zone_id is not used
```

### Multiple Slots

You can define multiple time slots for a single delivery day:

```terraform
slots = [
  {
    shipping_method = "standard"
    capacity        = 50
    delivery_time_range = {
      time_from = "10:00:00"
      time_to   = "12:00:00"
    }
  },
  {
    shipping_method = "express"
    capacity        = 25
    delivery_time_range = {
      time_from = "14:00:00"
      time_to   = "16:00:00"
    }
  }
]
```

Each slot can have different:
- Shipping methods
- Capacity limits
- Time ranges
- Cut-off times

### Timezone Considerations

Use proper IANA timezone identifiers:
- ✅ "Europe/Warsaw"
- ✅ "America/New_York"
- ✅ "Asia/Tokyo"
- ❌ "EST" (ambiguous)
- ❌ "GMT+1" (not IANA format)

The timezone affects how delivery times and cut-off times are interpreted.

## Use Cases

### Grocery Delivery Service

Schedule Friday and Saturday deliveries with morning and afternoon slots:

```terraform
resource "emporix_shipping_zone" "city" {
  id   = "city-center"
  site = "main"
  name = {
    en = "City Center"
  }

  ship_to = [
    { country = "PL" }
  ]
}

resource "emporix_delivery_time" "grocery_friday" {
  name      = "grocery-friday"
  site_code = "main"
  zone_id   = emporix_shipping_zone.city.id

  day = {
    weekday = "FRIDAY"
  }

  slots = [
    {
      # Morning slot with higher capacity
      shipping_method = "standard"
      capacity        = 100
      delivery_time_range = {
        time_from = "10:00:00"
        time_to   = "12:00:00"
      }
    },
    {
      # Afternoon slot
      shipping_method = "standard"
      capacity        = 75
      delivery_time_range = {
        time_from = "14:00:00"
        time_to   = "16:00:00"
      }
    }
  ]
}
```

### Furniture Delivery

Limited capacity, specific day, long time windows:

```terraform
resource "emporix_shipping_zone" "metro" {
  id   = "metro-area"
  site = "main"
  name = {
    en = "Metro Area"
  }

  ship_to = [
    { country = "US" }
  ]
}

resource "emporix_delivery_time" "furniture_saturday" {
  name      = "furniture-delivery"
  site_code = "main"
  zone_id   = emporix_shipping_zone.metro.id

  day = {
    weekday = "SATURDAY"
  }

  slots = [
    {
      shipping_method = "white-glove"
      capacity        = 10 # Limited capacity
      delivery_time_range = {
        time_from = "09:00:00"
        time_to   = "17:00:00" # All-day window
      }
    }
  ]
}
```

### Express Next-Day

Next-day delivery with early cut-off:

```terraform
resource "emporix_delivery_time" "express" {
  name               = "express-next-day"
  site_code          = "main"
  is_for_all_zones   = true
  delivery_day_shift = 1 # Next day

  day = {
    weekday = "MONDAY"
  }

  slots = [
    {
      shipping_method = "express"
      capacity        = 50
      delivery_time_range = {
        time_from = "08:00:00"
        time_to   = "20:00:00"
      }
      cut_off_time = {
        time                = "2023-06-12T15:00:00.000Z"
        delivery_cycle_name = "express"
      }
    }
  ]
}
```
