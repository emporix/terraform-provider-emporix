terraform {
  required_providers {
    emporix = {
      source = "emporix/emporix"
    }
  }
}

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

# Shipping zones for delivery times
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

# Shipping methods required for delivery time slots
resource "emporix_shipping_method" "standard_downtown" {
  id      = "standard-shipping"
  site    = "main"
  zone_id = emporix_shipping_zone.downtown.id

  name = {
    en = "Standard Shipping"
    pl = "Dostawa Standardowa"
  }

  active = true

  fees = [
    {
      min_order_value = {
        amount   = 0
        currency = "PLN"
      }
      cost = {
        amount   = 15.00
        currency = "PLN"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.downtown]
}

resource "emporix_shipping_method" "express_downtown" {
  id      = "express-shipping"
  site    = "main"
  zone_id = emporix_shipping_zone.downtown.id

  name = {
    en = "Express Shipping"
    pl = "Dostawa Ekspresowa"
  }

  active = true

  fees = [
    {
      min_order_value = {
        amount   = 0
        currency = "PLN"
      }
      cost = {
        amount   = 25.00
        currency = "PLN"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.downtown]
}

resource "emporix_shipping_method" "standard_metro" {
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

# Friday delivery with morning and afternoon slots
resource "emporix_delivery_time" "friday_delivery" {
  name               = "friday-delivery-slots"
  site_code          = "main"
  is_delivery_day    = true
  zone_id            = emporix_shipping_zone.downtown.id
  time_zone_id       = "Europe/Warsaw"
  delivery_day_shift = 0

  day = {
    weekday = "FRIDAY"
  }

  slots = [
    # Morning slot (10:00-12:00) with standard shipping
    {
      shipping_method = emporix_shipping_method.standard_downtown.id
      capacity        = 50

      delivery_time_range = {
        time_from = "10:00:00"
        time_to   = "12:00:00"
      }

      cut_off_time = {
        time                = "2023-06-12T06:00:00.000Z"
        delivery_cycle_name = "morning"
      }
    },
    # Afternoon slot (14:00-16:00) with express shipping
    {
      shipping_method = emporix_shipping_method.express_downtown.id
      capacity        = 30

      delivery_time_range = {
        time_from = "14:00:00"
        time_to   = "16:00:00"
      }

      cut_off_time = {
        time                = "2023-06-12T10:00:00.000Z"
        delivery_cycle_name = "afternoon"
      }
    }
  ]

  depends_on = [
    emporix_shipping_zone.downtown,
    emporix_shipping_method.standard_downtown,
    emporix_shipping_method.express_downtown
  ]
}

# Saturday delivery - single large slot
resource "emporix_delivery_time" "saturday_delivery" {
  name               = "saturday-delivery"
  site_code          = "main"
  is_delivery_day    = true
  zone_id            = emporix_shipping_zone.downtown.id
  time_zone_id       = "Europe/Warsaw"
  delivery_day_shift = 0

  day = {
    weekday = "SATURDAY"
  }

  slots = [
    {
      shipping_method = emporix_shipping_method.standard_downtown.id
      capacity        = 100

      delivery_time_range = {
        time_from = "09:00:00"
        time_to   = "17:00:00"
      }

      cut_off_time = {
        time                = "2023-06-12T20:00:00.000Z"
        delivery_cycle_name = "saturday"
      }
    }
  ]

  depends_on = [
    emporix_shipping_zone.downtown,
    emporix_shipping_method.standard_downtown
  ]
}

# Express delivery - Next day delivery
resource "emporix_delivery_time" "express_next_day" {
  name               = "express-next-day"
  site_code          = "main"
  is_delivery_day    = true
  zone_id            = emporix_shipping_zone.metro.id
  time_zone_id       = "America/New_York"
  delivery_day_shift = 1 # Next day delivery

  day = {
    weekday = "MONDAY"
  }

  slots = [
    {
      shipping_method = emporix_shipping_method.standard_metro.id
      capacity        = 30

      delivery_time_range = {
        time_from = "08:00:00"
        time_to   = "20:00:00"
      }

      cut_off_time = {
        time                = "2023-06-11T15:00:00.000Z"
        delivery_cycle_name = "express"
      }
    }
  ]

  depends_on = [
    emporix_shipping_zone.metro,
    emporix_shipping_method.standard_metro
  ]
}

# Output delivery time names for reference
output "delivery_time_names" {
  description = "Created delivery time configuration names"
  value = {
    friday   = emporix_delivery_time.friday_delivery.name
    saturday = emporix_delivery_time.saturday_delivery.name
    express  = emporix_delivery_time.express_next_day.name
  }
}

output "shipping_zone_ids" {
  description = "Created shipping zone IDs"
  value = {
    downtown = emporix_shipping_zone.downtown.id
    metro    = emporix_shipping_zone.metro.id
  }
}

output "shipping_method_ids" {
  description = "Created shipping method IDs"
  value = {
    standard_downtown = emporix_shipping_method.standard_downtown.id
    express_downtown  = emporix_shipping_method.express_downtown.id
    standard_metro    = emporix_shipping_method.standard_metro.id
  }
}
