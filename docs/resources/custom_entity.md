---
page_title: "emporix_custom_entity Resource - terraform-provider-emporix"
subcategory: ""
description: |-
  Manages a custom entity instance in Emporix.
---

# emporix_custom_entity (Resource)

Manages a custom entity instance in Emporix. Custom entity instances are data records that live under a custom schema type, managed via the `emporix_custom_entity_type` resource (a distinct resource from `emporix_schema`).

The `type` argument must match the `id` of an existing `emporix_custom_entity_type` resource. You may optionally also define an `emporix_schema` resource to enforce a validated structure for instances - reference this type by setting the schema's `types` to the custom type's own `id` (e.g. `types = [emporix_custom_entity_type.document.id]`), not the generic `CUSTOM_ENTITY` literal. This is not required for basic usage; without a schema, `mixins` is unvalidated freeform data.

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

resource "emporix_custom_entity" "welcome_doc" {
  type = emporix_custom_entity_type.document.id
  name = {
    en = "Welcome Document"
  }
}
```

### With an Owner

Ownership is immutable once set - changing `owner` forces replacement. `owner.user_id` is required whenever `owner` is set; `owner.legal_entity_id` may only be set when `owner.type` is `CUSTOMER`.

```terraform
resource "emporix_custom_entity" "employee_doc" {
  type = emporix_custom_entity_type.document.id
  name = {
    en = "Employee Handbook"
  }
  owner = {
    type    = "EMPLOYEE"
    user_id = "employee-123"
  }
}
```

### With Mixin Data

`mixins` holds arbitrary instance data as a JSON-encoded string.

```terraform
resource "emporix_custom_entity" "invoice_doc" {
  type = emporix_custom_entity_type.document.id
  name = {
    en = "Invoice #1042"
  }
  mixins = jsonencode({
    invoiceNumber = "1042"
    amount        = 199.90
    currency      = "EUR"
  })
}
```

### With a Validating Schema

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

## Schema

### Required

- `type` (String) The custom schema type this instance belongs to - the `id` of an existing `emporix_custom_entity_type` resource (e.g. "DOCUMENT"). Case-sensitive; used as a URL path segment. Cannot be changed after creation. Changing this forces a new resource to be created.
- `name` (Map of String) Display name as a map of language code to name (e.g., {"en": "My Document"}). Provide at least one language translation.

### Optional

- `id` (String) Custom entity instance identifier. If not provided, the API will generate one automatically. Cannot be changed after creation. Changing this forces a new resource to be created.
- `owner` (Attributes) Ownership of this instance. Once set, ownership cannot be changed; modifying it forces replacement. (see [below for nested schema](#nestedatt--owner))
- `mixins` (String) Arbitrary instance data as a JSON-encoded string (e.g. `jsonencode({...})`). Defaults to `"{}"`. The API stores this as a JSON-encoded string value, not a nested object, so its content is opaque to the API unless validated by an attached `emporix_schema`.

### Read-Only

- `media` (List of String) IDs of media assets assigned to this instance. Read-only here; media is assigned through Emporix's media management APIs, not through this resource.
- `created_at` (String) Timestamp when the instance was created.
- `modified_at` (String) Timestamp when the instance was last modified.
- `version` (Number) Instance version (managed by the API).

<a id="nestedatt--owner"></a>
### Nested Schema for `owner`

Required:

- `type` (String) Owner type. Valid values when setting an owner: `EMPLOYEE`, `CUSTOMER`. (`SERVICE` can appear when reading back an instance whose owner was auto-assigned by the API under a `manage_own` scope, but cannot be set explicitly.)
- `user_id` (String) ID of the owning user. Required when `owner` is set.

Optional:

- `legal_entity_id` (String) Legal entity ID of the owner. Only valid when `owner.type` is `CUSTOMER`.

## Import

Custom entity instances can be imported using the format `type:id`:

```shell
terraform import emporix_custom_entity.welcome_doc DOCUMENT:doc-123
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
