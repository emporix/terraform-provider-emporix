# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2024-12-16

### Added

- **New Resource: emporix_currency** - Manage currencies in Emporix
  - Full CRUD operations
  - ISO-4217 compliant codes
- **New Resource: emporix_tenant_configuration** - Manage tenant configurations in Emporix
  - Full CRUD operations support (Create, Read, Update, Delete)
  - Store key-value pairs where values can be any valid JSON (object, string, array, or boolean)
  - Configuration key structure:
    - Immutable keys (changing key requires resource replacement)
    - Support for dotted notation (e.g., `customer.passwordreset.redirecturl`)

[0.3.0]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.3.0

## [0.2.0] - 2024-12-08

### Added

- **New Resource: emporix_country** - Manage country active status
  - Read and Update operations (countries are pre-populated, cannot be created/deleted)
  - Only `active` field can be modified
  - Supports ISO 3166-1 alpha-2 country codes (2-letter codes)

- **New Resource: emporix_paymentmode** - Manage payment mode configurations
  - Full CRUD operations support (Create, Read, Update, Delete)
  - Support for multiple payment providers:
    - INVOICE - Simple invoice payment
    - CASH_ON_DELIVERY - Cash on delivery payment

[0.2.0]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.2.0

## [0.1.2] - 2024-12-05

### Added

- **`emporix_sitesettings` resource - Optional Fields Nulling Support** - All optional fields can now be set back to null by removing them from Terraform configuration
  - Allows removing fields that were added via UI or API
  - Explicitly sends `null` in PATCH requests when fields are removed from config
  - Supports null transitions: `null → value → null`
  - Distinguishes between `0` and `null` (e.g., location coordinates)
  - Works for all optional fields:
    - Top-level: `includes_tax`, `available_currencies`, `tax_calculation_address_type`, `decimal_points`, `cart_calculation_scale`
    - Nested in `home_base`: `location` (entire object), `address.street`, `address.street_number`, `address.state`
    - Nested in `assisted_buying`: entire object and `storefront_url` field
    - `mixins`: entire list

- **Performance improvements**
  - HTTP connection pooling (client reuse)
  - 30-second request timeout configured
  - Context propagation for proper cancellation
  - Reduced memory allocations

[0.1.2]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.1.2

## [0.1.1] - 2024-11-24

### Added
- Initial release of the Emporix Terraform Provider
- Support for managing Emporix Site Settings via `emporix_sitesettings` resource
- Full CRUD operations (Create, Read, Update, Delete) for site settings
- Import support for existing sites
- Environment variable configuration support
- Comprehensive site configuration including:
  - Basic settings (code, name, active, default)
  - Language and currency configuration
  - Tax and pricing settings
  - Home base with address and location
  - Assisted buying configuration
  - Custom mixins support

### Technical Details
- Built with Terraform Plugin Framework v1.15.0 (latest stable)
- Uses Protocol version 6 for modern Terraform features
- Go 1.21 support (stable, widely compatible)
- Compatible with Terraform >= 1.0

### Provider Configuration
- Registry address: `registry.terraform.io/emporix/emporix`
- Supports authentication via access token
- Configurable API URL (defaults to https://api.emporix.io)

[0.1.1]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.1.1
