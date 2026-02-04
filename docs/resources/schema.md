---
page_title: "emporix_schema Resource - terraform-provider-emporix"
subcategory: ""
description: |-
  Schema resource for managing schemas in Emporix.
---

# emporix_schema (Resource)

Manages schemas in Emporix. Schemas define the structure and validation rules for various entity types in the system such as products, customers, orders, and custom entities.

**Delete Behavior:** When you remove the resource from Terraform or run `terraform destroy`, the schema is **deleted** from Emporix. Note that the database entry is removed, but any associated Cloudinary files persist.

## Example Usage

### Basic Schema with Text Attributes

```terraform
resource "emporix_schema" "product_custom" {
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
    }
  ]
}
```

### Schema with Multiple Attribute Types

```terraform
resource "emporix_schema" "customer_extended" {
  id = "customer-extended-fields"
  name = {
    en = "Customer Extended Fields"
    de = "Erweiterte Kundenfelder"
  }
  types = ["CUSTOMER"]

  attributes = [
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
```

### Schema with ENUM Attributes

```terraform
resource "emporix_schema" "product_rating" {
  id = "product-rating-schema"
  name = {
    en = "Product Rating Schema"
  }
  types = ["PRODUCT"]

  attributes = [
    {
      key = "qualityRating"
      name = {
        en = "Quality Rating"
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
```

### Schema with Nested OBJECT Attributes

```terraform
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
        }
      ]
    }
  ]
}
```

### Schema with ARRAY Type

```terraform
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
```

### Schema for Multiple Entity Types

```terraform
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
```

### Dynamic Schema Creation

```terraform
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
```

## Schema

### Required

- `id` (String) Schema identifier. Cannot be changed after creation. Changing this forces a new resource to be created.
- `name` (Map of String) Schema name as a map of language code to name (e.g., {"en": "Product Schema", "de": "Produktschema"}). Provide at least one language translation.
- `types` (List of String) List of schema types this schema applies to. Valid values: `CART`, `CATEGORY`, `COMPANY`, `COUPON`, `CUSTOMER`, `CUSTOMER_ADDRESS`, `ORDER`, `PRODUCT`, `QUOTE`, `RETURN`, `PRICE_LIST`, `SITE`, `CUSTOM_ENTITY`, `VENDOR`.
- `attributes` (List of Object) List of schema attributes defining the structure. Each attribute has:
  - `key` (String, Required) Unique attribute identifier.
  - `name` (Map of String, Required) Attribute name as a map of language code to name.
  - `type` (String, Required) Attribute type. Valid values: `TEXT`, `NUMBER`, `DECIMAL`, `BOOLEAN`, `DATE`, `TIME`, `DATE_TIME`, `ENUM`, `ARRAY`, `OBJECT`, `REFERENCE`.
  - `metadata` (Object, Required) Attribute metadata with:
    - `read_only` (Boolean, Required) Whether the attribute is read-only.
    - `localized` (Boolean, Required) Whether the attribute is localized.
    - `required` (Boolean, Required) Whether the attribute is required.
    - `nullable` (Boolean, Required) Whether the attribute can be null.
  - `description` (Map of String, Optional) Attribute description as a map of language code to description.
  - `values` (List of Object, Optional) List of allowed values for `ENUM` or `REFERENCE` types. Each value has:
    - `value` (String, Required) Allowed value string.
  - `attributes` (List of Object, Optional) Nested attributes for `OBJECT` type. Has same structure as parent attributes (without further nesting).
  - `array_type` (Object, Optional) Array type configuration for `ARRAY` attributes. Contains:
    - `type` (String, Required) Element type for the array.
    - `localized` (Boolean, Optional) Whether array elements are localized.
    - `values` (List of Object, Optional) List of allowed values for `ENUM` array elements (same structure as values above).

## Outputs

In addition to all input arguments, the following attributes are exported:

- `schema_url` (String) The URL of the schema, as returned by the API in the `metadata.url` field. This can be used to reference the schema in other configurations or external systems.

All input arguments (`id`, `name`, `types`, `attributes`) are also available as outputs and can be referenced from other resources or outputs.

### Referencing Outputs

```terraform
resource "emporix_schema" "product_custom" {
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

# Access the schema URL
output "product_schema_url" {
  value = emporix_schema.product_custom.schema_url
}

# Access other attributes
output "product_schema_details" {
  value = {
    id         = emporix_schema.product_custom.id
    name       = emporix_schema.product_custom.name
    types      = emporix_schema.product_custom.types
    schema_url = emporix_schema.product_custom.schema_url
  }
}
```

## Import

You can import existing schemas:

```shell
terraform import emporix_schema.product product-custom-fields
```

After importing, Terraform will manage the schema. You'll need to provide all required fields in your configuration after import.

## Required OAuth Scopes

To manage schemas, your client_id/secret pair (used in provider section) must have the following scopes:

**Required Scopes:**
- `schema.schema_read` - Required for reading schema information
- `schema.schema_manage` - Required for creating, updating, and deleting schemas

## Attribute Types

The following attribute types are supported:

### Simple Types
- `TEXT` - Text string
- `NUMBER` - Integer number
- `DECIMAL` - Decimal number with fractional part
- `BOOLEAN` - True/false value
- `DATE` - Date (ISO 8601 format)
- `TIME` - Time (ISO 8601 format)
- `DATE_TIME` - Date and time (ISO 8601 format)

### Complex Types
- `ENUM` - Enumeration with predefined values (requires `values` list)
- `REFERENCE` - Reference to another entity (requires `values` list)
- `ARRAY` - Array of elements (requires `array_type` configuration)
- `OBJECT` - Nested object with attributes (requires `attributes` list)

## Schema Types

The following entity types can have schemas applied:

### Commerce Entities
- `PRODUCT` - Product entities
- `CATEGORY` - Category entities
- `PRICE_LIST` - Price list entities

### Order Management
- `CART` - Shopping cart entities
- `ORDER` - Order entities
- `QUOTE` - Quote entities
- `RETURN` - Return entities

### Customer Management
- `CUSTOMER` - Customer entities
- `CUSTOMER_ADDRESS` - Customer address entities
- `COMPANY` - Company entities

### Other
- `COUPON` - Coupon entities
- `SITE` - Site entities
- `VENDOR` - Vendor entities
- `CUSTOM_ENTITY` - Custom entity types

## Attribute Metadata Flags

Each attribute has metadata that controls its behavior:

### read_only
- `true` - Attribute is read-only and cannot be modified by users
- `false` - Attribute can be modified

### localized
- `true` - Attribute supports multiple languages (values are maps of language code to value)
- `false` - Attribute has single value across all languages

### required
- `true` - Attribute must be provided when creating entity
- `false` - Attribute is optional

### nullable
- `true` - Attribute can have null value
- `false` - Attribute cannot be null

## Working with Nested Attributes

### OBJECT Type
OBJECT type attributes can contain nested attributes. Example:

```terraform
{
  key = "address"
  name = { en = "Address" }
  type = "OBJECT"
  metadata = {
    read_only = false
    localized = false
    required  = false
    nullable  = true
  }
  attributes = [
    {
      key = "street"
      name = { en = "Street" }
      type = "TEXT"
      metadata = {
        read_only = false
        localized = false
        required  = true
        nullable  = false
      }
    },
    {
      key = "city"
      name = { en = "City" }
      type = "TEXT"
      metadata = {
        read_only = false
        localized = false
        required  = true
        nullable  = false
      }
    }
  ]
}
```

### ARRAY Type
ARRAY type attributes require `array_type` configuration:

```terraform
{
  key = "categories"
  name = { en = "Categories" }
  type = "ARRAY"
  metadata = {
    read_only = false
    localized = false
    required  = false
    nullable  = true
  }
  array_type = {
    type      = "TEXT"
    localized = false
  }
}
```

For ENUM arrays, include values:

```terraform
{
  key = "sizes"
  name = { en = "Available Sizes" }
  type = "ARRAY"
  metadata = {
    read_only = false
    localized = false
    required  = false
    nullable  = true
  }
  array_type = {
    type      = "ENUM"
    localized = false
    values = [
      { value = "S" },
      { value = "M" },
      { value = "L" }
    ]
  }
}
```

## Delete Warning

⚠️ **Important:** Deleting a schema removes the schema definition from the system. However:
- Database entry is removed
- Cloudinary files (if any) persist
- Existing entity data with schema-defined attributes may become orphaned

Make sure the schema is not actively used before deleting it.

## Best Practices

1. **Use descriptive IDs**: Choose meaningful schema IDs like `product-custom-fields` instead of generic names.

2. **Provide translations**: Include translations for all languages your application supports.

3. **Set appropriate metadata**: Configure `read_only`, `localized`, `required`, and `nullable` flags based on your business logic.

4. **Use ENUM for predefined values**: When attributes have a fixed set of allowed values, use `ENUM` type with `values` list.

5. **Organize complex data with OBJECT**: Group related attributes using `OBJECT` type for better structure.

6. **Version your schemas**: Consider including version numbers in schema IDs for easier migration (e.g., `product-custom-v1`).

## Notes

- Schema IDs are immutable. If you need to change an ID, you must destroy and recreate the resource.
- The `name` field is required and must contain at least one language translation.
- Schemas support multiple entity types - a single schema can apply to multiple types.
- Updates to schemas require providing the `metadata.version` field, which is handled automatically by the provider.
- Nested attributes (in OBJECT type) cannot be infinitely nested - only one level of nesting is supported.
