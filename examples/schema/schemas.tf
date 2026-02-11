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

# Example 5: Schema with deeply nested OBJECT attributes (OBJECT within OBJECT)
resource "emporix_schema" "customer_address" {
  id = "customer-address-schema"
  name = {
    en = "Customer Address Schema"
  }
  types = ["CUSTOMER"]

  attributes = [
    {
      key = "primaryAddress"
      name = {
        en = "Primary Address"
      }
      description = {
        en = "Customer's primary address with nested structure"
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
          key = "street"
          name = {
            en = "Street"
          }
          type = "TEXT"
          metadata = {
            read_only  = false
            localized  = false
            required   = true
            nullable   = false
          }
        },
        {
          key = "city"
          name = {
            en = "City"
          }
          type = "TEXT"
          metadata = {
            read_only  = false
            localized  = false
            required   = true
            nullable   = false
          }
        },
        {
          # Nested OBJECT within OBJECT - GPS coordinates
          key = "coordinates"
          name = {
            en = "GPS Coordinates"
          }
          description = {
            en = "Geographic coordinates for the address"
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
              key = "latitude"
              name = {
                en = "Latitude"
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
              key = "longitude"
              name = {
                en = "Longitude"
              }
              type = "DECIMAL"
              metadata = {
                read_only  = false
                localized  = false
                required   = true
                nullable   = false
              }
            }
          ]
        },
        {
          # Nested ENUM within OBJECT
          key = "addressType"
          name = {
            en = "Address Type"
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
              value = "home"
            },
            {
              value = "work"
            },
            {
              value = "billing"
            },
            {
              value = "shipping"
            }
          ]
        },
        {
          # Nested ARRAY within OBJECT
          key = "phoneNumbers"
          name = {
            en = "Phone Numbers"
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
  ]
}

# Example 6: Schema with 4 levels of OBJECT nesting
resource "emporix_schema" "organization_hierarchy" {
  id = "organization-hierarchy-schema"
  name = {
    en = "Organization Hierarchy Schema"
  }
  types = ["CUSTOM_ENTITY"]

  attributes = [
    {
      # Level 1: Organization
      key = "organization"
      name = {
        en = "Organization"
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
          key = "orgName"
          name = {
            en = "Organization Name"
          }
          type = "TEXT"
          metadata = {
            read_only  = false
            localized  = false
            required   = true
            nullable   = false
          }
        },
        {
          # Level 2: Department
          key = "department"
          name = {
            en = "Department"
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
              key = "deptName"
              name = {
                en = "Department Name"
              }
              type = "TEXT"
              metadata = {
                read_only  = false
                localized  = false
                required   = true
                nullable   = false
              }
            },
            {
              # Level 3: Team
              key = "team"
              name = {
                en = "Team"
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
                  key = "teamName"
                  name = {
                    en = "Team Name"
                  }
                  type = "TEXT"
                  metadata = {
                    read_only  = false
                    localized  = false
                    required   = true
                    nullable   = false
                  }
                },
                {
                  # Level 4: memberName (deepest level - basic types only)
                  key = "memberName"
                  name = {
                    en = "Member Name"
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
                  # Level 4: memberRole (deepest level - basic types only)
                  key = "memberRole"
                  name = {
                    en = "Member Role"
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
          ]
        }
      ]
    }
  ]
}

# Example 7: Schema with ARRAY type
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

# Example 8: Schema for multiple entity types
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

# Example 9: Using for_each for multiple similar schemas
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

# Example 10: Schema without explicit ID (auto-generated by API)
# When no ID is provided, the Emporix API will automatically generate one
resource "emporix_schema" "auto_generated_id" {
  # Note: 'id' is not specified - it will be auto-generated by the API
  name = {
    en = "Auto Generated ID Schema"
    de = "Schema mit automatischer ID"
  }
  types = ["CUSTOM_ENTITY"]

  attributes = [
    {
      key = "customField"
      name = {
        en = "Custom Field"
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
      key = "customNumber"
      name = {
        en = "Custom Number"
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

# Outputs
output "product_schemas" {
  description = "Product schema IDs and URLs"
  value = {
    simple     = emporix_schema.product_simple.id
    rating     = emporix_schema.product_rating.id
    dimensions = emporix_schema.product_dimensions.id
    tags       = emporix_schema.product_tags.id
  }
}

output "product_schema_urls" {
  description = "Product schema URLs"
  value = {
    simple     = emporix_schema.product_simple.schema_url
    rating     = emporix_schema.product_rating.schema_url
    dimensions = emporix_schema.product_dimensions.schema_url
    tags       = emporix_schema.product_tags.schema_url
  }
}

output "customer_schema" {
  description = "Customer schema details"
  value = {
    extended = {
      id         = emporix_schema.customer_extended.id
      name       = lookup(emporix_schema.customer_extended.name, "en", "")
      types      = emporix_schema.customer_extended.types
      schema_url = emporix_schema.customer_extended.schema_url
    }
    address = {
      id         = emporix_schema.customer_address.id
      name       = lookup(emporix_schema.customer_address.name, "en", "")
      types      = emporix_schema.customer_address.types
      schema_url = emporix_schema.customer_address.schema_url
    }
  }
}

output "entity_schemas" {
  description = "Entity schema IDs"
  value       = [for schema in emporix_schema.entities : schema.id]
}

output "entity_schema_urls" {
  description = "Entity schema URLs"
  value       = { for key, schema in emporix_schema.entities : key => schema.schema_url }
}

output "organization_hierarchy_schema" {
  description = "Organization hierarchy schema with 4-level nesting"
  value = {
    id         = emporix_schema.organization_hierarchy.id
    name       = lookup(emporix_schema.organization_hierarchy.name, "en", "")
    schema_url = emporix_schema.organization_hierarchy.schema_url
  }
}

output "auto_generated_id_schema" {
  description = "Schema with auto-generated ID"
  value = {
    id         = emporix_schema.auto_generated_id.id
    name       = lookup(emporix_schema.auto_generated_id.name, "en", "")
    schema_url = emporix_schema.auto_generated_id.schema_url
  }
}

output "all_schema_ids" {
  description = "All schema IDs created"
  value = concat(
    [
      emporix_schema.product_simple.id,
      emporix_schema.product_rating.id,
      emporix_schema.customer_extended.id,
      emporix_schema.product_dimensions.id,
      emporix_schema.customer_address.id,
      emporix_schema.organization_hierarchy.id,
      emporix_schema.product_tags.id,
      emporix_schema.shared_metadata.id,
      emporix_schema.auto_generated_id.id
    ],
    [for schema in emporix_schema.entities : schema.id]
  )
}

output "all_schema_urls" {
  description = "All schema URLs created"
  value = concat(
    [
      emporix_schema.product_simple.schema_url,
      emporix_schema.product_rating.schema_url,
      emporix_schema.customer_extended.schema_url,
      emporix_schema.product_dimensions.schema_url,
      emporix_schema.customer_address.schema_url,
      emporix_schema.organization_hierarchy.schema_url,
      emporix_schema.product_tags.schema_url,
      emporix_schema.shared_metadata.schema_url,
      emporix_schema.auto_generated_id.schema_url
    ],
    [for schema in emporix_schema.entities : schema.schema_url]
  )
}