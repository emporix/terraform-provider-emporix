---
page_title: "Provider Configuration Guide"
subcategory: "Getting Started"
description: |-
  Learn how to configure the Emporix Terraform Provider with proper authentication and credentials.
---

# Provider Configuration Guide

This guide explains how to configure the Emporix Terraform Provider for managing your Emporix resources.

## Prerequisites

Before configuring the provider, you need:

1. **Emporix Account** - An active Emporix tenant
2. **API Keys** - OAuth2 client credentials or access token
3. **Terraform** - Version 1.0 or later

## Getting Your Credentials

### Create Tenant

Follow the instructions: https://developer.emporix.io/ce/getting-started/creating-a-tenant

## Provider Configuration

### Basic Configuration

The simplest configuration using client credentials (recommended):

```terraform
terraform {
  required_providers {
    emporix = {
      source  = "emporix/emporix"
      version = "<provider version>"
    }
  }
}

provider "emporix" {
  tenant        = "your-tenant-name"
  client_id     = "your-client-id"
  client_secret = "your-client-secret"
}
```

## Authentication Methods

The provider supports two authentication methods:

### Method 1: Client Credentials (Recommended)

Uses OAuth2 client credentials flow. The provider automatically obtains and refreshes access tokens.

```terraform
provider "emporix" {
  tenant        = "your-tenant"
  client_id     = "abc123"
  client_secret = "secret456"
}
```

**Advantages:**
- ✅ Automatic token refresh
- ✅ Long-lived credentials
- ✅ Suitable for automation
- ✅ Best for CI/CD pipelines

#### Using Custom API Keys (Strongly Recommended)

For better security, it's highly recommended to use **Custom API Keys** instead of the Management API key. Custom API keys allow you to:

- ✅ **Limit access to specific scopes** - Only grant the permissions your Terraform configuration needs
- ✅ **Separate concerns** - Different keys for different purposes (e.g., separate keys for country management, currency management, etc.)
- ✅ **Enhanced security** - If a key is compromised, only specific resources are at risk

**How to create Custom API Keys:**

See [Emporix Custom API Keys Documentation](https://developer.emporix.io/ce/getting-started/developer-portal/manage-apikeys#custom-api-keys)

**Example: Creating a Custom API Key for Currency Management**

If you're only managing currencies with Terraform:

1. Create a custom API key with scopes:
   - `currency.currency_read`
   - `currency.currency_manage`
2. Use this key in your provider configuration:

```terraform
provider "emporix" {
  tenant        = "your-tenant"
  client_id     = "custom-key-client-id"      # From your custom API key
  client_secret = "custom-key-secret"         # From your custom API key
}
```

**Note:** Each resource's documentation lists the required scopes. Create custom API keys with only the scopes you need for your specific use case.

### Method 2: Pre-Generated Access Token

Uses a manually generated access token. Tokens typically expire after a set period.

```terraform
provider "emporix" {
  tenant       = "your-tenant"
  access_token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Use cases:**
- Testing and development
- Short-term operations
- When client credentials aren't available

**Note:** Tokens expire and must be manually refreshed.

## Secure Credential Management

**Never commit credentials to version control!** Use one of these methods:

### Option 1: Environment Variables (Recommended)

Set environment variables:

```bash
export EMPORIX_TENANT="your-tenant"
export EMPORIX_CLIENT_ID="your-client-id"
export EMPORIX_CLIENT_SECRET="your-client-secret"
```

Configure provider to read from environment:

```terraform
provider "emporix" {
  # Reads from EMPORIX_TENANT
  # Reads from EMPORIX_CLIENT_ID
  # Reads from EMPORIX_CLIENT_SECRET
}
```

### Option 2: Terraform Variables

Create `terraform.tfvars` (add to `.gitignore`):

```hcl
emporix_tenant        = "your-tenant"
emporix_client_id     = "your-client-id"
emporix_client_secret = "your-client-secret"
```

Define variables in `variables.tf`:

```terraform
variable "emporix_tenant" {
  description = "Emporix tenant name"
  type        = string
  sensitive   = false
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
```

Use variables in provider configuration:

```terraform
provider "emporix" {
  tenant        = var.emporix_tenant
  client_id     = var.emporix_client_id
  client_secret = var.emporix_client_secret
}
```

### Option 3: Secrets Manager

Use a secrets manager (AWS Secrets Manager, HashiCorp Vault, etc.):

```terraform
data "aws_secretsmanager_secret_version" "emporix" {
  secret_id = "emporix/credentials"
}

locals {
  emporix_creds = jsondecode(data.aws_secretsmanager_secret_version.emporix.secret_string)
}

provider "emporix" {
  tenant        = local.emporix_creds.tenant
  client_id     = local.emporix_creds.client_id
  client_secret = local.emporix_creds.client_secret
}
```

## Remote State Management

Note: the Emporix provider does **not** persist `client_id`, `client_secret`, or `access_token` into the state file. These are declared `Sensitive` in the provider schema and used only in-memory to obtain an API client — they never appear in any resource's state attributes. That said, the general best practices below (locking, encryption, access control) still apply: the state file is not your `.tf` configuration, but it does store the ID and current attribute values of every resource you manage (which can include sensitive resource attributes), so it is not something to store or share casually.

### Why Use Remote State

**Avoid relying on local state (`terraform.tfstate`) for anything beyond quick experiments:**

- ⚠️ **No locking** - Concurrent `terraform apply` runs from different machines/CI jobs can corrupt state or apply conflicting changes
- ⚠️ **No collaboration** - Team members and CI/CD pipelines need a shared, consistent view of state
- ⚠️ **No durability** - A local file can be lost, accidentally deleted, or diverge between team members' machines

### Recommended Backends

Use a remote backend that supports encryption at rest and state locking:

**Terraform Cloud / HCP Terraform (Recommended for most teams):**

```terraform
terraform {
  cloud {
    organization = "your-org"
    workspaces {
      name = "emporix-production"
    }
  }
}
```

- ✅ Built-in state encryption, locking, and versioning
- ✅ Run history and audit log
- ✅ Can inject `EMPORIX_*` credentials as workspace-level sensitive environment variables, so they never appear in `.tf` files or CI logs

**AWS S3 + DynamoDB (state locking):**

```terraform
terraform {
  backend "s3" {
    bucket         = "mycompany-terraform-state"
    key            = "emporix/production/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-state-lock"
  }
}
```

- ✅ `encrypt = true` enables server-side encryption (use a customer-managed KMS key for stricter control)
- ✅ `dynamodb_table` provides state locking to prevent concurrent modifications
- ✅ Restrict bucket access via IAM policy to only the users/roles that need it

**Azure Storage:**

```terraform
terraform {
  backend "azurerm" {
    resource_group_name  = "terraform-state-rg"
    storage_account_name = "mycompanytfstate"
    container_name       = "emporix-state"
    key                  = "production.terraform.tfstate"
  }
}
```

**Google Cloud Storage:**

```terraform
terraform {
  backend "gcs" {
    bucket = "mycompany-terraform-state"
    prefix = "emporix/production"
  }
}
```

### State Security Checklist

1. **Enable encryption at rest** - All the backends above support this; always turn it on
2. **Enable state locking** - Prevents two `terraform apply` runs from corrupting state concurrently
3. **Restrict access** - Grant read/write access to the state backend only to the CI/CD service principal and operators who need it; state reveals the resource IDs and current attribute values of every managed resource, which can include sensitive data
4. **Enable versioning** - Turn on bucket/container versioning so you can recover from accidental `terraform apply` mistakes or state corruption
5. **Never commit `.tfstate` files to version control** - Add `*.tfstate` and `*.tfstate.backup` to `.gitignore` (already shown in the [Complete Example](#gitignore) below); this applies even more strictly to remote-state configurations, since a locally downloaded state file is just as sensitive
6. **Isolate state per environment** - Use separate state files (or workspaces) for dev/staging/production so a mistake in one environment can't affect another; see [Multi-Environment Setup](#multi-environment-setup)
7. **Avoid `terraform state pull` / `terraform show` in shared logs** - These commands print the full state to stdout, which may include sensitive attributes on managed resources

### Migrating from Local to Remote State

If you started with local state, migrate safely:

```bash
# 1. Add a backend block to your terraform block (as shown above)
# 2. Re-initialize - Terraform will prompt to migrate existing state
terraform init

# 3. Verify the migration
terraform state list
```

Terraform will ask for confirmation before copying local state to the new backend. Keep a backup of `terraform.tfstate` until you've confirmed the migration succeeded.

## Provider Configuration Reference

### Arguments

All provider arguments are optional if corresponding environment variables are set.

| Argument | Environment Variable | Type | Required | Description |
|----------|---------------------|------|----------|-------------|
| `tenant` | `EMPORIX_TENANT` | string | Yes* | Emporix tenant identifier |
| `client_id` | `EMPORIX_CLIENT_ID` | string | Yes** | OAuth2 client ID |
| `client_secret` | `EMPORIX_CLIENT_SECRET` | string | Yes** | OAuth2 client secret |
| `access_token` | `EMPORIX_ACCESS_TOKEN` | string | Yes*** | Pre-generated access token |

\* Required for all authentication methods  
\** Required when using client credentials authentication  
\*** Required when using access token authentication

### Authentication Precedence

If multiple authentication methods are configured, the provider uses this precedence:

1. **Access Token** (if provided)
2. **Client Credentials** (if both client_id and client_secret are provided)

## Complete Example

### Project Structure

```
my-terraform-project/
├── .gitignore
├── main.tf
├── variables.tf
├── terraform.tfvars  (gitignored)
└── outputs.tf
```

### .gitignore

```
# Terraform
.terraform/
.terraform.lock.hcl
*.tfstate
*.tfstate.backup

# Sensitive files
terraform.tfvars
*.auto.tfvars
```

### variables.tf

```terraform
variable "emporix_tenant" {
  description = "Emporix tenant name"
  type        = string
  sensitive   = false
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
```

### terraform.tfvars

```hcl
emporix_tenant        = "my-company"
emporix_client_id     = "abc123xyz789"
emporix_client_secret = "secret-value-here"
```

### main.tf

```terraform
terraform {
  required_version = ">= 1.0"
  
  required_providers {
    emporix = {
      source  = "YOUR_NAMESPACE/emporix"
      version = "~> 0.1.0"
    }
  }
}

provider "emporix" {
  tenant        = var.emporix_tenant
  client_id     = var.emporix_client_id
  client_secret = var.emporix_client_secret
}

# Your resources here
resource "emporix_sitesettings" "main" {
  code              = "main-site"
  name              = "Main Site"
  active            = true
  default_language  = "en"
  languages         = ["en"]
  currency          = "USD"
  ship_to_countries = ["US"]
  
  home_base = {
    address = {
      country  = "US"
      zip_code = "10001"
      city     = "New York"
    }
  }
}
```

### outputs.tf

```terraform
output "site_code" {
  description = "The site code"
  value       = emporix_sitesettings.main.code
}
```

## CI/CD Configuration

### GitHub Actions

```yaml
name: Terraform

on:
  push:
    branches: [main]
  pull_request:

jobs:
  terraform:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        
      - name: Terraform Init
        run: terraform init
        env:
          EMPORIX_TENANT: ${{ secrets.EMPORIX_TENANT }}
          EMPORIX_CLIENT_ID: ${{ secrets.EMPORIX_CLIENT_ID }}
          EMPORIX_CLIENT_SECRET: ${{ secrets.EMPORIX_CLIENT_SECRET }}
          
      - name: Terraform Plan
        run: terraform plan
        env:
          EMPORIX_TENANT: ${{ secrets.EMPORIX_TENANT }}
          EMPORIX_CLIENT_ID: ${{ secrets.EMPORIX_CLIENT_ID }}
          EMPORIX_CLIENT_SECRET: ${{ secrets.EMPORIX_CLIENT_SECRET }}
```

### GitLab CI

```yaml
variables:
  TF_ROOT: ${CI_PROJECT_DIR}

terraform:
  image: hashicorp/terraform:latest
  script:
    - terraform init
    - terraform plan
    - terraform apply -auto-approve
  only:
    - main
```

**GitLab CI/CD Variables:**
- `EMPORIX_TENANT`
- `EMPORIX_CLIENT_ID` (protected, masked)
- `EMPORIX_CLIENT_SECRET` (protected, masked)

## Troubleshooting

### Authentication Fails

**Error:** `Failed to authenticate with Emporix API`

**Solutions:**
1. Verify credentials are correct
2. Check tenant name matches your account
3. Ensure client has proper scopes
4. Verify network connectivity to `api.emporix.io`

### Token Expired

**Error:** `Access token has expired`

**Solutions:**
- If using `access_token`: Generate a new token
- If using client credentials: Provider should auto-refresh (check credentials)

### Missing Permissions

**Error:** `Insufficient permissions`

**Solutions:**
1. Check credentials scopes in Developer Portal (https://app.emporix.io)

## Best Practices

1. **Use Client Credentials** - Preferred for automation and production
2. **Environment Variables** - Best for local development
3. **Secrets Management** - Use proper secrets managers in production
4. **Least Privilege** - Grant minimum required scopes
5. **Credential Rotation** - Regularly rotate client secrets
6. **Never Commit Secrets** - Always use `.gitignore`
7. **Separate Tenants** - Use different tenants for dev/staging/production
8. **Use Remote State** - Store state in an encrypted, locked remote backend rather than locally; see [Remote State Management](#remote-state-management)

## Multi-Environment Setup

### Development

```terraform
# dev.tfvars
emporix_tenant = "mycompany-dev"
```

### Staging

```terraform
# staging.tfvars
emporix_tenant = "mycompany-staging"
```

### Production

```terraform
# prod.tfvars
emporix_tenant = "mycompany-prod"
```

Usage:
```bash
terraform plan -var-file=dev.tfvars
terraform apply -var-file=prod.tfvars
```

## Next Steps

- Explore [Resources Documentation](https://registry.terraform.io/providers/emporix/emporix/latest/docs/resources/sitesettings)
- Review [Examples](https://github.com/emporix/terraform-provider-emporix/tree/master/examples)
- Check [Emporix API Documentation](https://developer.emporix.io/api-references)

## Support

For issues or questions:
- Provider Issues: support@emporix.com
- [Emporix Documentation](https://developer.emporix.io)