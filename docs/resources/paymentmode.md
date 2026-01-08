---
page_title: "emporix_paymentmode Resource - terraform-provider-emporix"
subcategory: ""
description: |-
  Payment mode resource for configuring payment methods in Emporix.
---

# emporix_paymentmode (Resource)

Manages payment mode configurations in Emporix. Payment modes define the available payment methods.

## Example Usage

### Invoice Payment

```terraform
resource "emporix_paymentmode" "invoice" {
  code             = "invoice"
  active           = true
  payment_provider = "INVOICE"
}
```

### Cash on Delivery Payment

```terraform
resource "emporix_paymentmode" "cod" {
  code             = "cash_on_delivery"
  active           = true
  payment_provider = "CASH_ON_DELIVERY"
}
```

### Disabled Payment Mode

```terraform
resource "emporix_paymentmode" "test_mode" {
  code             = "test_payment"
  active           = false
  payment_provider = "INVOICE"
}
```

## Schema

### Required

- `code` (String) Code of the payment mode (unique identifier). Changing this forces a new resource to be created.
- `payment_provider` (String) Payment provider type. Currently supported: `INVOICE`, `CASH_ON_DELIVERY`

### Optional

- `active` (Boolean) Indicates whether the payment mode is active. Defaults to `true`.
- `configuration` (Map of String) Map of configuration values for the payment gateway. Not required for INVOICE and CASH_ON_DELIVERY.

### Read-Only

- `id` (String) Unique identifier of the payment mode (UUID)

## Supported Payment Providers

The resource supports the following payment providers:

- **INVOICE** - Simple invoice payment (no configuration required)
- **CASH_ON_DELIVERY** - Cash on delivery payment (no configuration required)

## Import

Payment modes can be imported using their ID:

```shell
terraform import emporix_paymentmode.example 92d77b2b-9385-43ad-a859-55176fbcbd36
```

## Required OAuth Scopes

To manage payment modes, your client_id/secret pair (used in provider section) must have the following scopes:

**Required Scopes:**
- `payment-gateway.paymentmodes_read` - Required for reading payment mode information
- `payment-gateway.paymentmodes_manage` - Required for creating, updating, and deleting payment modes

## Notes

- The `code` attribute is used as the unique identifier and cannot be changed after creation.
- Changing the `code` will force the creation of a new resource.
- The `payment_provider` field cannot be changed after creation.
- INVOICE and CASH_ON_DELIVERY providers don't require any configuration.
- The API returns a UUID as the `id` which is used for updates and deletion.
