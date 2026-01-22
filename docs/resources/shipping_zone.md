---
page_title: "emporix_shipping_zone Resource - terraform-provider-emporix"
subcategory: ""
description: |-
  Shipping zone resource for configuring delivery zones in Emporix.
---

# emporix_shipping_zone (Resource)

Manages shipping zones in Emporix. Shipping zones define geographical areas where shipments can be delivered, along with the countries and postal codes included in each zone.

## Example Usage

### Basic Shipping Zone

```terraform
resource "emporix_shipping_zone" "germany" {
  id   = "zone-germany"
  site = "main"
  name = {
    en = "Germany Zone"
  }

  ship_to = [
    {
      country = "DE"
    }
  ]
}
```

### Shipping Zone with Postal Code

```terraform
resource "emporix_shipping_zone" "berlin" {
  id   = "zone-berlin"
  site = "main"
  name = {
    en = "Berlin Area"
  }

  ship_to = [
    {
      country     = "DE"
      postal_code = "10*"
    }
  ]
}
```

### Default Shipping Zone

```terraform
resource "emporix_shipping_zone" "default" {
  id      = "zone-default"
  site    = "main"
  name    = {
    en = "Default Delivery Zone"
  }
  default = true

  ship_to = [
    {
      country = "DE"
    },
    {
      country = "AT"
    },
    {
      country = "CH"
    }
  ]
}
```

### Zone with Translations

```terraform
resource "emporix_shipping_zone" "europe" {
  id   = "zone-europe"
  site = "main"
  name = {
    en = "Europe Zone"
    de = "Europa Zone"
    fr = "Zone Europe"
  }

  ship_to = [
    {
      country = "DE"
    },
    {
      country = "FR"
    },
    {
      country = "IT"
    }
  ]
}
```

### German Language Zone

```terraform
resource "emporix_shipping_zone" "german_zone" {
  id   = "zone-de"
  site = "main"
  name = {
    de = "Deutsche Zone"
  }

  ship_to = [
    {
      country = "DE"
    }
  ]
}
```

## Schema

### Required

- `id` (String) Shipping zone's unique identifier. Changing this forces a new resource to be created.
- `site` (String) Site identifier. Typically 'main' for single-shop tenants. Changing this forces a new resource to be created.
- `name` (Map of String) Zone name as a map of language codes to translated names. Use a single entry for single-language zones (e.g., `{en = "Zone Name"}`) or multiple entries for multi-language zones (e.g., `{en = "English", de = "Deutsch", fr = "Français"}`). The map is sent to the API as-is.
- `ship_to` (List of Object) Collection of shipping destinations. At least one destination is required. **Important:** Each country can only appear once in the list.
  - `country` (String) Country code (e.g., 'DE', 'US', 'FR').
  - `postal_code` (String, Optional) Postal code or postal code pattern. Supports wildcards (e.g., '70*' for all codes starting with 70).

### Optional

- `default` (Boolean) Flag indicating whether the zone is the default delivery zone for the site. **Note:** The Emporix API automatically sets this to `true` for the first shipping zone created, regardless of the value specified in your configuration.

## Name Format

The `name` attribute is always a Map of language codes to translated names. The map is sent to the API exactly as specified.

### Single Language

For zones with a single language, use a map with one entry:

```terraform
name = {
  en = "My Zone Name"
}
```

**Supported language codes:** Any ISO language code (en, de, fr, es, it, pl, etc.)

**Examples:**
```terraform
# English
name = { en = "English Zone" }

# German
name = { de = "Deutsche Zone" }

# French
name = { fr = "Zone Française" }

# Polish
name = { pl = "Strefa Polska" }
```

### Multiple Languages

For multi-language zones, use a map with multiple entries:

```terraform
name = {
  en = "English Name"
  de = "Deutscher Name"
  fr = "Nom Français"
}
```

## Postal Code Patterns

Postal codes support pattern matching with wildcards:

- `70*` - Matches all postal codes starting with 70 (70000, 70190, 70199, etc.)
- `1010` - Exact match for postal code 1010
- Empty - Matches entire country

**Important Limitation:** Each country can only appear **once** in the `ship_to` list. The Emporix API does not support multiple entries for the same country with different postal codes.

**Not Allowed:**
```terraform
ship_to = [
  { country = "DE", postal_code = "10*" },  # Berlin
  { country = "DE", postal_code = "80*" }   # Munich - ERROR: DE appears twice!
]
```

**Allowed:**
```terraform
# Option 1: Single postal code pattern per country
ship_to = [
  { country = "DE", postal_code = "10*" }   # Only Berlin
]

# Option 2: Entire country (no postal code filter)
ship_to = [
  { country = "DE" }   # All of Germany
]

# Option 3: Multiple countries (each appears once)
ship_to = [
  { country = "DE", postal_code = "10*" },  # Germany - Berlin
  { country = "AT", postal_code = "10*" },  # Austria - Vienna
  { country = "FR", postal_code = "75*" }   # France - Paris
]
```

If you need to cover multiple postal code ranges within the same country, you must create separate shipping zones.

## Destination Ordering

**Important:** The `ship_to` list is automatically sorted alphabetically by country code (primary) and postal code (secondary). This matches the Emporix API's storage format and ensures consistent state.

**Example - Your configuration:**
```terraform
ship_to = [
  { country = "FR" },
  { country = "DE" },
  { country = "AT" }
]
```

**State will show (sorted):**
```terraform
ship_to = [
  { country = "AT" },
  { country = "DE" },
  { country = "FR" }
]
```

This is expected behavior. The order you specify in your configuration does not matter - destinations will always be stored and displayed in sorted order to match the API's storage format.

**Sorting Rules:**
1. Primary sort: Country code (alphabetical, e.g., AT < DE < FR)
2. Secondary sort: Postal code (alphabetical, e.g., 70* < 80*)

## Import

Shipping zones can be imported using their ID and site:

```shell
terraform import emporix_shipping_zone.example main:zone-germany
```

Note: The import ID format is `site:zone_id`.

## Required OAuth Scopes

To manage shipping zones, your API client must have the following scopes:

**Required Scopes:**
- `shipping.shipping_read` - Required for reading shipping zone information
- `shipping.shipping_manage` - Required for creating, updating, and deleting shipping zones

## Notes

- The `id` and `site` attributes cannot be changed after creation. Any change will force the resource to be recreated.
- Each zone must have at least one shipping destination in the `ship_to` list.
- If `default` is set to `true`, this zone becomes the fallback for addresses that don't match other zones.
- Postal code patterns are case-insensitive and support the `*` wildcard for prefix matching.
- The `name` field will be returned by the API exactly as stored - as a simple string or JSON map depending on input format.
