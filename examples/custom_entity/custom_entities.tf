# Example Terraform configuration for Emporix Custom Entities
#
# Custom entity instances are data records that live under a custom schema type,
# registered via emporix_custom_entity_type (a distinct resource from emporix_schema).
# The `type` argument of emporix_custom_entity must match the `id` of such a
# emporix_custom_entity_type resource. Optionally, an emporix_schema resource
# (types = ["CUSTOM_ENTITY"]) can additionally be defined to validate instance
# structure, but it is not required for the basic usage shown here.

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
# Required scopes:
#   - schema.schema_manage (to manage the emporix_custom_entity_type resource itself)
#   - schema.custominstance_manage, or the per-type custom.<lowercase-type>_manage (to manage instances)
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

# The custom schema type that our custom entity instances will belong to.
# Creating this also auto-generates the per-type OAuth scopes
# custom.document_manage, custom.document_manage_own, custom.document_read,
# custom.document_read_own (the type id is lowercased in the scope name).
resource "emporix_custom_entity_type" "document" {
  id = "DOCUMENT"
  name = {
    en = "Document"
  }
}

# Optional: define a generic schema to validate the structure of DOCUMENT
# instances. Not required - custom entity instances work with unstructured
# "mixins" data even without one.
#
# The link to the custom type is the "types" field itself: per the API docs,
# "types" accepts either a predefined value or "any custom schema type id" -
# so referencing the custom type's own id here (not the generic literal
# "CUSTOM_ENTITY") is what actually attaches this schema to DOCUMENT instances.
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

# Example 1: Plain custom entity instance, no owner, no mixins.
resource "emporix_custom_entity" "welcome_doc" {
  type = emporix_custom_entity_type.document.id
  name = {
    en = "Welcome Document"
  }
}

# Example 2: Custom entity instance with an explicit owner.
# Ownership is immutable once set - changing "owner" forces replacement.
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

# Example 3: Custom entity instance with arbitrary mixin data.
# "mixins" holds instance-specific data as a JSON-encoded string.
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

# Outputs
output "welcome_doc" {
  description = "The welcome document instance"
  value = {
    id         = emporix_custom_entity.welcome_doc.id
    created_at = emporix_custom_entity.welcome_doc.created_at
    version    = emporix_custom_entity.welcome_doc.version
  }
}

output "invoice_doc_mixins" {
  description = "The mixin data stored on the invoice document instance"
  value       = emporix_custom_entity.invoice_doc.mixins
}
