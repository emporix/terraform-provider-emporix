package provider

import (
	"context"
	"fmt"
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
