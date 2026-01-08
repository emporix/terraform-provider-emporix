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
terraform.tfstate
terraform.tfstate.backup

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