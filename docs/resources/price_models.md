---
page_title: "emporix_price_models Resource - terraform-provider-emporix"
subcategory: ""
description: |-
  Manages price models that define a pricing strategy for prices.
---

# emporix_price_models (Resource)

Manages an Emporix price model. A price model defines the pricing strategy (`BASIC`, `VOLUME`, or `TIERED`) and the
quantity tiers that prices assigned to it are calculated against. See the
[Price Models API](https://developer.emporix.io/api-references-1/readme/api-reference-26/price-models) for details.

**Delete Behavior:** When you remove the resource from Terraform or run `terraform destroy`, the price model is
**deleted** from Emporix. If prices are still assigned to the price model, the delete fails unless `force_delete` is set to `true`. The tenant's current default price model can never be deleted, even with `force_delete` - a different price model must be assigned as the default first.

## Pricing Strategies

- **BASIC** - Same price per unit, regardless of the ordered quantity. Requires exactly one tier with
  `min_quantity.quantity = 0`.
- **VOLUME** - Lower price per unit depending on the total ordered quantity.
- **TIERED** - Lower price per unit based on the tier that the total ordered quantity falls into; pricing applies
  per tier range.

For `VOLUME` and `TIERED` price models, each tier's `min_quantity.quantity` must be unique, tiers must be in
ascending order, the first tier must start at `0`, and all tiers must share the same `min_quantity.unit_code`.
`measurement_unit.unit_code` must also match every tier's `min_quantity.unit_code`, for all pricing strategies.
Additionally, every tier's `min_quantity.quantity` must be an integer multiple of `measurement_unit.quantity`.

## Example Usage

### Basic Price Model (Flat Pricing)

```terraform
resource "emporix_price_models" "standard" {
  id           = "standard-pricing"
  includes_tax = true

  name = {
    en = "Standard Pricing"
    de = "Standardpreis"
  }

  tier_definition = {
    tier_type = "BASIC"
    tiers = [
      {
        min_quantity = {
          quantity  = 0
          unit_code = "piece"
        }
      }
    ]
  }

  measurement_unit = {
    quantity  = 1
    unit_code = "piece"
  }
}
```

### Volume Pricing

```terraform
resource "emporix_price_models" "volume" {
  id           = "volume-pricing"
  includes_tax = true
  default      = false

  name = {
    en = "Volume Pricing"
  }

  description = {
    en = "Lower unit price the more you buy"
  }

  tier_definition = {
    tier_type = "VOLUME"
    tiers = [
      {
        min_quantity = {
          quantity  = 0
          unit_code = "piece"
        }
      },
      {
        min_quantity = {
          quantity  = 10
          unit_code = "piece"
        }
      },
      {
        min_quantity = {
          quantity  = 50
          unit_code = "piece"
        }
      }
    ]
  }

  measurement_unit = {
    quantity  = 1
    unit_code = "piece"
  }
}
```

### Tiered Pricing

```terraform
resource "emporix_price_models" "tiered" {
  id           = "tiered-pricing"
  includes_tax = false

  name = {
    en = "Tiered Pricing"
  }

  tier_definition = {
    tier_type = "TIERED"
    tiers = [
      {
        min_quantity = {
          quantity  = 0
          unit_code = "kg"
        }
      },
      {
        min_quantity = {
          quantity  = 100
          unit_code = "kg"
        }
      }
    ]
  }

  measurement_unit = {
    quantity  = 1
    unit_code = "kg"
  }
}
```

## Schema

### Required

- `includes_tax` (Boolean) Whether prices calculated with this price model are gross (`true`, tax included) or net (`false`, tax excluded).
- `name` (Map of String) Price model name as a map of language codes to translated names. Example: `{en = "Standard Pricing"}`. At least one language is required.
- `tier_definition` (Object) Defines the pricing strategy and quantity tiers for this price model. (see [tier_definition](#tier_definition) below)
- `measurement_unit` (Object) Measurement unit that this price model's tiers are expressed in. Required by the API even for `BASIC` price models. (see [measurement_unit](#measurement_unit) below)

### Optional

- `id` (String) Unique identifier of the price model. Generated automatically if not provided. Cannot be changed after creation. Changing this forces a new resource to be created.
- `includes_markup` (Boolean) Whether the price model operates in markup preview mode. Defaults to `false`.
- `default` (Boolean) Whether this is the tenant's default price model. Defaults to `false`.
- `description` (Map of String) Optional price model description as a map of language codes to translated descriptions.
- `force_delete` (Boolean) If `true`, deleting this resource also deletes (asynchronously) all prices assigned to this price model. Requires the `price.pricemodel_manage_admin` scope. Defaults to `false`.

### Nested Schema for `tier_definition`

Required block. Defines the pricing strategy and quantity tiers.

**Required:**

- `tier_type` (String) Pricing strategy. One of `BASIC`, `VOLUME`, or `TIERED`.
- `tiers` (List of Objects) Quantity tiers for this price model. (see [tiers](#tiers) below)

### Nested Schema for `tiers`

Required nested block list.

**Required:**

- `min_quantity` (Object) Minimum ordered quantity from which this tier applies. (see [min_quantity](#min_quantity) below)

**Read-Only:**

- `id` (String) Tier identifier, assigned automatically by the API.

### Nested Schema for `min_quantity`

**Required:**

- `quantity` (Number) Minimum quantity value. Must be zero or a positive value.
- `unit_code` (String) Unit code for the quantity. Must match across all tiers of the same price model.

### Nested Schema for `measurement_unit`

**Required:**

- `quantity` (Number) Measurement unit quantity. Must be zero or a positive value.
- `unit_code` (String) Measurement unit code.

## Import

Price models can be imported using their ID:

```shell
terraform import emporix_price_models.standard standard-pricing
```

## Required OAuth Scopes

To manage price models, your client_id/secret pair must have:

**Required Scopes:**
- `price.pricemodel_read` - Required for reading price models
- `price.pricemodel_manage` - Required for creating, updating, and deleting price models
- `price.pricemodel_manage_admin` - Required to delete a price model that still has prices assigned to it (via `force_delete = true`)

## Outputs

### id

The unique identifier of the price model.

```terraform
output "standard_price_model_id" {
  value = emporix_price_models.standard.id
}
```

### tier_definition

The tier definition, including any tier IDs assigned by the API.

```terraform
output "volume_tiers" {
  value = emporix_price_models.volume.tier_definition.tiers
}
```
