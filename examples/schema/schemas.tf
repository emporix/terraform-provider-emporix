# Example Terraform configuration for Emporix Schemas

terraform {
  required_providers {
    emporix = {
      source  = "emporix/emporix"
      version = "~> 0.1"
    }
  }
}

# Configure the Emporix provider
# Recommended: Use a Custom API Key with only the required scopes
# Required scope: schema.schema_manage
# See: https://developer.emporix.io/ce/getting-started/developer-portal/manage-apikeys#custom-api-keys
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

# Example 1: Simple schema with text attributes
resource "emporix_schema" "product_simple" {
  id = "product-custom-fields"
  name = {
    en = "Product Custom Fields"
  }
  types = ["PRODUCT"]

  attributes = [
    {
      key = "manufacturer"
      name = {
        en = "Manufacturer"
      }
      description = {
        en = "Product manufacturer name"
      }
      type = "TEXT"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
    },
    {
      key = "warranty_period"
      name = {
        en = "Warranty Period"
      }
      type = "NUMBER"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
    }
  ]
}

# Example 2: Schema with ENUM attribute (predefined values)
resource "emporix_schema" "product_rating" {
  id = "product-rating-schema"
  name = {
    en = "Product Rating Schema"
    de = "Produktbewertungsschema"
  }
  types = ["PRODUCT"]

  attributes = [
    {
      key = "qualityRating"
      name = {
        en = "Quality Rating"
        de = "Qualitätsbewertung"
      }
      type = "ENUM"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
      values = [
        {
          value = "excellent"
        },
        {
          value = "good"
        },
        {
          value = "average"
        },
        {
          value = "poor"
        }
      ]
    }
  ]
}

# Example 3: Schema with multiple attribute types
resource "emporix_schema" "customer_extended" {
  id = "customer-extended-fields"
  name = {
    en = "Customer Extended Fields"
  }
  types = ["CUSTOMER"]

  attributes = [
    {
      key = "preferredLanguage"
      name = {
        en = "Preferred Language"
      }
      type = "TEXT"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
    },
    {
      key = "loyaltyPoints"
      name = {
        en = "Loyalty Points"
      }
      type = "NUMBER"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = false
      }
    },
    {
      key = "isPremiumMember"
      name = {
        en = "Is Premium Member"
      }
      type = "BOOLEAN"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = false
      }
    },
    {
      key = "registrationDate"
      name = {
        en = "Registration Date"
      }
      type = "DATE"
      metadata = {
        read_only  = true
        localized  = false
        required   = false
        nullable   = true
      }
    }
  ]
}

# Example 4: Schema with nested OBJECT attributes
resource "emporix_schema" "product_dimensions" {
  id = "product-dimensions-schema"
  name = {
    en = "Product Dimensions Schema"
  }
  types = ["PRODUCT"]

  attributes = [
    {
      key = "dimensions"
      name = {
        en = "Dimensions"
      }
      type = "OBJECT"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
      attributes = [
        {
          key = "length"
          name = {
            en = "Length"
          }
          type = "DECIMAL"
          metadata = {
            read_only  = false
            localized  = false
            required   = true
            nullable   = false
          }
        },
        {
          key = "width"
          name = {
            en = "Width"
          }
          type = "DECIMAL"
          metadata = {
            read_only  = false
            localized  = false
            required   = true
            nullable   = false
          }
        },
        {
          key = "height"
          name = {
            en = "Height"
          }
          type = "DECIMAL"
          metadata = {
            read_only  = false
            localized  = false
            required   = true
            nullable   = false
          }
        },
        {
          key = "unit"
          name = {
            en = "Unit"
          }
          type = "TEXT"
          metadata = {
            read_only  = false
            localized  = false
            required   = true
            nullable   = false
          }
        }
      ]
    }
  ]
}

# Example 5: Schema with ARRAY type
resource "emporix_schema" "product_tags" {
  id = "product-tags-schema"
  name = {
    en = "Product Tags Schema"
  }
  types = ["PRODUCT"]

  attributes = [
    {
      key = "tags"
      name = {
        en = "Product Tags"
      }
      type = "ARRAY"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
      array_type = {
        type      = "TEXT"
        localized = false
      }
    }
  ]
}

# Example 6: Schema for multiple entity types
resource "emporix_schema" "shared_metadata" {
  id = "shared-metadata-schema"
  name = {
    en = "Shared Metadata Schema"
  }
  types = ["PRODUCT", "CATEGORY"]

  attributes = [
    {
      key = "seoTitle"
      name = {
        en = "SEO Title"
      }
      type = "TEXT"
      metadata = {
        read_only  = false
        localized  = true
        required   = false
        nullable   = true
      }
    },
    {
      key = "seoDescription"
      name = {
        en = "SEO Description"
      }
      type = "TEXT"
      metadata = {
        read_only  = false
        localized  = true
        required   = false
        nullable   = true
      }
    }
  ]
}

# Example 7: Using for_each for multiple similar schemas
locals {
  entity_schemas = {
    order = {
      id    = "order-custom-schema"
      name  = "Order Custom Schema"
      types = ["ORDER"]
    }
    cart = {
      id    = "cart-custom-schema"
      name  = "Cart Custom Schema"
      types = ["CART"]
    }
    quote = {
      id    = "quote-custom-schema"
      name  = "Quote Custom Schema"
      types = ["QUOTE"]
    }
  }
}

resource "emporix_schema" "entities" {
  for_each = local.entity_schemas

  id = each.value.id
  name = {
    en = each.value.name
  }
  types = each.value.types

  attributes = [
    {
      key = "customNote"
      name = {
        en = "Custom Note"
      }
      type = "TEXT"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
    }
  ]
}

# Outputs
output "product_schemas" {
  description = "Product schema IDs"
  value = {
    simple     = emporix_schema.product_simple.id
    rating     = emporix_schema.product_rating.id
    dimensions = emporix_schema.product_dimensions.id
    tags       = emporix_schema.product_tags.id
  }
}

output "customer_schema" {
  description = "Customer schema details"
  value = {
    id    = emporix_schema.customer_extended.id
    name  = lookup(emporix_schema.customer_extended.name, "en", "")
    types = emporix_schema.customer_extended.types
  }
}

output "entity_schemas" {
  description = "Entity schema IDs"
  value       = [for schema in emporix_schema.entities : schema.id]
}

output "all_schema_ids" {
  description = "All schema IDs created"
  value = concat(
    [
      emporix_schema.product_simple.id,
      emporix_schema.product_rating.id,
      emporix_schema.customer_extended.id,
      emporix_schema.product_dimensions.id,
      emporix_schema.product_tags.id,
      emporix_schema.shared_metadata.id
    ],
    [for schema in emporix_schema.entities : schema.id]
  )
}
