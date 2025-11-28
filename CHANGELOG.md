# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2024-11-24

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

[0.1.0]: https://github.com/yourusername/terraform-provider-emporix/releases/tag/v0.1.0
