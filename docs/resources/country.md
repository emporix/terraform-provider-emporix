---
page_title: "emporix_country Resource - terraform-provider-emporix"
subcategory: ""
description: |-
  Country resource for managing country active status in Emporix.
---

# emporix_country (Resource)

Manages a country's active status in Emporix.

**Important:** Countries are pre-populated by Emporix. When you add this resource to your Terraform configuration, it automatically adopts the existing country and allows you to manage its active status. No import required!

**Delete Behavior:** When you remove the resource from Terraform or run `terraform destroy`, the country is **deactivated** (active = false) instead of being deleted.

## Example Usage

### Activate a Country

```terraform
resource "emporix_country" "usa" {
  code   = "US"
  active = true
}
```

### Using Default (Active = True)

```terraform
# active defaults to true, so this is the same as above
resource "emporix_country" "usa" {
  code = "US"
}
```

### Deactivate a Country

```terraform
resource "emporix_country" "canada" {
  code   = "CA"
  active = false
}
```

### Managing Multiple Countries

```terraform
resource "emporix_country" "usa" {
  code = "US"
  # active defaults to true
}

resource "emporix_country" "canada" {
  code = "CA"
  # active defaults to true
}

resource "emporix_country" "mexico" {
  code   = "MX"
  active = false
}
```

### Using Variables

```terraform
variable "active_countries" {
  description = "List of country codes to activate"
  type        = list(string)
  default     = ["US", "CA", "GB", "DE"]
}

resource "emporix_country" "countries" {
  for_each = toset(var.active_countries)

  code = each.value
  # active defaults to true
}
```

## Schema

### Required

- `code` (String) Country code (ISO 3166-1 alpha-2, 2-letter code). Cannot be changed after creation. Changing this forces a new resource to be created.

### Optional

- `active` (Boolean) Whether the country is active for the tenant. Only active countries are visible in the system. **Defaults to `true`**.

### Read-Only

- `name` (Map of String) Localized country names. Read-only, populated from Emporix.
- `regions` (List of String) Regions the country belongs to. Read-only, populated from Emporix.

## Import

Import is optional but supported. If you have countries managed outside Terraform, you can import them:

```shell
# Import United States
terraform import emporix_country.usa US

# Import United Kingdom
terraform import emporix_country.uk GB
```

However, in most cases you don't need to import - just add the resource and Terraform will adopt it automatically.

## Required OAuth Scopes

To manage countries, your client_id/secret pair (used in provider section) must have the following scopes:

**Required Scopes:**
- `country.country_read` - Required for reading country information
- `country.country_manage` - Required for updating country active status

## Important Notes

### Countries Are Pre-populated

Countries are pre-populated by Emporix and cannot be created through the API. When you add a country resource to your Terraform configuration:

1. Terraform fetches the existing country from Emporix
2. If you specified an `active` value (or it defaults to `true`), and it differs from the current state, Terraform updates it
3. The country is added to Terraform state

**No import required!** Just add the resource:

```terraform
resource "emporix_country" "usa" {
  code = "US"
}
```

### Delete Behavior: Countries Are Deactivated

When you remove a country resource from your Terraform configuration or run `terraform destroy`:

1. Terraform sets the country's `active` field to `false`
2. The country is deactivated in Emporix
3. The country is removed from Terraform state

**The country is not deleted from Emporix** - it's only deactivated. This is the expected behavior since countries are system data and cannot be deleted.

```terraform
# Before: country is active
resource "emporix_country" "test" {
  code = "DE"
}

# After removing from config:
# terraform apply
# → Country "DE" is deactivated (active = false)
# → Removed from Terraform state
```

### Active Field Defaults to True

If you don't specify the `active` field, it defaults to `true`:

```terraform
# These are equivalent:
resource "emporix_country" "usa" {
  code = "US"
}

resource "emporix_country" "usa" {
  code   = "US"
  active = true
}
```

### Only Active Status Can Be Modified

You can only modify the `active` field. All other fields (`name`, `regions`) are read-only and managed by Emporix.

### Country Codes

Country codes follow the ISO 3166-1 alpha-2 standard (2-letter codes). Common examples:
- `US` - United States
- `GB` - United Kingdom
- `DE` - Germany
- `FR` - France
- `CA` - Canada
- `PL` - Poland
- `ES` - Spain
- `IT` - Italy

For a complete list, refer to [ISO 3166-1 alpha-2](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2).

## Workflow

### Activate Countries

```terraform
resource "emporix_country" "usa" {
  code = "US"
}

resource "emporix_country" "canada" {
  code = "CA"
}
```

```bash
terraform apply
# → US and CA are activated (if not already active)
```

### Deactivate a Country

**Option 1: Set active = false**
```terraform
resource "emporix_country" "usa" {
  code   = "US"
  active = false
}
```

**Option 2: Remove from configuration**
```terraform
# Remove the resource from your .tf file
# Then run:
```
```bash
terraform apply
# → US is deactivated
```

### Reactivate a Country

```terraform
# Add it back with active = true (or omit active, defaults to true)
resource "emporix_country" "usa" {
  code = "US"
}
```

```bash
terraform apply
# → US is activated again
```

## Advanced Usage

### Bulk Operations

Activate multiple countries:

```terraform
locals {
  active_countries = ["US", "CA", "GB", "DE", "FR", "ES", "IT"]
}

resource "emporix_country" "active" {
  for_each = toset(local.active_countries)

  code = each.value
}
```

### Conditional Activation

```terraform
variable "environment" {
  type = string
}

variable "production_countries" {
  type    = list(string)
  default = ["US", "CA", "GB"]
}

variable "staging_countries" {
  type    = list(string)
  default = ["US"]
}

resource "emporix_country" "countries" {
  for_each = toset(
    var.environment == "production" 
      ? var.production_countries 
      : var.staging_countries
  )

  code = each.value
}
```

### Mix Active and Inactive

```terraform
resource "emporix_country" "active_us" {
  code = "US"
  # defaults to active = true
}

resource "emporix_country" "inactive_test" {
  code   = "DE"
  active = false
}
```

### Output Country Information

```terraform
resource "emporix_country" "usa" {
  code = "US"
}

output "usa_info" {
  value = {
    code    = emporix_country.usa.code
    active  = emporix_country.usa.active
    name_en = lookup(emporix_country.usa.name, "en", "United States")
    regions = emporix_country.usa.regions
  }
}
```

## Lifecycle Examples

### Activate on Create
```terraform
resource "emporix_country" "usa" {
  code = "US"
}
```
```bash
terraform apply
# → Fetches US from Emporix
# → Sets active = true (if not already)
# → Adds to state
```

### Deactivate Explicitly
```terraform
resource "emporix_country" "usa" {
  code   = "US"
  active = false
}
```
```bash
terraform apply
# → Updates US in Emporix: active = false
```

### Remove from Terraform
```terraform
# Remove resource from configuration
```
```bash
terraform apply
# → Deactivates US: active = false
# → Removes from state
```

### Reactivate
```terraform
resource "emporix_country" "usa" {
  code = "US"
}
```
```bash
terraform apply
# → Fetches US from Emporix
# → Sets active = true
# → Adds to state
```
