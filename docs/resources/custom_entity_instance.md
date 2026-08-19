---
page_title: "emporix_custom_entity_instance Resource - terraform-provider-emporix"
subcategory: ""
description: |-
  Manages a custom entity instance in Emporix.
---

# emporix_custom_entity_instance (Resource)

Manages a custom entity instance in Emporix. Custom entity instances are data records that live under a custom schema type, managed via the `emporix_custom_entity_type` resource (a distinct resource from `emporix_schema`).

The `type` argument must match the `id` of an existing `emporix_custom_entity_type` resource. `mixins` defaults to an empty object (`"{}"`), which needs no schema - but any non-empty `mixins` content must be declared as an attribute in an `emporix_schema` attached to this instance's type (reference the type by setting the schema's `types` to the custom type's own `id`, e.g. `types = [emporix_custom_entity_type.document.id]`, not the generic `CUSTOM_ENTITY` literal). See the Notes section below for the exact `mixins` format.

Both `type` and `owner` are immutable after creation - changing either forces replacement.

## Example Usage

### Basic Instance

```terraform
resource "emporix_custom_entity_type" "document" {
  id = "DOCUMENT"
  name = {
    en = "Document"
  }
}

resource "emporix_custom_entity_instance" "welcome_doc" {
  type = emporix_custom_entity_type.document.id
  name = {
    en = "Welcome Document"
  }
}
```

### With Mixin Data

`mixins` holds instance data as a JSON-encoded string, nested under a top-level key equal to the governing `emporix_schema`'s own `id`. This example uses its own type (`INVOICE`) and matching schema (`invoice_fields`), distinct from `DOCUMENT`/`document_fields` above.

```terraform
resource "emporix_custom_entity_type" "invoice" {
  id = "INVOICE"
  name = {
    en = "Invoice"
  }
}

resource "emporix_schema" "invoice_fields" {
  id = "invoice-fields"
  name = {
    en = "Invoice Fields"
  }
  types = [emporix_custom_entity_type.invoice.id]

  attributes = [
    {
      key = "invoiceNumber"
      name = { en = "Invoice Number" }
      type = "TEXT"
      metadata = {
        read_only = false
        localized = false
        required  = false
        nullable  = true
      }
    },
    {
      key = "amount"
      name = { en = "Amount" }
      type = "DECIMAL"
      metadata = {
        read_only = false
        localized = false
        required  = false
        nullable  = true
      }
    },
    {
      key = "currency"
      name = { en = "Currency" }
      type = "TEXT"
      metadata = {
        read_only = false
        localized = false
        required  = false
        nullable  = true
      }
    }
  ]
}

resource "emporix_custom_entity_instance" "invoice_doc" {
  type = emporix_custom_entity_type.invoice.id
  name = {
    en = "Invoice #1042"
  }
  mixins = jsonencode({
    "invoice-fields" = {
      invoiceNumber = "1042"
      amount        = 199.90
      currency      = "EUR"
    }
  })

  # Ensures the schema exists before the instance is validated against it.
  depends_on = [emporix_schema.invoice_fields]
}
```

### `DOCUMENT`'s Schema

```terraform
resource "emporix_schema" "document_fields" {
  id = "document-fields"
  name = {
    en = "Document Fields"
  }
  types = [emporix_custom_entity_type.document.id]

  attributes = [
    {
      key = "note"
      name = {
        en = "Note"
      }
      type = "TEXT"
      metadata = {
        read_only = false
        localized = false
        required  = false
        nullable  = true
      }
    }
  ]
}
```

### Using for_each

```terraform
locals {
  welcome_docs = {
    "welcome-1" = "Welcome Document 1"
    "welcome-2" = "Welcome Document 2"
  }
}

resource "emporix_custom_entity_instance" "welcome_docs" {
  for_each = local.welcome_docs

  type = emporix_custom_entity_type.document.id
  name = {
    en = each.value
  }
}
```

## Schema

### Required

- `type` (String) The custom schema type this instance belongs to - the `id` of an existing `emporix_custom_entity_type` resource (e.g. "DOCUMENT"). Case-sensitive; used as a URL path segment. Cannot be changed after creation. Changing this forces a new resource to be created.
- `name` (Map of String) Display name as a map of language code to name (e.g., {"en": "My Document"}). Provide at least one language translation.

### Optional

- `id` (String) Custom entity instance identifier. If not provided, the API will generate one automatically. Cannot be changed after creation. Changing this forces a new resource to be created.
- `owner` (Attributes) Ownership of this instance. Cannot be changed after creation. Changing this forces a new resource to be created. (see [below for nested schema](#nestedatt--owner))
- `mixins` (String) Instance data as a JSON-encoded string (e.g. `jsonencode({...})`). Defaults to `"{}"`. Each field must be nested under a top-level key equal to the `id` of the `emporix_schema` that declares it.

### Read-Only

- `media` (List of String) IDs of media assets assigned to this instance. Read-only here; media is assigned through Emporix's media management APIs, not through this resource.
- `created_at` (String) Timestamp when the instance was created.

<a id="nestedatt--owner"></a>
### Nested Schema for `owner`

Required:

- `type` (String) Type of the owner. Valid values: `EMPLOYEE`, `CUSTOMER`.
- `user_id` (String) Identifier of the employee or customer associated with the owner. Must be a real, existing user ID on your tenant.

Optional:

- `legal_entity_id` (String) Legal entity identifier. Can be provided only when `type` is `CUSTOMER`.

## Import

Custom entity instances can be imported using the format `type:id`:

```shell
terraform import emporix_custom_entity_instance.welcome_doc DOCUMENT:doc-123
```

## Required OAuth Scopes

Write operations require one of:
- `schema.custominstance_manage` - Manage instances of any custom type
- `custom.<lowercase-type>_manage` - Manage instances of this specific type (e.g. `custom.document_manage`)
- `custom.<lowercase-type>_manage_own` - Manage only instances owned by the caller

Read operations require one of:
- `schema.custominstance_read` - Read instances of any custom type
- `custom.<lowercase-type>_read` - Read instances of this specific type
- `custom.<lowercase-type>_read_own` - Read only instances owned by the caller

These per-type scopes are auto-generated when the corresponding `emporix_custom_entity_type` is created.

## Notes

- `mixins` fields are nested one level deeper than the attribute definitions - under a top-level key equal to the governing `emporix_schema`'s own `id`, not the field names directly (see the `INVOICE` example above).
- This provider can't look up a platform user's identifier on your behalf. If you set `owner`, source `user_id` yourself from your tenant's own user administration (e.g. the IAM users API or the Management Dashboard).
- `owner.type` can also read back as `SERVICE` for an instance whose owner was auto-assigned by the API under a `manage_own` scope, but `SERVICE` cannot be set explicitly.
- Updates require providing the current `metadata.version`, which is handled automatically by the provider and not exposed as a resource attribute.
