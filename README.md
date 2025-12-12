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

**⚠️ IMPORTANT:** After extracting the provider, run setup first to initialize the module dependencies.

### Quick Start

```bash
# 1. First-time setup (required after extracting)
make setup

# 2. Build the provider
make build

# 3. Install locally for development
make install
```

### Manual Build (without Makefile)

```bash
# 1. Initialize module dependencies
go mod tidy

# 2. Build
go build -o terraform-provider-emporix

# 3. Install (optional)
OS_ARCH=$(go env GOOS)_$(go env GOARCH)
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/emporix/emporix/0.1.0/${OS_ARCH}
cp terraform-provider-emporix ~/.terraform.d/plugins/registry.terraform.io/emporix/emporix/0.1.0/${OS_ARCH}/
```

### Troubleshooting Build Issues

If you see an error like `package terraform-provider-emporix/internal/resources is not in std`:

```bash
# Quick fix
make setup
make build

# Or manually
go mod tidy && go build
```

## Installing the Provider Locally

For local development, the easiest way is to use the Makefile:

```bash
make install
```

This will:
1. Build the provider
2. Detect your OS and architecture
3. Install to the correct Terraform plugins directory
4. Set proper permissions

### Manual Installation

If you prefer to install manually:

```bash
# Build the provider
go build -o terraform-provider-emporix

# Determine your OS and architecture
OS_ARCH=$(go env GOOS)_$(go env GOARCH)

# Create the plugins directory structure
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/emporix/emporix/0.1.0/${OS_ARCH}

# Copy the provider
cp terraform-provider-emporix ~/.terraform.d/plugins/registry.terraform.io/emporix/emporix/0.1.0/${OS_ARCH}/

# Make it executable
chmod +x ~/.terraform.d/plugins/registry.terraform.io/emporix/emporix/0.1.0/${OS_ARCH}/terraform-provider-emporix
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


## Support

For issues related to:
- The provider: Create an issue in this repository
- Emporix API: Refer to [Emporix Developer Portal](https://developer.emporix.io)

## Testing

This provider includes comprehensive acceptance tests for all resources.

### Quick Start

```bash
# 1. Run automated setup
make clean
make setup

# 2. Configure credentials
# create .env.test file with credentials

# 3. Load credentials and run tests
source .env.test
make testacc

# Run specific resource tests
make testacc-country
make testacc-paymentmode
make testacc-sitesettings
```

### Manual Setup (Alternative)

```bash
# Install dependencies
go mod download
go mod tidy

# Set credentials
export TF_ACC=1
export EMPORIX_TENANT="your-test-tenant"
export EMPORIX_CLIENT_ID="your-client-id"
export EMPORIX_CLIENT_SECRET="your-client-secret"

# Run tests
make testacc
```

### Requirements

- Go 1.23+
- Valid Emporix test tenant and credentials
- OAuth scopes: `site.site_read`, `site.site_manage`, `payment.payment_manage`, `payment.payment_read`, `country.country_read`, `country.country_manage`

**Important:** Always use a test tenant, never production!

### Dependency Versions

This provider uses the latest stable Terraform plugin dependencies (December 2024):
- `terraform-plugin-framework v1.15.0`
- `terraform-plugin-go v0.29.0`
- `terraform-plugin-log v0.9.0`
- `terraform-plugin-testing v1.14.0`

**Requirements:**
- Go 1.23 or higher (required by latest plugin packages)
