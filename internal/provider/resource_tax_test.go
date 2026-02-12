package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTaxResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTaxDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTaxResourceConfig_basic("PL"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_tax.test", "country_code", "PL"),
					resource.TestCheckResourceAttr("emporix_tax.test", "tax_classes.#", "2"),
					resource.TestCheckResourceAttr("emporix_tax.test", "tax_classes.0.code", "TEST_STANDARD"),
					resource.TestCheckResourceAttr("emporix_tax.test", "tax_classes.0.rate", "0.23"),
					resource.TestCheckResourceAttr("emporix_tax.test", "tax_classes.0.is_default", "true"),
					resource.TestCheckResourceAttr("emporix_tax.test", "tax_classes.1.code", "TEST_REDUCED"),
					resource.TestCheckResourceAttr("emporix_tax.test", "tax_classes.1.rate", "0.08"),
					resource.TestCheckResourceAttrSet("emporix_tax.test", "tax_classes.0.name.en"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "emporix_tax.test",
				ImportState:                          true,
				ImportStateId:                        "PL",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "country_code",
			},
		},
	})
}

func TestAccTaxResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTaxDestroy,
		Steps: []resource.TestStep{
			// Create with 2 tax classes
			{
				Config: testAccTaxResourceConfig_basic("CZ"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_tax.test", "country_code", "CZ"),
					resource.TestCheckResourceAttr("emporix_tax.test", "tax_classes.#", "2"),
					resource.TestCheckResourceAttr("emporix_tax.test", "tax_classes.0.rate", "0.23"),
				),
			},
			// Update to 3 tax classes and change rates
			{
				Config: testAccTaxResourceConfig_updated("CZ"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_tax.test", "country_code", "CZ"),
					resource.TestCheckResourceAttr("emporix_tax.test", "tax_classes.#", "3"),
					resource.TestCheckResourceAttr("emporix_tax.test", "tax_classes.0.rate", "0.21"),
					resource.TestCheckResourceAttr("emporix_tax.test", "tax_classes.1.rate", "0.15"),
					resource.TestCheckResourceAttr("emporix_tax.test", "tax_classes.2.code", "TEST_ZERO"),
				),
			},
		},
	})
}

func TestAccTaxResource_multipleDefaults(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccTaxResourceConfig_multipleDefaults("RO"),
				ExpectError: regexp.MustCompile("Multiple Default Tax Classes"),
			},
		},
	})
}

func TestAccTaxResource_requiresReplace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTaxDestroy,
		Steps: []resource.TestStep{
			// Create with HU
			{
				Config: testAccTaxResourceConfig_basic("HU"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_tax.test", "country_code", "HU"),
				),
			},
			// Change country_code to SK (should require replace)
			{
				Config: testAccTaxResourceConfig_basic("SK"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_tax.test", "country_code", "SK"),
				),
			},
		},
	})
}

func TestAccTaxResource_multipleCountries(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTaxDestroy,
		Steps: []resource.TestStep{
			// Create multiple tax configurations
			{
				Config: testAccTaxResourceConfig_multiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_tax.de", "country_code", "DE"),
					resource.TestCheckResourceAttr("emporix_tax.fr", "country_code", "FR"),
					resource.TestCheckResourceAttr("emporix_tax.it", "country_code", "IT"),
				),
			},
		},
	})
}

func TestAccTaxResource_withDescriptions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTaxDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaxResourceConfig_withDescriptions("GB"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_tax.test", "country_code", "GB"),
					resource.TestCheckResourceAttrSet("emporix_tax.test", "tax_classes.0.description.en"),
				),
			},
		},
	})
}

// testAccTaxResourceConfig_basic generates a basic tax configuration
func testAccTaxResourceConfig_basic(countryCode string) string {
	return fmt.Sprintf(`
resource "emporix_tax" "test" {
  country_code = %[1]q

  tax_classes = [
    {
      code = "TEST_STANDARD"
      name = {
        en = "Standard Rate"
      }
      rate       = 0.23
      is_default = true
      order      = 1
    },
    {
      code = "TEST_REDUCED"
      name = {
        en = "Reduced Rate"
      }
      rate       = 0.08
      is_default = false
      order      = 2
    }
  ]
}
`, countryCode)
}

// testAccTaxResourceConfig_updated generates an updated tax configuration with 3 classes
func testAccTaxResourceConfig_updated(countryCode string) string {
	return fmt.Sprintf(`
resource "emporix_tax" "test" {
  country_code = %[1]q

  tax_classes = [
    {
      code = "TEST_STANDARD"
      name = {
        en = "Standard Rate"
        de = "Standardsatz"
      }
      rate       = 0.21
      is_default = true
      order      = 1
    },
    {
      code = "TEST_REDUCED"
      name = {
        en = "Reduced Rate"
      }
      rate  = 0.15
      order = 2
    },
    {
      code = "TEST_ZERO"
      name = {
        en = "Zero Rate"
      }
      rate  = 0.0
      order = 3
    }
  ]
}
`, countryCode)
}

// testAccTaxResourceConfig_multipleDefaults generates a config with multiple defaults (should fail validation)
func testAccTaxResourceConfig_multipleDefaults(countryCode string) string {
	return fmt.Sprintf(`
resource "emporix_tax" "test" {
  country_code = %[1]q

  tax_classes = [
    {
      code       = "TEST_STANDARD"
      name       = { en = "Standard" }
      rate       = 0.19
      is_default = true
    },
    {
      code       = "TEST_REDUCED"
      name       = { en = "Reduced" }
      rate       = 0.07
      is_default = true
    }
  ]
}
`, countryCode)
}

// testAccTaxResourceConfig_multiple generates multiple tax configurations
func testAccTaxResourceConfig_multiple() string {
	return `
resource "emporix_tax" "de" {
  country_code = "DE"

  tax_classes = [
    {
      code       = "TEST_STANDARD"
      name       = { en = "Standard VAT" }
      rate       = 0.19
      is_default = true
      order      = 1
    },
    {
      code  = "TEST_REDUCED"
      name  = { en = "Reduced VAT" }
      rate  = 0.07
      order = 2
    }
  ]
}

resource "emporix_tax" "fr" {
  country_code = "FR"

  tax_classes = [
    {
      code       = "TEST_STANDARD"
      name       = { en = "Standard VAT" }
      rate       = 0.20
      is_default = true
      order      = 1
    },
    {
      code  = "TEST_REDUCED"
      name  = { en = "Reduced VAT" }
      rate  = 0.055
      order = 2
    }
  ]
}

resource "emporix_tax" "it" {
  country_code = "IT"

  tax_classes = [
    {
      code       = "TEST_STANDARD"
      name       = { en = "Standard VAT" }
      rate       = 0.22
      is_default = true
      order      = 1
    },
    {
      code  = "TEST_REDUCED"
      name  = { en = "Reduced VAT" }
      rate  = 0.10
      order = 2
    }
  ]
}
`
}

// testAccTaxResourceConfig_withDescriptions generates a config with descriptions
func testAccTaxResourceConfig_withDescriptions(countryCode string) string {
	return fmt.Sprintf(`
resource "emporix_tax" "test" {
  country_code = %[1]q

  tax_classes = [
    {
      code = "TEST_STANDARD"
      name = {
        en = "Standard VAT"
      }
      rate       = 0.20
      is_default = true
      order      = 1
      description = {
        en = "Standard 20%% VAT rate for most goods and services"
      }
    },
    {
      code = "TEST_REDUCED"
      name = {
        en = "Reduced VAT"
      }
      rate  = 0.05
      order = 2
      description = {
        en = "Reduced 5%% VAT rate for specific goods"
      }
    }
  ]
}
`, countryCode)
}

// testAccCheckTaxDestroy verifies that taxes have been deleted
func testAccCheckTaxDestroy(s *terraform.State) error {
	ctx := context.Background()

	// Get configured client
	client, err := getTestClient()
	if err != nil {
		return fmt.Errorf("failed to get test client: %w", err)
	}

	// Iterate through all resources in state
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "emporix_tax" {
			continue
		}

		countryCode := rs.Primary.Attributes["country_code"]

		// Retry checking if the resource is deleted with exponential backoff
		// The Tax API may have eventual consistency
		maxRetries := 10
		for i := 0; i < maxRetries; i++ {
			// Try to get the tax
			_, err := client.GetTax(ctx, countryCode)

			// If not found, resource was successfully destroyed
			if IsNotFound(err) {
				break
			}

			// If other error, fail the test
			if err != nil {
				return fmt.Errorf("unexpected error checking tax: %w", err)
			}

			// If this is the last retry and tax still exists, fail
			if i == maxRetries-1 {
				return fmt.Errorf("tax %s still exists after destroy (tried %d times)", countryCode, maxRetries)
			}

			// Wait before retrying (exponential backoff: 100ms, 200ms, 400ms, ...)
			time.Sleep(time.Duration(100*(1<<uint(i))) * time.Millisecond)
		}
	}

	return nil
}
