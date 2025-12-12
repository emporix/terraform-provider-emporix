---
page_title: "emporix_tenant_configuration Resource - terraform-provider-emporix"
subcategory: ""
description: |-
  Tenant configuration resource for managing tenant configurations in Emporix.
---

# emporix_tenant_configuration (Resource)

Manages tenant configurations in Emporix. Tenant configurations store key-value pairs where values can be any valid JSON (object, string, array, or boolean).

**Delete Behavior:** When you remove the resource from Terraform or run `terraform destroy`, the configuration is **deleted** from Emporix.

## Example Usage

### String Value

```terraform
resource "emporix_tenant_configuration" "project_country" {
  key   = "project_country"
  value = jsonencode("US")
}
```

### Object Value

```terraform
resource "emporix_tenant_configuration" "tax_config" {
  key = "taxConfiguration"
  value = jsonencode({
    taxClassOrder = ["FULL", "HALF", "ZERO"]
    taxClasses = {
      FULL = 19
      HALF = 7
      ZERO = 0
    }
  })
}
```

### Array Value

```terraform
resource "emporix_tenant_configuration" "project_currencies" {
  key = "project_currencies"
  value = jsonencode([
    {
      id       = "USD"
      label    = "US Dollar"
      default  = true
      required = true
    },
    {
      id       = "EUR"
      label    = "Euro"
      default  = false
      required = false
    }
  ])
}
```

### JSON String Value (Double Encoding)

Some configurations like `project_lang` and `project_curr` require the value to be a JSON string:

```terraform
# Using double jsonencode for JSON string values
resource "emporix_tenant_configuration" "project_lang" {
  key = "project_lang"
  value = jsonencode(jsonencode([
    {
      id       = "en"
      label    = "English"
      default  = true
      required = true
    },
    {
      id       = "de"
      label    = "German"
      default  = false
      required = false
    }
  ]))
}

# Cleaner approach using locals
locals {
  languages = [
    {id = "en", label = "English", default = true, required = true},
    {id = "de", label = "German", default = false, required = false}
  ]
}

resource "emporix_tenant_configuration" "project_lang_clean" {
  key   = "project_lang"
  value = jsonencode(jsonencode(local.languages))
}
```
```

### Secured Configuration

```terraform
resource "emporix_tenant_configuration" "api_key" {
  key     = "external_api_key"
  value   = jsonencode("secret-key-value")
  secured = true  # Value will be encrypted
}
```

### Boolean Value

```terraform
resource "emporix_tenant_configuration" "feature_flag" {
  key   = "enable_new_checkout"
  value = jsonencode(true)
}
```

### Complex Configuration

```terraform
resource "emporix_tenant_configuration" "unit_config" {
  key = "unitConf"
  value = jsonencode([
    {
      name        = "UnitConfiguration"
      description = "This object holds the unit configurations"
      units = {
        LTR = {
          availableUnitValue = 1
          conversion = {
            MLT = 1000
          }
        }
        GRM = {
          availableUnitValue = 100
          conversion = {
            KGM = 0.001
          }
        }
        KGM = {
          availableUnitValue = 1
          conversion = {
            GRM = 1000
          }
        }
        MLT = {
          availableUnitValue = 100
          conversion = {
            LTR = 0.001
          }
        }
      }
    }
  ])
}
```

### Multiple Configurations

```terraform
resource "emporix_tenant_configuration" "storefront_host" {
  key   = "storefront.host"
  value = jsonencode("example.com")
}

resource "emporix_tenant_configuration" "storefront_page" {
  key   = "storefront.htmlPage"
  value = jsonencode("index.html")
}

resource "emporix_tenant_configuration" "email_from" {
  key   = "cust.notification.email.from"
  value = jsonencode("[email protected]")
}
```

### Using for_each

```terraform
locals {
  configurations = {
    "project_country" = jsonencode("US")
    "project_lang"    = jsonencode([{
      id       = "en"
      label    = "English"
      default  = true
      required = true
    }])
    "storefront.host" = jsonencode("example.com")
  }
}

resource "emporix_tenant_configuration" "configs" {
  for_each = local.configurations

  key   = each.key
  value = each.value
}
```

## Schema

### Required

- `key` (String) Configuration key (unique identifier). Cannot be changed after creation. Changing this forces a new resource to be created.
- `value` (String) Configuration value as JSON string. Can be any valid JSON: object, string, array, or boolean. Use `jsonencode()` to convert Terraform values to JSON strings.

### Optional

- `secured` (Boolean) Flag indicating whether the configuration should be encrypted. Defaults to `false`. Set to `true` for sensitive data like API keys or secrets.

### Read-Only

- `version` (Number) Configuration version (managed by API). Used for optimistic locking during updates.

## Import

You can import existing tenant configurations:

```shell
terraform import emporix_tenant_configuration.tax_config taxConfiguration
```

After importing, Terraform will manage the configuration. Note that you'll need to provide the `value` field in your configuration file after import.

## Using jsonencode()

Terraform's `jsonencode()` function converts Terraform values to JSON strings:

```terraform
# String
value = jsonencode("simple string")
# Result: "\"simple string\""

# Number
value = jsonencode(42)
# Result: "42"

# Boolean
value = jsonencode(true)
# Result: "true"

# Object
value = jsonencode({
  key1 = "value1"
  key2 = 123
  key3 = true
})
# Result: "{\"key1\":\"value1\",\"key2\":123,\"key3\":true}"

# Array
value = jsonencode(["item1", "item2", "item3"])
# Result: "[\"item1\",\"item2\",\"item3\"]"
```

### JSON String Values (Double Encoding)

Some Emporix configurations expect the value to be a **JSON string** (not a direct object or array). For example, `project_lang` expects:

```json
{
  "key": "project_lang",
  "value": "[{\"id\":\"en\",\"label\":\"English\",\"default\":true,\"required\":true}]"
}
```

Note that the `value` is a **string** containing JSON, not a direct array. To achieve this in Terraform, use **double jsonencode()**:

```terraform
# WRONG - Direct array (value would be an array, not a string)
resource "emporix_tenant_configuration" "wrong" {
  key = "project_lang"
  value = jsonencode([
    {id = "en", label = "English", default = true, required = true}
  ])
}

# CORRECT - JSON string value (value is a string containing JSON)
resource "emporix_tenant_configuration" "correct" {
  key = "project_lang"
  value = jsonencode(jsonencode([
    {id = "en", label = "English", default = true, required = true}
  ]))
}

# CLEANER - Using locals for readability
locals {
  languages = [
    {id = "en", label = "English", default = true, required = true},
    {id = "de", label = "German", default = false, required = false}
  ]
  languages_json_string = jsonencode(local.languages)
}

resource "emporix_tenant_configuration" "project_lang" {
  key   = "project_lang"
  value = jsonencode(local.languages_json_string)
}
```

**When to use double jsonencode():**
- ✅ `project_lang` - Languages configuration (JSON string)
- ✅ `project_curr` - Currencies configuration (JSON string)
- ✅ Any configuration where the API expects a JSON string value

**When to use single jsonencode():**
- ✅ `project_country` - Simple string value
- ✅ `taxConfiguration` - Direct object value
- ✅ Most other configurations


## Common Configuration Keys

### Project Settings
- `project_country` - Default country code
- `project_lang` - Available languages configuration
- `project_curr` - Available currencies configuration

### Storefront Settings
- `storefront.host` - Storefront host domain
- `storefront.htmlPage` - Default HTML page

### Customer Settings
- `customer.deletion.redirecturl` - URL for customer deletion confirmation
- `customer.passwordreset.redirecturl` - URL for password reset
- `customer.changeemail.redirecturl` - URL for email change confirmation
- `cust.notification.email.from` - From email address for notifications

### Configuration Settings
- `taxConfiguration` - Tax class configuration
- `unitConf` - Unit conversion configuration
- `packagingConf` - Packaging configuration

## Working with Different Value Types

### String Values
```terraform
resource "emporix_tenant_configuration" "simple_string" {
  key   = "my_setting"
  value = jsonencode("plain text value")
}
```

### Numeric Values
```terraform
resource "emporix_tenant_configuration" "timeout" {
  key   = "api_timeout_seconds"
  value = jsonencode(30)
}
```

### Boolean Values
```terraform
resource "emporix_tenant_configuration" "enabled" {
  key   = "feature_enabled"
  value = jsonencode(false)
}
```

### Object Values
```terraform
resource "emporix_tenant_configuration" "settings" {
  key = "app_settings"
  value = jsonencode({
    timeout    = 30
    retries    = 3
    debug_mode = false
    endpoints = {
      api = "https://api.example.com"
      cdn = "https://cdn.example.com"
    }
  })
}
```

### Array Values
```terraform
resource "emporix_tenant_configuration" "allowed_ips" {
  key   = "firewall.whitelist"
  value = jsonencode(["192.168.1.1", "192.168.1.2", "10.0.0.0/8"])
}
```

## Best Practices

### 1. Use Descriptive Keys
```terraform
# Good
resource "emporix_tenant_configuration" "email_from" {
  key   = "cust.notification.email.from"
  value = jsonencode("[email protected]")
}

# Avoid
resource "emporix_tenant_configuration" "config1" {
  key   = "c1"
  value = jsonencode("[email protected]")
}
```

### 2. Secure Sensitive Data
```terraform
# Always set secured = true for sensitive data
resource "emporix_tenant_configuration" "api_secret" {
  key     = "external_api_secret"
  value   = jsonencode(var.api_secret)  # Use variables for secrets
  secured = true
}
```

### 3. Use Variables for Environment-Specific Values
```terraform
variable "environment" {
  type = string
}

resource "emporix_tenant_configuration" "api_url" {
  key   = "api.base_url"
  value = jsonencode(var.environment == "prod" ? "https://api.prod.com" : "https://api.dev.com")
}
```

### 4. Validate JSON Structure
```terraform
# Use try() and can() to validate
locals {
  config_value = {
    setting = "value"
    count   = 10
  }
  
  # Validate it can be encoded
  validated_value = can(jsonencode(local.config_value)) ? jsonencode(local.config_value) : jsonencode({})
}

resource "emporix_tenant_configuration" "config" {
  key   = "my_config"
  value = local.validated_value
}
```

## Delete Warning

⚠️ **Important:** Deleting a tenant configuration permanently removes it from your tenant. Make sure the configuration is not actively used by your applications before deleting it.

## Notes

- Configuration keys are immutable. If you need to change a key, you must destroy and recreate the resource.
- The `value` field must be valid JSON. Use Terraform's `jsonencode()` function to ensure proper formatting.
- Version is automatically managed by the API and used for optimistic locking during updates.
- Secured configurations have their values encrypted at rest in the Emporix backend.
- Changes to secured flag require updating the configuration.

## Accessing Configuration Values

After creating configurations, you can reference them in outputs:

```terraform
resource "emporix_tenant_configuration" "country" {
  key   = "project_country"
  value = jsonencode("US")
}

output "country_config_key" {
  value = emporix_tenant_configuration.country.key
}

output "country_config_value" {
  value = jsondecode(emporix_tenant_configuration.country.value)
}

output "country_config_version" {
  value = emporix_tenant_configuration.country.version
}
```

## Troubleshooting

### Invalid JSON Error

If you get "Invalid JSON" error, validate your JSON:

```terraform
# Wrong - missing quotes
value = jsonencode(invalid)

# Correct
value = jsonencode("invalid")  # String
# or
value = jsonencode(var.invalid)  # Variable
```

### Version Conflict (409)

If you get a 409 Conflict error, it means another process updated the configuration. Refresh your state:

```bash
terraform refresh
terraform apply
```

### Configuration Not Found (404)

If configuration doesn't exist during read/update, it may have been deleted outside Terraform:

```bash
# Remove from state
terraform state rm emporix_tenant_configuration.my_config

# Re-import if it exists
terraform import emporix_tenant_configuration.my_config my_config_key
```
