package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSiteSettingsResource_basic(t *testing.T) {
	code := fmt.Sprintf("test-site-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSiteSettingsDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSiteSettingsResourceConfigBasic(code),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "code", code),
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "name", "Test Site"),
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "active", "true"),
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "default_language", "en"),
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "currency", "USD"),
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "languages.#", "1"),
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "languages.0", "en"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "emporix_sitesettings.test",
				ImportState:                          true,
				ImportStateId:                        code,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "code",
			},
			// Update and Read testing
			{
				Config: testAccSiteSettingsResourceConfigUpdated(code),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "code", code),
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "name", "Updated Test Site"),
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "active", "false"),
				),
			},
		},
	})
}

func TestAccSiteSettingsResource_multipleLanguages(t *testing.T) {
	code := fmt.Sprintf("test-lang-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSiteSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteSettingsResourceConfigMultiLanguage(code),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "code", code),
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "default_language", "en"),
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "languages.#", "3"),
				),
			},
		},
	})
}

func TestAccSiteSettingsResource_multipleCurrencies(t *testing.T) {
	code := fmt.Sprintf("test-curr-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSiteSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteSettingsResourceConfigMultiCurrency(code),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "code", code),
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "currency", "USD"),
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "available_currencies.#", "3"),
				),
			},
		},
	})
}

func TestAccSiteSettingsResource_shipToCountries(t *testing.T) {
	code := fmt.Sprintf("test-ship-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSiteSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteSettingsResourceConfigWithCountries(code),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "code", code),
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "ship_to_countries.#", "3"),
				),
			},
			// Update countries
			{
				Config: testAccSiteSettingsResourceConfigWithCountriesUpdated(code),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "code", code),
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "ship_to_countries.#", "2"),
				),
			},
		},
	})
}

func TestAccSiteSettingsResource_includesTax(t *testing.T) {
	code := fmt.Sprintf("test-tax-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSiteSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteSettingsResourceConfigWithTax(code, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "code", code),
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "includes_tax", "true"),
				),
			},
			// Update tax setting
			{
				Config: testAccSiteSettingsResourceConfigWithTax(code, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "code", code),
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "includes_tax", "false"),
				),
			},
		},
	})
}

func TestAccSiteSettingsResource_requiresReplace(t *testing.T) {
	code1 := fmt.Sprintf("test-replace1-%d", time.Now().Unix())
	code2 := fmt.Sprintf("test-replace2-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSiteSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteSettingsResourceConfigBasic(code1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "code", code1),
				),
			},
			// Change code (should require replace)
			{
				Config: testAccSiteSettingsResourceConfigBasic(code2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "code", code2),
				),
			},
		},
	})
}

// testAccSiteSettingsResourceConfigBasic generates a basic site settings configuration
func testAccSiteSettingsResourceConfigBasic(code string) string {
	return fmt.Sprintf(`
resource "emporix_sitesettings" "test" {
  code             = %[1]q
  name             = "Test Site"
  active           = true
  default_language = "en"
  languages        = ["en"]
  currency         = "USD"
  ship_to_countries = ["US"]

  home_base = {
    address = {
      zip_code = "10001"
      city     = "New York"
      country  = "US"
    }
  }
}
`, code)
}

// testAccSiteSettingsResourceConfigUpdated generates an updated site settings configuration
func testAccSiteSettingsResourceConfigUpdated(code string) string {
	return fmt.Sprintf(`
resource "emporix_sitesettings" "test" {
  code             = %[1]q
  name             = "Updated Test Site"
  active           = false
  default_language = "en"
  languages        = ["en"]
  currency         = "USD"
  ship_to_countries = ["US"]

  home_base = {
    address = {
      zip_code = "10001"
      city     = "New York"
      country  = "US"
    }
  }
}
`, code)
}

// testAccSiteSettingsResourceConfigMultiLanguage generates a site with multiple languages
func testAccSiteSettingsResourceConfigMultiLanguage(code string) string {
	return fmt.Sprintf(`
resource "emporix_sitesettings" "test" {
  code             = %[1]q
  name             = "Multi-Language Site"
  active           = true
  default_language = "en"
  languages        = ["en", "de", "fr"]
  currency         = "USD"
  ship_to_countries = ["US"]

  home_base = {
    address = {
      zip_code = "10001"
      city     = "New York"
      country  = "US"
    }
  }
}
`, code)
}

// testAccSiteSettingsResourceConfigMultiCurrency generates a site with multiple currencies
func testAccSiteSettingsResourceConfigMultiCurrency(code string) string {
	return fmt.Sprintf(`
resource "emporix_sitesettings" "test" {
  code                 = %[1]q
  name                 = "Multi-Currency Site"
  active               = true
  default_language     = "en"
  languages            = ["en"]
  currency             = "USD"
  available_currencies = ["USD", "EUR", "GBP"]
  ship_to_countries    = ["US"]

  home_base = {
    address = {
      zip_code = "10001"
      city     = "New York"
      country  = "US"
    }
  }
}
`, code)
}

// testAccSiteSettingsResourceConfigWithCountries generates a site with ship-to countries
func testAccSiteSettingsResourceConfigWithCountries(code string) string {
	return fmt.Sprintf(`
resource "emporix_sitesettings" "test" {
  code              = %[1]q
  name              = "Site With Countries"
  active            = true
  default_language  = "en"
  languages         = ["en"]
  currency          = "USD"
  ship_to_countries = ["US", "CA", "MX"]

  home_base = {
    address = {
      zip_code = "10001"
      city     = "New York"
      country  = "US"
    }
  }
}
`, code)
}

// testAccSiteSettingsResourceConfigWithCountriesUpdated generates a site with updated countries
func testAccSiteSettingsResourceConfigWithCountriesUpdated(code string) string {
	return fmt.Sprintf(`
resource "emporix_sitesettings" "test" {
  code              = %[1]q
  name              = "Site With Countries"
  active            = true
  default_language  = "en"
  languages         = ["en"]
  currency          = "USD"
  ship_to_countries = ["US", "CA"]

  home_base = {
    address = {
      zip_code = "10001"
      city     = "New York"
      country  = "US"
    }
  }
}
`, code)
}

// testAccSiteSettingsResourceConfigWithTax generates a site with tax settings
func testAccSiteSettingsResourceConfigWithTax(code string, includesTax bool) string {
	return fmt.Sprintf(`
resource "emporix_sitesettings" "test" {
  code             = %[1]q
  name             = "Site With Tax"
  active           = true
  default_language = "en"
  languages        = ["en"]
  currency         = "USD"
  includes_tax     = %[2]t
  ship_to_countries = ["US"]

  home_base = {
    address = {
      zip_code = "10001"
      city     = "New York"
      country  = "US"
    }
  }
}
`, code, includesTax)
}

func TestAccSiteSettingsResource_mixins(t *testing.T) {
	code := fmt.Sprintf("test-mixin-%d", time.Now().Unix())
	schemaId := fmt.Sprintf("ts-mxn-sch-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSiteSettingsDestroy,
		Steps: []resource.TestStep{
			// Create site with a mixin that references a schema
			{
				Config: testAccSiteSettingsResourceConfigWithMixin(code, schemaId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "code", code),
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "name", "Site With Mixin"),
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "mixins.#", "1"),
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "mixins.0.name", schemaId),
					resource.TestCheckResourceAttrSet("emporix_sitesettings.test", "mixins.0.schema_url"),
				),
			},
			// Update mixin fields
			{
				Config: testAccSiteSettingsResourceConfigWithMixinUpdated(code, schemaId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "code", code),
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "mixins.#", "1"),
					resource.TestCheckResourceAttr("emporix_sitesettings.test", "mixins.0.name", schemaId),
				),
			},
		},
	})
}

func testAccSiteSettingsResourceConfigWithMixin(code string, schemaId string) string {
	return fmt.Sprintf(`
resource "emporix_schema" "site_mixin" {
  id = %[2]q
  name = {
    en = "Site Mixin Schema"
  }
  types = ["SITE"]

  attributes = [
    {
      key = "displayName"
      name = {
        en = "Display Name"
      }
      type = "TEXT"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
    },
    {
      key = "maxItems"
      name = {
        en = "Max Items"
      }
      type = "NUMBER"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
    },
    {
      key = "isEnabled"
      name = {
        en = "Is Enabled"
      }
      type = "BOOLEAN"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
    }
  ]
}

resource "emporix_sitesettings" "test" {
  code             = %[1]q
  name             = "Site With Mixin"
  active           = true
  default_language = "en"
  languages        = ["en"]
  currency         = "USD"
  ship_to_countries = ["US"]

  home_base = {
    address = {
      zip_code = "10001"
      city     = "New York"
      country  = "US"
    }
  }

  mixins = [
    {
      name       = emporix_schema.site_mixin.id
      schema_url = emporix_schema.site_mixin.schema_url
      fields     = jsonencode({
        displayName = "Test Store"
        maxItems    = 100
        isEnabled   = true
      })
    }
  ]
}
`, code, schemaId)
}

func testAccSiteSettingsResourceConfigWithMixinUpdated(code string, schemaId string) string {
	return fmt.Sprintf(`
resource "emporix_schema" "site_mixin" {
  id = %[2]q
  name = {
    en = "Site Mixin Schema"
  }
  types = ["SITE"]

  attributes = [
    {
      key = "displayName"
      name = {
        en = "Display Name"
      }
      type = "TEXT"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
    },
    {
      key = "maxItems"
      name = {
        en = "Max Items"
      }
      type = "NUMBER"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
    },
    {
      key = "isEnabled"
      name = {
        en = "Is Enabled"
      }
      type = "BOOLEAN"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
    }
  ]
}

resource "emporix_sitesettings" "test" {
  code             = %[1]q
  name             = "Site With Mixin"
  active           = true
  default_language = "en"
  languages        = ["en"]
  currency         = "USD"
  ship_to_countries = ["US"]

  home_base = {
    address = {
      zip_code = "10001"
      city     = "New York"
      country  = "US"
    }
  }

  mixins = [
    {
      name       = emporix_schema.site_mixin.id
      schema_url = emporix_schema.site_mixin.schema_url
      fields     = jsonencode({
        displayName = "Updated Store Name"
        maxItems    = 250
        isEnabled   = false
      })
    }
  ]
}
`, code, schemaId)
}

// TestAccSiteSettingsResource_multiMixinOrdering covers what TestAccSiteSettingsResource_mixins
// cannot: it uses one mixin, and a one-element list has no order to get wrong.
//
// Three mixins rather than two (at two, a lucky order passes half the time) and two apply steps, for
// a second independent draw.
func TestAccSiteSettingsResource_multiMixinOrdering(t *testing.T) {
	stamp := time.Now().Unix()
	code := fmt.Sprintf("test-mixin-order-%d", stamp)

	// Declared alphabetically in the config, which is the order a correct provider hands back. The
	// -a-/-b-/-c- infixes keep the ids themselves in that order.
	schemaIDs := [3]string{
		fmt.Sprintf("ts-mxn-a-%d", stamp),
		fmt.Sprintf("ts-mxn-b-%d", stamp),
		fmt.Sprintf("ts-mxn-c-%d", stamp),
	}

	// The list is diffed by index, so asserting name-per-index IS asserting the order.
	orderChecks := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("emporix_sitesettings.test", "code", code),
		resource.TestCheckResourceAttr("emporix_sitesettings.test", "mixins.#", "3"),
		resource.TestCheckResourceAttr("emporix_sitesettings.test", "mixins.0.name", schemaIDs[0]),
		resource.TestCheckResourceAttr("emporix_sitesettings.test", "mixins.1.name", schemaIDs[1]),
		resource.TestCheckResourceAttr("emporix_sitesettings.test", "mixins.2.name", schemaIDs[2]),
		resource.TestCheckResourceAttrSet("emporix_sitesettings.test", "mixins.0.schema_url"),
		resource.TestCheckResourceAttrSet("emporix_sitesettings.test", "mixins.1.schema_url"),
		resource.TestCheckResourceAttrSet("emporix_sitesettings.test", "mixins.2.schema_url"),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSiteSettingsDestroy,
		Steps: []resource.TestStep{
			// Create the site with three mixins.
			{
				Config: testAccSiteSettingsResourceConfigWithThreeMixins(code, schemaIDs, "Test Store", 100, true),
				Check:  orderChecks,
			},
			// Change a field value in every mixin: an update, and a second draw.
			{
				Config: testAccSiteSettingsResourceConfigWithThreeMixins(code, schemaIDs, "Updated Store", 250, false),
				Check:  orderChecks,
			},
		},
	})
}

// TestAccSiteSettingsResource_multiMixinOrderingUnsortedConfig declares the same three mixins in
// reverse order: state must follow the config, not the alphabet. Sorting alone would make this a
// permanent diff.
//
// The import step covers the path with no prior order to follow, where the list arrives sorted.
func TestAccSiteSettingsResource_multiMixinOrderingUnsortedConfig(t *testing.T) {
	stamp := time.Now().Unix()
	code := fmt.Sprintf("test-mixin-unsorted-%d", stamp)

	schemaIDs := [3]string{
		fmt.Sprintf("ts-mxu-a-%d", stamp),
		fmt.Sprintf("ts-mxu-b-%d", stamp),
		fmt.Sprintf("ts-mxu-c-%d", stamp),
	}
	reversed := [3]string{schemaIDs[2], schemaIDs[1], schemaIDs[0]}

	// State must follow the config, not the alphabet.
	orderChecks := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("emporix_sitesettings.test", "mixins.#", "3"),
		resource.TestCheckResourceAttr("emporix_sitesettings.test", "mixins.0.name", reversed[0]),
		resource.TestCheckResourceAttr("emporix_sitesettings.test", "mixins.1.name", reversed[1]),
		resource.TestCheckResourceAttr("emporix_sitesettings.test", "mixins.2.name", reversed[2]),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSiteSettingsDestroy,
		Steps: []resource.TestStep{
			// Create with the list declared c, b, a.
			{
				Config: testAccSiteSettingsResourceConfigWithThreeMixins(code, reversed, "Test Store", 100, true),
				Check:  orderChecks,
			},
			// A value change keeps the declared order.
			{
				Config: testAccSiteSettingsResourceConfigWithThreeMixins(code, reversed, "Updated Store", 250, false),
				Check:  orderChecks,
			},
			// Import: no prior order exists, so the list lands sorted. Deterministic, not random.
			{
				ResourceName:      "emporix_sitesettings.test",
				ImportState:       true,
				ImportStateId:     code,
				ImportStateVerify: true,
				// The resource has no `id`; a site is identified by `code`.
				ImportStateVerifyIdentifierAttribute: "code",
				ImportStateVerifyIgnore:              []string{"mixins"},
			},
		},
	})
}

// testAccSiteSettingsThreeMixinSchemas declares one SITE schema per mixin: mixins are keyed by
// schema name, so three mixins need three schemas.
func testAccSiteSettingsThreeMixinSchemas(schemaIDs [3]string) string {
	var b strings.Builder

	for i, id := range schemaIDs {
		b.WriteString(fmt.Sprintf(`
resource "emporix_schema" "site_mixin_%[1]d" {
  id = %[2]q
  name = {
    en = "Site Mixin Schema %[1]d"
  }
  types = ["SITE"]

  attributes = [
    {
      key = "displayName"
      name = {
        en = "Display Name"
      }
      type = "TEXT"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
    },
    {
      key = "maxItems"
      name = {
        en = "Max Items"
      }
      type = "NUMBER"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
    },
    {
      key = "isEnabled"
      name = {
        en = "Is Enabled"
      }
      type = "BOOLEAN"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
    }
  ]
}
`, i, id))
	}

	return b.String()
}

// testAccSiteSettingsResourceConfigWithThreeMixins builds a site with one mixin per schema, in the
// order given. Every `fields` is non-empty because buildPatchData skips a mixin whose fields are
// empty, which would mask the ordering under test.
func testAccSiteSettingsResourceConfigWithThreeMixins(code string, schemaIDs [3]string, displayName string, maxItems int, isEnabled bool) string {
	var mixins strings.Builder

	for i := range schemaIDs {
		mixins.WriteString(fmt.Sprintf(`
    {
      name       = emporix_schema.site_mixin_%[1]d.id
      schema_url = emporix_schema.site_mixin_%[1]d.schema_url
      fields = jsonencode({
        displayName = %[2]q
        maxItems    = %[3]d
        isEnabled   = %[4]t
      })
    },`, i, fmt.Sprintf("%s %d", displayName, i), maxItems+i, isEnabled))
	}

	return testAccSiteSettingsThreeMixinSchemas(schemaIDs) + fmt.Sprintf(`
resource "emporix_sitesettings" "test" {
  code             = %[1]q
  name             = "Site With Three Mixins"
  active           = true
  default_language = "en"
  languages        = ["en"]
  currency         = "USD"
  ship_to_countries = ["US"]

  home_base = {
    address = {
      zip_code = "10001"
      city     = "New York"
      country  = "US"
    }
  }

  mixins = [%[2]s
  ]
}
`, code, mixins.String())
}

// testAccCheckSiteSettingsDestroy verifies that site settings have been deleted
func testAccCheckSiteSettingsDestroy(s *terraform.State) error {
	ctx := context.Background()

	// Get configured client
	client, err := getTestClient()
	if err != nil {
		return fmt.Errorf("failed to get test client: %w", err)
	}

	// Iterate through all resources in state
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "emporix_sitesettings" {
			continue
		}

		code := rs.Primary.Attributes["code"]

		// Try to get the site
		_, err := client.GetSite(ctx, code)

		// If not found, resource was successfully destroyed
		if IsNotFound(err) {
			continue
		}

		// If other error, fail the test
		if err != nil {
			return fmt.Errorf("unexpected error checking site settings: %w", err)
		}

		// If no error, site still exists
		return fmt.Errorf("site settings %s still exists after destroy", code)
	}

	return nil
}
