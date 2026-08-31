# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.11.2] - 2026-08-31

### Added

- **emporix_webhook**
  - Multi-target `events_configuration`: multiple entries may share the same `event_type`, each with a read-only `id`
  - New per-entry attributes: `filter`, `excluded_fields`, `active`, `name`
  - Updates now use per-entry PATCH instead of replacing the whole list

### Fixes

- **emporix_webhook**
  - `SVIX`/`SVIX_SHARED` no longer send fields the API rejects, now validated up front
  - Clearer error when creating an inactive webhook would leave the tenant with none active
  - `terraform destroy` now always force-deletes a webhook, active or not, instead of requiring a separate deactivation step

[0.11.2]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.11.2

## [0.11.1] - 2026-08-31

### Improvements

- **Documentation: emporix_price_models**
  - documented that the tenant's current default price model can never be deleted, even with `force_delete` - a different price model must be assigned as the default first

[0.11.1]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.11.1

## [0.11.0] - 2026-08-24

### Added

- **New Resource: emporix_price_models** - Manage price models in Emporix
  - Full CRUD operations
  - Supports `BASIC`, `VOLUME`, and `TIERED` pricing strategies via `tier_definition`
  - Localized `name`/`description` fields
- **emporix_sitesettings**
  - added `timezone` to the `home_base` block - IANA timezone identifier for the site's home base location (e.g. `Europe/Paris`, `America/New_York`, `UTC`)

[0.11.0]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.11.0

## [0.10.0] - 2026-08-20

### Added

- **New Resource: emporix_custom_entity_type** - Manage custom entities types in Emporix
  - Full CRUD operations
- **New Resource: emporix_custom_entity_instance** - Manage custom entities instances in Emporix
  - Full CRUD operations, plus import using the `type:id` format

[0.10.0]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.10.0

## [0.9.4] - 2026-08-13

### Fixes

- **emporix_sitesettings**
  - fixed non-deterministic ordering of the `mixins` list when more than one mixin was configured, which could cause "Provider produced inconsistent result after apply" errors or spurious reordering diffs on `terraform plan`/`apply`

[0.9.4]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.9.4

## [0.9.3] - 2026-08-10

### Documentation

- **Provider Configuration Guide**
  - added a "Remote State Management" section covering remote backend options (Terraform Cloud/HCP Terraform, S3+DynamoDB, Azure Storage, GCS), a state security checklist, and guidance for migrating from local to remote state

[0.9.3]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.9.3

## [0.9.2] - 2026-07-23

### Fixes

- **emporix_webhook**
  - fixed "Value Conversion Error" during `terraform plan` when `events_configuration` was assigned from a variable or module input whose declared type omitted the Computed attributes (`destination_url`, `secret_key`, `subscribed`); Terraform widens such a value to match the resource schema and marks it `unknown`, which the provider previously couldn't handle in `ValidateConfig`. The same value written as a literal list directly in the resource block was never affected.

[0.9.2]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.9.2

## [0.9.1] - 2026-07-20

### Added

- **emporix_webhook**
  - `events_configuration.subscribed` field to intentionally subscribe/unsubscribe an event type while keeping its `destination_url`/`secret_key`/`headers` overrides configured

### Fixes

- **emporix_webhook**
  - fixed "inconsistent values for sensitive attribute" error by unconditionally refreshing event subscription status from the API on every Create/Read/Update
  - fixed 400 "value must not be empty" error when clearing `events_configuration` or `headers`: empty values are now sent as a `REMOVE` patch operation instead of `UPSERT` with an empty array
  - removing an event from `events_configuration` (or removing the whole block) now correctly unsubscribes it, instead of leaving a stale subscription on the API side
  - `destination_url` on a nested event now correctly falls back to the resource-level `destination_url` at plan time (both when omitted and when explicitly set to `""`), instead of failing with "destinationUrl must not be blank"
  - `secret_key_exists` and `version` no longer show as `(known after apply)` on every update; unrelated changes no longer produce a non-empty plan on the next `apply`
  - a failed event-subscription update no longer aborts the resource `Create`/`Update` after the underlying webhook was already created/updated, which could leave the resource out of Terraform state or apply outdated state

[0.9.1]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.9.1

## [0.9.0] - 2026-07-06

### Added

- **New Resource: emporix_webhook** - Manage webhook configurations in Emporix

[0.9.0]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.9.0

## [0.8.2] - 2026-04-08

### Improvements

- **emporix_schema**
  - add posibility to create schema with array of objects

[0.8.2]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.8.2

## [0.8.1] - 2026-03-06

### Fixes

- **emporix_tax**
  - correct docs for rate field
  - fixed examples (rate field)

[0.8.1]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.8.1

## [0.8.0] - 2026-02-12

### Added

- **New Resource: emporix_tax** - Manage taxes in Emporix
  - Full CRUD operations

[0.8.0]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.8.0

## [0.7.0] - 2026-02-10

### BREAKING CHANGE
- **emporix_schema**
  - unlimited OBJECT nesting (requires resource remove and import to state)

## Changes
- **emporix_schema**
  - make id field optional

## Fixes
- **emporix_sitesettings**
  - make possible to update the mixin fields

[0.7.0]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.7.0

## [0.6.2] - 2026-02-05

### Fixes
- **emporix_schema**
  - minor documentation fixes (correct nesting)

[0.6.2]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.6.2

## [0.6.1] - 2026-02-04

### Improvements

- **emporix_schema**
  - added schema_url to resource output

### Fixes
- **emporix_delivery_time**
  - correct time format (HH:MM) in delivery_time_range section

[0.6.1]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.6.1

## [0.6.0] - 2026-02-02

### Added

- **New Resource: emporix_schema** - Manage mixin schemas in Emporix
  - Full CRUD operations

[0.6.0]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.6.0

## [0.5.0] - 2026-01-29

### Added

- **New Resource: emporix_shipping_method** - Manage shipping methods in Emporix
  - Full CRUD operations
- **New Resource: emporix_delivery_time** - Manage delivery times in Emporix
  - Full CRUD operations

[0.5.0]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.5.0

## [0.4.1] - 2026-01-27

### Improvements

- **Documentation: emporix_tenant_configuration**
  - Documentation improvements

[0.4.1]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.4.1

## [0.4.0] - 2026-01-22

### Added

- **New Resource: emporix_shipping_zone** - Manage shipping zones in Emporix
  - Full CRUD operations

[0.4.0]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.4.0

## [0.3.0] - 2026-01-08

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
- Documentation improvements.

[0.3.0]: https://github.com/emporix/terraform-provider-emporix/releases/tag/v0.3.0

## [0.2.0] - 2025-12-08

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

## [0.1.2] - 2025-12-05

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

## [0.1.1] - 2025-11-24

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
