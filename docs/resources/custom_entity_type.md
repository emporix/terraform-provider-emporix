---
page_title: "emporix_custom_entity_type Resource - terraform-provider-emporix"
subcategory: ""
description: |-
  Manages a custom schema type in Emporix - the container that custom entity instances belong to.
---

# emporix_custom_entity_type (Resource)

Manages a custom schema type in Emporix - the container that custom entity instances (`emporix_custom_entity_instance`) belong to.

This is a distinct resource from `emporix_schema`: it registers the *type* (e.g. `"DOCUMENT"`) under which instances live at `/schema/{tenant}/custom-entities/{type}/instances`. Creating a custom entity type auto-generates a set of per-type OAuth scopes:

- `custom.<lowercase-id>_manage`
- `custom.<lowercase-id>_manage_own`
- `custom.<lowercase-id>_read`
- `custom.<lowercase-id>_read_own`

These let you grant access to a specific custom type without handing out the tenant-wide `schema.custominstance_*` scopes. Deleting the type does not remove the scopes it generated.

## Example Usage

```terraform
resource "emporix_custom_entity_type" "document" {
  id = "DOCUMENT"
  name = {
    en = "Document"
  }
}
```

### With Multiple Translations

```terraform
resource "emporix_custom_entity_type" "warranty" {
  id = "WARRANTY"
  name = {
    en = "Warranty"
    de = "Garantie"
  }
}
```

## Schema

### Required

- `id` (String) Unique code for the custom type. Must start with an uppercase letter and contain only uppercase letters, digits, and underscores (e.g. "DOCUMENT"). Cannot be `AVAILABILITY` or `LOCATION` (reserved by the platform). Cannot be changed after creation. Changing this forces a new resource to be created.
- `name` (Map of String) Localized custom type name as a map of language code to name (e.g., {"en": "Document"}). Provide at least one language translation.

### Read-Only

- `created_at` (String) Timestamp when the custom type was created.
- `modified_at` (String) Timestamp when the custom type was last modified.
- `version` (Number) Custom type version, used for optimistic locking on updates (managed by the API).

## Import

Custom entity types can be imported using their `id`:

```shell
terraform import emporix_custom_entity_type.document DOCUMENT
```

## Required OAuth Scopes

- `schema.schema_manage` - Required for creating, updating, and deleting custom entity types
- `schema.schema_read` - Required for reading custom entity type information

## Notes

- Deletion fails with an error if any `emporix_schema` or `emporix_custom_entity_instance` resources still reference this type.
- To validate the `mixins` of instances of this type, define an `emporix_schema` resource with `types` set to this resource's `id` (e.g. `types = [emporix_custom_entity_type.document.id]`) rather than the generic `CUSTOM_ENTITY` literal. See `emporix_custom_entity_instance` for the required `mixins` format.
- Instances with non-empty `mixins` require an attached schema; without one, any mixin key is rejected.
