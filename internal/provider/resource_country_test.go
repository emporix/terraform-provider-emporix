package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCountryResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCountryResourceConfig("US", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_country.test", "code", "US"),
					resource.TestCheckResourceAttr("emporix_country.test", "active", "true"),
					resource.TestCheckResourceAttrSet("emporix_country.test", "name.en"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "emporix_country.test",
				ImportState:                          true,
				ImportStateId:                        "US",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "code",
			},
			// Update and Read testing
			{
				Config: testAccCountryResourceConfig("US", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_country.test", "code", "US"),
					resource.TestCheckResourceAttr("emporix_country.test", "active", "false"),
				),
			},
			// Update back to active
			{
				Config: testAccCountryResourceConfig("US", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_country.test", "code", "US"),
					resource.TestCheckResourceAttr("emporix_country.test", "active", "true"),
				),
			},
		},
	})
}

func TestAccCountryResource_defaultActive(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create without specifying active (should default to true)
			{
				Config: testAccCountryResourceConfigNoActive("CA"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_country.test", "code", "CA"),
					resource.TestCheckResourceAttr("emporix_country.test", "active", "true"),
				),
			},
		},
	})
}

func TestAccCountryResource_requiresReplace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with GB
			{
				Config: testAccCountryResourceConfig("GB", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_country.test", "code", "GB"),
				),
			},
			// Change code to DE (should require replace)
			{
				Config: testAccCountryResourceConfig("DE", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_country.test", "code", "DE"),
				),
			},
		},
	})
}

func TestAccCountryResource_multipleCountries(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create multiple countries
			{
				Config: testAccCountryResourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_country.us", "code", "US"),
					resource.TestCheckResourceAttr("emporix_country.us", "active", "true"),
					resource.TestCheckResourceAttr("emporix_country.ca", "code", "CA"),
					resource.TestCheckResourceAttr("emporix_country.ca", "active", "true"),
					resource.TestCheckResourceAttr("emporix_country.gb", "code", "GB"),
					resource.TestCheckResourceAttr("emporix_country.gb", "active", "false"),
				),
			},
		},
	})
}

func TestAccCountryResource_readOnlyFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCountryResourceConfig("FR", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_country.test", "code", "FR"),
					// Check that read-only fields are populated
					resource.TestCheckResourceAttrSet("emporix_country.test", "name.%"),
					resource.TestCheckResourceAttrSet("emporix_country.test", "regions.#"),
				),
			},
		},
	})
}

// testAccCountryResourceConfig generates a basic country resource configuration
func testAccCountryResourceConfig(code string, active bool) string {
	return fmt.Sprintf(`
resource "emporix_country" "test" {
  code   = %[1]q
  active = %[2]t
}
`, code, active)
}

// testAccCountryResourceConfigNoActive generates a country resource without active field
func testAccCountryResourceConfigNoActive(code string) string {
	return fmt.Sprintf(`
resource "emporix_country" "test" {
  code = %[1]q
}
`, code)
}

// testAccCountryResourceConfigMultiple generates a configuration with multiple countries
func testAccCountryResourceConfigMultiple() string {
	return `
resource "emporix_country" "us" {
  code   = "US"
  active = true
}

resource "emporix_country" "ca" {
  code   = "CA"
  active = true
}

resource "emporix_country" "gb" {
  code   = "GB"
  active = false
}
`
}
