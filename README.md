# Terraform Provider for Emporix

This is a Terraform provider for managing Emporix Site Settings resources.

Built with the latest **Terraform Plugin Framework v1.15.0** (May 2024).

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21

## Technology Stack

This provider is built using:
- **Terraform Plugin Framework v1.15.0** - The latest stable version (May 2024)
- **Protocol version 6** - For access to the newest Terraform features
- **Go 1.21** - Stable Go version with wide compatibility

### Why Plugin Framework v1.15.0?

The Terraform Plugin Framework is HashiCorp's recommended way to build providers, offering:
- ✅ Enhanced type system with custom types and nested attributes
- ✅ Better handling of null and unknown values
- ✅ Improved plan modification and validation
- ✅ Resource import by identity support
- ✅ Protocol version 6 for Terraform 1.0+ features
- ✅ Compatible with Terraform 0.12+ (though 1.0+ recommended)

## Building the Provider

1. Clone the repository
2. Navigate to the provider directory
3. First-time setup (regenerate go.sum with correct checksums):

```bash
# Option 1: Use the setup script
chmod +x scripts/setup.sh
./scripts/setup.sh

# Option 2: Manual setup
rm -f go.sum
go mod download
go mod tidy
```

4. Build the provider:

```bash
go build -o terraform-provider-emporix
```

## Installing the Provider Locally

For local development, you can install the provider in your local Terraform plugins directory:

```bash
# Build the provider
go build -o terraform-provider-emporix

# Determine your OS and architecture
OS_ARCH=$(go env GOOS)_$(go env GOARCH)

# Create the plugins directory structure
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/emporix/emporix/0.1.0/${OS_ARCH}

# Copy the provider
cp terraform-provider-emporix ~/.terraform.d/plugins/registry.terraform.io/emporix/emporix/0.1.0/${OS_ARCH}/
```

Or use the Makefile which handles this automatically:
```bash
make install
```

Common OS/Architecture combinations:
- Linux: `linux_amd64`
- macOS Intel: `darwin_amd64`
- macOS Apple Silicon: `darwin_arm64`
- Windows: `windows_amd64`

## Usage

### Provider Configuration

The provider supports two authentication methods:

#### Method 1: Client Credentials (Recommended)

The provider automatically generates an access token using your OAuth2 client credentials:

```hcl
terraform {
  required_providers {
    emporix = {
      source  = "emporix/emporix"
      version = "~> 0.1"
    }
  }
}

provider "emporix" {
  tenant        = "your-tenant-name"
  client_id     = "your-client-id"
  client_secret = "your-client-secret"
  # Optional: scopes (if not provided, no scope parameter is sent)
  # scope         = "tenant=your-tenant-name site.site_read site.site_manage"
}
```

#### Method 2: Pre-generated Access Token

If you already have an access token:

```hcl
provider "emporix" {
  tenant       = "your-tenant-name"
  access_token = "your-access-token"
  api_url      = "https://api.emporix.io" # Optional, defaults to this
}
```

### Environment Variables

Instead of hardcoding credentials in your Terraform configuration, you can use environment variables:

```bash
# Method 1: Using client credentials (recommended)
export EMPORIX_TENANT="your-tenant-name"
export EMPORIX_CLIENT_ID="your-client-id"
export EMPORIX_CLIENT_SECRET="your-client-secret"
# Optional: scopes (if not set, no scope parameter is sent)
# export EMPORIX_SCOPE="tenant=your-tenant-name site.site_read site.site_manage"

# Method 2: Using pre-generated token
export EMPORIX_TENANT="your-tenant-name"
export EMPORIX_ACCESS_TOKEN="your-access-token"

# Optional for both methods
export EMPORIX_API_URL="https://api.emporix.io"
```

### Configuration Reference

| Attribute | Description | Environment Variable | Required |
|-----------|-------------|---------------------|----------|
| `tenant` | Emporix tenant name (lowercase) | `EMPORIX_TENANT` | Yes |
| `client_id` | OAuth2 client ID for token generation | `EMPORIX_CLIENT_ID` | If no `access_token` |
| `client_secret` | OAuth2 client secret for token generation | `EMPORIX_CLIENT_SECRET` | If no `access_token` |
| `access_token` | Pre-generated OAuth2 access token | `EMPORIX_ACCESS_TOKEN` | If no client credentials |
| `scope` | OAuth2 scopes (space-separated) | `EMPORIX_SCOPE` | No |
| `api_url` | Emporix API base URL | `EMPORIX_API_URL` | No (defaults to `https://api.emporix.io`) |

**Note:** 
- You must provide either `access_token` OR both `client_id` and `client_secret`. 
- The provider will automatically generate a token from client credentials if no access token is provided.
- If `scope` is not provided, no scope parameter is sent in the OAuth request.

### Resource: emporix_sitesettings

#### Basic Example

```hcl
resource "emporix_sitesettings" "example" {
  code             = "example-site"
  name             = "Example Site"
  active           = true
  default          = false
  default_language = "en"
  languages        = ["en", "de", "fr"]
  currency         = "USD"
  
  ship_to_countries = ["US", "CA"]
  
  home_base = {
    address = {
      country       = "US"
      zip_code      = "10001"
      city          = "New York"
      street        = "Broadway"
      street_number = "123"
      state         = "NY"
    }
    location = {
      latitude  = 40.7589
      longitude = -73.9851
    }
  }
}
```

#### Full Example with All Options

```hcl
resource "emporix_sitesettings" "full_example" {
  code                          = "full-example"
  name                          = "Full Example Site"
  active                        = true
  default                       = false
  includes_tax                  = true
  default_language              = "en"
  languages                     = ["en", "de", "fr", "es"]
  currency                      = "EUR"
  available_currencies          = ["EUR", "USD", "GBP"]
  ship_to_countries             = ["US", "GB", "DE", "FR"]
  tax_calculation_address_type  = "SHIPPING_ADDRESS"
  decimal_points                = 2
  
  home_base = {
    address = {
      country       = "US"
      zip_code      = "10036"
      city          = "New York"
      street        = "Broadway"
      street_number = "1500"
      state         = "NY"
    }
    location = {
      latitude  = 40.7568658044745
      longitude = -73.9858713458565
    }
  }
  
  assisted_buying = {
    storefront_url = "https://example.com/store"
  }
  
  mixins = {
    productSettings = "https://example.com/mixins/product-settings.json"
    customSettings  = "https://example.com/mixins/custom-settings.json"
  }
}
```

### Importing Existing Sites

You can import existing Emporix sites into Terraform:

```bash
terraform import emporix_sitesettings.example example-site-code
```

## Resource Schema

### Required Arguments

- `code` - (String) Site unique identifier (code). Cannot be changed after creation.
- `name` - (String) Site name.
- `default_language` - (String) Site's default language (ISO 639-1, 2-letter lowercase code).
- `languages` - (List of String) Languages supported by the site (ISO 639-1).
- `currency` - (String) Currency used by the site (ISO 4217, 3-letter uppercase code).

### Optional Arguments

- `active` - (Boolean) Flag indicating whether the site is active. Defaults to `true`.
- `default` - (Boolean) Flag indicating whether the site is the tenant default site. Defaults to `false`.
- `includes_tax` - (Boolean) Whether prices should be returned in gross (true) or net (false).
- `available_currencies` - (List of String) List of currencies supported by the site.
- `ship_to_countries` - (List of String) Country codes to which the site ships (ISO 3166-1 alpha-2).
- `tax_calculation_address_type` - (String) Tax calculation basis. Defaults to `BILLING_ADDRESS`.
- `decimal_points` - (Number) Decimal points used in cart calculation. Defaults to `2`.
- `home_base` - (Object) Home base configuration.
  - `address` - (Object) Address information.
    - `country` - (String, Required) Country code.
    - `zip_code` - (String) Postal code.
    - `city` - (String) City name.
    - `street` - (String) Street name.
    - `street_number` - (String) Street number.
    - `state` - (String) State or province.
  - `location` - (Object) Geographic coordinates.
    - `latitude` - (Number) Latitude.
    - `longitude` - (Number) Longitude.
- `assisted_buying` - (Object) Assisted buying configuration.
  - `storefront_url` - (String) Storefront URL.
- `mixins` - (Map of String) Custom mixins for extending site configuration.

## API Authentication

To use this provider, you need an OAuth2 access token from Emporix. The token should have the following scopes:

- `site.site_read` - For reading site information
- `site.site_manage` - For creating, updating, and deleting sites

Refer to the [Emporix API documentation](https://developer.emporix.io) for information on obtaining access tokens.

## Development

### Running Tests

```bash
go test -v ./...
```

### Building for Multiple Platforms

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o terraform-provider-emporix_linux_amd64

# macOS
GOOS=darwin GOARCH=amd64 go build -o terraform-provider-emporix_darwin_amd64

# Windows
GOOS=windows GOARCH=amd64 go build -o terraform-provider-emporix_windows_amd64.exe
```

## License

This provider is provided as-is for use with the Emporix platform.

## Troubleshooting

### Checksum Mismatch Error

If you get an error like:
```
verifying github.com/hashicorp/terraform-plugin-framework@v1.4.2: checksum mismatch
SECURITY ERROR
```

This happens because the go.sum file needs to be regenerated with the correct checksums for your system.

**Solution:**
```bash
# Use the Makefile target
make setup

# Or manually
rm -f go.sum
go mod download
go mod tidy

# Then build
make build
```

### "Unrecognized remote plugin message" Error

If you get this error after running `terraform apply`:
```
failed to instantiate provider "registry.terraform.io/emporix/emporix" to obtain schema: Unrecognized remote plugin message
```

This usually means one of the following:

1. **The provider wasn't rebuilt after changes**: Run `make clean && make install` to rebuild and reinstall
2. **Old cached files**: Remove `.terraform` and `.terraform.lock.hcl`, then run `terraform init` again
3. **Go dependencies issue**: Run `go mod download && go mod tidy` before building

**Solution:**
```bash
# Clean everything
make clean
rm -rf .terraform .terraform.lock.hcl

# Rebuild dependencies
go mod download
go mod tidy

# Rebuild and reinstall
make install

# Reinitialize Terraform
terraform init
```

### Provider Not Found

If Terraform can't find your provider, make sure:
1. The binary is named `terraform-provider-emporix` in the plugins directory
2. It's in the correct path: `~/.terraform.d/plugins/registry.terraform.io/emporix/emporix/0.1.0/<OS_ARCH>/`
3. Your Terraform configuration uses the correct source: `source = "emporix/emporix"`

## Support

For issues related to:
- The provider: Create an issue in this repository
- Emporix API: Refer to [Emporix Developer Portal](https://developer.emporix.io)
