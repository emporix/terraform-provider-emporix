# Terraform Provider for Emporix

This is a Terraform provider for managing Emporix resources.

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

## Testing

This provider includes comprehensive acceptance tests for all resources.

### Quick Start

```bash
# Install dependencies
go mod download

# Set credentials
export TF_ACC=1
export EMPORIX_TENANT="your-test-tenant"
export EMPORIX_CLIENT_ID="your-client-id"
export EMPORIX_CLIENT_SECRET="your-client-secret"

# Run all tests
make testacc

# Run specific resource tests
make testacc-country
make testacc-paymentmode
make testacc-sitesettings
```

### Test Files

- `internal/provider/resource_country_test.go` - Country resource tests
- `internal/provider/resource_paymentmode_test.go` - Payment mode tests
- `internal/provider/resource_sitesettings_test.go` - Site settings tests

### Requirements

- Go 1.21+
- Valid Emporix test tenant and credentials
- OAuth scopes: `site.site_read`, `site.site_manage`, `payment.payment_manage`, `payment.payment_read`, `country.country_read`, `country.country_manage`

**Important:** Always use a test tenant, never production!
