# Example Terraform configuration for Emporix Payment Modes

terraform {
  required_providers {
    emporix = {
      source  = "emporix/emporix"
      version = "~> 0.1"
    }
  }
}

# Configure the Emporix provider
provider "emporix" {
  tenant  = var.emporix_tenant
  api_url = var.emporix_api_url
  
  # Use client credentials
  client_id     = var.emporix_client_id
  client_secret = var.emporix_client_secret
  scope         = "tenant=${var.emporix_tenant} payment-gateway.paymentmodes_read payment-gateway.paymentmodes_manage"
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

# Example 1: Invoice payment (no configuration required)
resource "emporix_paymentmode" "invoice" {
  code             = "invoice"
  active           = true
  payment_provider = "INVOICE"
}

# Example 2: Cash on delivery payment (no configuration required)
resource "emporix_paymentmode" "cash_on_delivery" {
  code             = "cash_on_delivery"
  active           = true
  payment_provider = "CASH_ON_DELIVERY"
}

# Example 3: Disabled payment mode
resource "emporix_paymentmode" "test_mode" {
  code             = "test_payment"
  active           = false
  payment_provider = "INVOICE"
}

# Outputs
output "payment_mode_ids" {
  description = "Payment mode IDs"
  value = {
    invoice          = emporix_paymentmode.invoice.id
    cash_on_delivery = emporix_paymentmode.cash_on_delivery.id
    test_mode        = emporix_paymentmode.test_mode.id
  }
}

output "active_payment_modes" {
  description = "Codes of active payment modes"
  value = [
    for mode in [
      emporix_paymentmode.invoice,
      emporix_paymentmode.cash_on_delivery,
      emporix_paymentmode.test_mode,
    ] : mode.code if mode.active
  ]
}
