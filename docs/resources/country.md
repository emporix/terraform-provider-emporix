---
page_title: "emporix_country Resource - terraform-provider-emporix"
subcategory: ""
description: |-
  Country resource for managing country active status in Emporix.
---

# emporix_country (Resource)

Manages a country's active status in Emporix.

**Important:** Countries are pre-populated by Emporix and cannot be created or deleted through the API. This resource only manages the `active` field, which controls whether a country is visible in the system. You must import existing countries before managing them with Terraform.

## Example Usage

### Import and Activate a Country

```terraform
# First, import the country
# terraform import emporix_country.usa US

resource "emporix_country" "usa" {
  code   = "US"
  active = true
}
```

### Import and Deactivate a Country

```terraform
# First, import the country
# terraform import emporix_country.canada CA

resource "emporix_country" "canada" {
  code   = "CA"
  active = false
}
```

### Managing Multiple Countries

```terraform
# Import countries first:
# terraform import emporix_country.usa US
# terraform import emporix_country.canada CA
# terraform import emporix_country.mexico MX

resource "emporix_country" "usa" {
  code   = "US"
  active = true
}

resource "emporix_country" "canada" {
  code   = "CA"
  active = true
}

resource "emporix_country" "mexico" {
  code   = "MX"
  active = false
}
```

## Schema

### Required

- `code` (String) Country code (ISO 3166-1 alpha-2, 2-letter code). Cannot be changed after import. Changing this forces a new resource to be created.

### Optional

- `active` (Boolean) Whether the country is active for the tenant. Only active countries are visible in the system. Defaults to `true`.

### Read-Only

- `name` (Map of String) Localized country names. Read-only, populated from Emporix.
- `regions` (List of String) Regions the country belongs to. Read-only, populated from Emporix.

## Import

Countries **must be imported** before they can be managed. Use the country code as the import identifier:

```shell
# Import United States
terraform import emporix_country.usa US

# Import United Kingdom
terraform import emporix_country.uk GB

# Import Germany
terraform import emporix_country.germany DE

# Import Poland
terraform import emporix_country.poland PL
```

## Required OAuth Scopes

To manage countries, include the following scopes in your provider configuration:

```hcl
provider "emporix" {
  tenant        = "mytenant"
  client_id     = var.client_id
  client_secret = var.client_secret
  scope         = "tenant=mytenant country.country_read country.country_manage"
}
```

**Scopes:**
- `country.country_read` - Required for reading countries
- `country.country_manage` - Required for updating countries

## Important Notes

### Countries Cannot Be Created
Countries are pre-populated by Emporix and cannot be created through Terraform or the API. Attempting to create a country resource without importing it first will result in an error.

**Wrong (will fail):**
```terraform
resource "emporix_country" "usa" {
  code   = "US"
  active = true
}
# ❌ Error: Countries must be imported first
```

**Correct:**
```shell
# First import
terraform import emporix_country.usa US

# Then manage in Terraform
resource "emporix_country" "usa" {
  code   = "US"
  active = true
}
# ✅ Works correctly
```

### Countries Cannot Be Deleted
When you run `terraform destroy` or remove a country resource, Terraform will only remove it from the state file. The country itself remains in Emporix and is not deleted from the system.

### Only Active Status Can Be Modified
The only field that can be modified is `active`. All other fields (`name`, `regions`) are read-only and managed by Emporix.

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

The typical workflow for managing countries:

1. **Import existing countries** you want to manage:
   ```shell
   terraform import emporix_country.usa US
   terraform import emporix_country.canada CA
   ```

2. **Define them in your Terraform configuration**:
   ```terraform
   resource "emporix_country" "usa" {
     code   = "US"
     active = true
   }
   
   resource "emporix_country" "canada" {
     code   = "CA"
     active = false
   }
   ```

3. **Apply changes** to activate/deactivate countries:
   ```shell
   terraform apply
   ```

4. **Remove from Terraform state** when no longer needed:
   ```shell
   terraform destroy
   # Note: Countries remain in Emporix, only removed from Terraform state
   ```
