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

# Create a shipping zone first
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

# Standard shipping with single fee
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

# Express shipping with tiered pricing
resource "emporix_shipping_method" "express" {
  id      = "express-shipping"
  site    = "main"
  zone_id = emporix_shipping_zone.us_domestic.id

  name = {
    en = "Express Shipping"
    de = "Expressversand"
  }

  active = true

  fees = [
    # Base rate for orders under $50
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
    # Discounted rate for orders $50-$99.99
    {
      min_order_value = {
        amount   = 50
        currency = "USD"
      }
      cost = {
        amount   = 12.00
        currency = "USD"
      }
    },
    # Free express shipping for orders $100+
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

  shipping_tax_code = "SHIPPING-EXPRESS"

  depends_on = [emporix_shipping_zone.us_domestic]
}

# Economy shipping with max order value restriction
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

# International zone with different shipping methods
resource "emporix_shipping_zone" "international" {
  id   = "international"
  site = "main"

  name = {
    en = "International"
    de = "International"
  }

  ship_to = [
    { country = "CA" },
    { country = "MX" },
    { country = "GB" },
    { country = "DE" },
    { country = "FR" }
  ]
}

# International standard shipping
resource "emporix_shipping_method" "intl_standard" {
  id      = "international-standard"
  site    = "main"
  zone_id = emporix_shipping_zone.international.id

  name = {
    en = "International Standard"
    de = "International Standard"
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

# Outputs
output "shipping_methods" {
  description = "Created shipping methods"
  value = {
    standard     = emporix_shipping_method.standard.id
    express      = emporix_shipping_method.express.id
    economy      = emporix_shipping_method.economy.id
    intl_standard = emporix_shipping_method.intl_standard.id
  }
}

output "shipping_zones" {
  description = "Created shipping zones"
  value = {
    us_domestic   = emporix_shipping_zone.us_domestic.id
    international = emporix_shipping_zone.international.id
  }
}
