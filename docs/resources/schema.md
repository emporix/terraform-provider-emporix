---
page_title: "emporix_schema Resource - terraform-provider-emporix"
subcategory: ""
description: |-
  Schema resource for managing schemas in Emporix.
---

# emporix_schema (Resource)

Manages schemas in Emporix. Schemas define the structure and validation rules for various entity types in the system such as products, customers, orders, and custom entities.

**Delete Behavior:** When you remove the resource from Terraform or run `terraform destroy`, the schema is **deleted** from Emporix. Note that the database entry is removed, but any associated Cloudinary files persist.

## Upgrading to v0.7.0

~> **Breaking Change:** Version 0.7.0 introduces a breaking change to the `attributes` field. The internal representation changed from a static nested list to a dynamic type to support unlimited nesting depth.

**If you are upgrading from v0.6.x or earlier**, you will encounter this error when running `terraform plan` or `terraform apply`:

```
Error: Unable to Read Previously Saved State for UpgradeResourceState

AttributeName("attributes"): invalid JSON, expected "{", got "["
```

**To resolve this issue**, remove the affected resources from state and re-import them:

```bash
# Remove schema from state
terraform state rm emporix_schema.your_resource_name

# Re-import the schema
terraform import emporix_schema.your_resource_name <schema-id>
```

Or remove all schema resources at once:

```bash
terraform state rm 'emporix_schema.*'
```

Then run `terraform apply` to refresh the state with the new format.

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

### Schema with Auto-Generated ID

When you don't specify an `id`, the Emporix API will automatically generate one:

```terraform
resource "emporix_schema" "auto_generated" {
  # No 'id' specified - API will generate one automatically
  name = {
    en = "Auto Generated Schema"
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
    }
  ]
}

# Reference the auto-generated ID
output "auto_schema_id" {
  value = emporix_schema.auto_generated.id
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

### Schema with Deeply Nested OBJECT Attributes (OBJECT within OBJECT)

```terraform
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
          # Nested OBJECT within OBJECT - GPS coordinates
          key = "coordinates"
          name = {
            en = "GPS Coordinates"
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
            { value = "home" },
            { value = "work" },
            { value = "billing" }
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

- `name` (Map of String) Schema name as a map of language code to name (e.g., {"en": "Product Schema", "de": "Produktschema"}). Provide at least one language translation.
- `types` (List of String) List of schema types this schema applies to. Valid values: `CART`, `CATEGORY`, `COMPANY`, `COUPON`, `CUSTOMER`, `CUSTOMER_ADDRESS`, `ORDER`, `PRODUCT`, `QUOTE`, `RETURN`, `PRICE_LIST`, `SITE`, `CUSTOM_ENTITY`, `VENDOR`.
- `attributes` (Dynamic) List of schema attributes defining the structure. Supports unlimited nesting of OBJECT types. (see [below for nested schema](#nestedatt--attributes))

### Optional

- `id` (String) Schema identifier. If not provided, the API will generate one automatically. Cannot be changed after creation. Changing this forces a new resource to be created.

<a id="nestedatt--attributes"></a>
### Nested Schema for `attributes`

Required:

- `key` (String) Unique attribute identifier.
- `name` (Map of String) Attribute name as a map of language code to name.
- `type` (String) Attribute type. Valid values: `TEXT`, `NUMBER`, `DECIMAL`, `BOOLEAN`, `DATE`, `TIME`, `DATE_TIME`, `ENUM`, `ARRAY`, `OBJECT`, `REFERENCE`.
- `metadata` (Attributes) Attribute metadata. (see [below for nested schema](#nestedatt--attributes--metadata))

Optional:

- `description` (Map of String) Attribute description as a map of language code to description.
- `values` (Attributes List) List of allowed values for `ENUM` or `REFERENCE` types. (see [below for nested schema](#nestedatt--attributes--values))
- `attributes` (Dynamic List) Nested attributes for `OBJECT` type. Supports unlimited nesting depth. (see [below for nested schema](#nestedatt--attributes--attributes))
- `array_type` (Attributes) Array type configuration for `ARRAY` attributes. (see [below for nested schema](#nestedatt--attributes--array_type))

<a id="nestedatt--attributes--metadata"></a>
### Nested Schema for `attributes.metadata`

Required:

- `read_only` (Boolean) Whether the attribute is read-only.
- `localized` (Boolean) Whether the attribute is localized.
- `required` (Boolean) Whether the attribute is required.
- `nullable` (Boolean) Whether the attribute can be null.

<a id="nestedatt--attributes--values"></a>
### Nested Schema for `attributes.values`

Required:

- `value` (String) Allowed value string for `ENUM` or `REFERENCE` type.

<a id="nestedatt--attributes--attributes"></a>
### Nested Schema for `attributes.attributes`

Nested attributes for `OBJECT` type. Supports unlimited nesting depth - each nested attribute has the same structure as top-level attributes and can contain further nested OBJECT types.

Required:

- `key` (String) Unique attribute identifier.
- `name` (Map of String) Attribute name as a map of language code to name.
- `type` (String) Attribute type. Valid values: `TEXT`, `NUMBER`, `DECIMAL`, `BOOLEAN`, `DATE`, `TIME`, `DATE_TIME`, `ENUM`, `ARRAY`, `OBJECT`, `REFERENCE`.
- `metadata` (Object) Attribute metadata with `read_only`, `localized`, `required`, and `nullable` booleans.

Optional:

- `description` (Map of String) Attribute description as a map of language code to description.
- `values` (List of Object) List of allowed values for `ENUM` or `REFERENCE` types. Each value has a `value` string field.
- `attributes` (List) Further nested attributes for `OBJECT` type (recursive, unlimited depth).
- `array_type` (Object) Array type configuration for `ARRAY` attributes with `type`, `localized`, and optional `values` fields.

<a id="nestedatt--attributes--array_type"></a>
### Nested Schema for `attributes.array_type`

Required:

- `type` (String) Element type for the array.

Optional:

- `localized` (Boolean) Whether array elements are localized.
- `values` (Attributes List) List of allowed values for `ENUM` array elements. (see [below for nested schema](#nestedatt--attributes--array_type--values))

<a id="nestedatt--attributes--array_type--values"></a>
#### Nested Schema for `attributes.array_type.values`

Required:

- `value` (String) Allowed value for `ENUM` array element.

## Outputs

All input arguments (`id`, `name`, `types`, `attributes`) and the computed `schema_url` attribute are available as outputs and can be referenced from other resources or outputs.

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
OBJECT type attributes can contain nested attributes. Nested attributes support the full range of types including OBJECT (for deeper nesting), ENUM (with values), and ARRAY (with array_type). **The provider supports unlimited nesting depth** - you can nest OBJECT within OBJECT as deeply as your use case requires.

Each nested attribute has the same structure as top-level attributes:
- `key` - Unique identifier
- `name` - Localized name map
- `type` - Attribute type (TEXT, NUMBER, OBJECT, etc.)
- `metadata` - Read-only, localized, required, nullable flags
- `description` - Optional localized description
- `values` - For ENUM/REFERENCE types
- `attributes` - For nested OBJECT types (recursive)
- `array_type` - For ARRAY types

Example with deeply nested OBJECT:

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
      # Nested OBJECT within OBJECT
      key = "coordinates"
      name = { en = "GPS Coordinates" }
      type = "OBJECT"
      metadata = {
        read_only = false
        localized = false
        required  = false
        nullable  = true
      }
      attributes = [
        {
          key = "latitude"
          name = { en = "Latitude" }
          type = "DECIMAL"
          metadata = {
            read_only = false
            localized = false
            required  = true
            nullable  = false
          }
        },
        {
          key = "longitude"
          name = { en = "Longitude" }
          type = "DECIMAL"
          metadata = {
            read_only = false
            localized = false
            required  = true
            nullable  = false
          }
        }
      ]
    },
    {
      # Nested ENUM within OBJECT
      key = "addressType"
      name = { en = "Address Type" }
      type = "ENUM"
      metadata = {
        read_only = false
        localized = false
        required  = false
        nullable  = true
      }
      values = [
        { value = "home" },
        { value = "work" }
      ]
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

5. **Organize complex data with OBJECT**: Group related attributes using `OBJECT` type for better structure. You can nest OBJECT within OBJECT for complex hierarchical data (unlimited depth).

6. **Version your schemas**: Consider including version numbers in schema IDs for easier migration (e.g., `product-custom-v1`).

## Notes

- Schema IDs are immutable. If you need to change an ID, you must destroy and recreate the resource.
- The `name` field is required and must contain at least one language translation.
- Schemas support multiple entity types - a single schema can apply to multiple types.
- Updates to schemas require providing the `metadata.version` field, which is handled automatically by the provider.
- Nested OBJECT attributes support unlimited nesting depth.
