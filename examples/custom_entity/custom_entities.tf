# Example Terraform configuration for Emporix Custom Entities

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
# Creating this auto-generates per-type OAuth scopes (custom.document_manage, etc.).
resource "emporix_custom_entity_type" "document" {
  id = "DOCUMENT"
  name = {
    en = "Document"
  }
}

# Declares the mixin fields DOCUMENT instances are allowed to use.
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

# Example 1: Plain custom entity instance, no mixins.
resource "emporix_custom_entity_instance" "welcome_doc" {
  type = emporix_custom_entity_type.document.id
  name = {
    en = "Welcome Document"
  }
}

# A second custom type, distinct from DOCUMENT.
resource "emporix_custom_entity_type" "invoice" {
  id = "INVOICE"
  name = {
    en = "Invoice"
  }
}

# Declares the mixin fields INVOICE instances are allowed to use.
resource "emporix_schema" "invoice_fields" {
  id = "invoice-fields"
  name = {
    en = "Invoice Fields"
  }
  types = [emporix_custom_entity_type.invoice.id]

  attributes = [
    {
      key = "invoiceNumber"
      name = {
        en = "Invoice Number"
      }
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
      name = {
        en = "Amount"
      }
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
      name = {
        en = "Currency"
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

# Example 2: Custom entity instance with mixin data matching invoice_fields above.
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

# Example 3: Creating multiple DOCUMENT instances in bulk with for_each.
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

# Outputs
output "welcome_doc" {
  description = "The welcome document instance"
  value = {
    id         = emporix_custom_entity_instance.welcome_doc.id
    created_at = emporix_custom_entity_instance.welcome_doc.created_at
  }
}

output "invoice_doc_mixins" {
  description = "The mixin data stored on the invoice document instance"
  value       = emporix_custom_entity_instance.invoice_doc.mixins
}

output "welcome_docs" {
  description = "IDs of the bulk-created welcome document instances"
  value       = { for k, v in emporix_custom_entity_instance.welcome_docs : k => v.id }
}
