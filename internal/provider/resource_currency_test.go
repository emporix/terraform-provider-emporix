package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccCurrencyResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCurrencyDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCurrencyResourceConfig("PLN"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_currency.test", "code", "PLN"),
					resource.TestCheckResourceAttrSet("emporix_currency.test", "name.en"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "emporix_currency.test",
				ImportState:                          true,
				ImportStateId:                        "PLN",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "code",
			},
		},
	})
}

func TestAccCurrencyResource_requiresReplace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCurrencyDestroy,
		Steps: []resource.TestStep{
			// Create with SEK
			{
				Config: testAccCurrencyResourceConfig("SEK"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_currency.test", "code", "SEK"),
				),
			},
			// Change code to NOK (should require replace)
			{
				Config: testAccCurrencyResourceConfig("NOK"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_currency.test", "code", "NOK"),
				),
			},
		},
	})
}

func TestAccCurrencyResource_multipleCurrencies(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCurrencyDestroy,
		Steps: []resource.TestStep{
			// Create multiple currencies
			{
				Config: testAccCurrencyResourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_currency.czk", "code", "CZK"),
					resource.TestCheckResourceAttr("emporix_currency.huf", "code", "HUF"),
					resource.TestCheckResourceAttr("emporix_currency.ron", "code", "RON"),
				),
			},
		},
	})
}

func TestAccCurrencyResource_readOnlyFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCurrencyDestroy,
		Steps: []resource.TestStep{
			// Verify name is properly stored
			{
				Config: testAccCurrencyResourceConfig("DKK"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_currency.test", "code", "DKK"),
					// Name should be set from configuration
					resource.TestCheckResourceAttr("emporix_currency.test", "name.en", "DKK"),
				),
			},
		},
	})
}

// testAccCurrencyResourceConfig generates a currency configuration
func testAccCurrencyResourceConfig(code string) string {
	return fmt.Sprintf(`
resource "emporix_currency" "test" {
  code = %[1]q
  name = {
    en = %[1]q
  }
}
`, code)
}

// testAccCurrencyResourceConfigMultiple generates a configuration with multiple currencies
func testAccCurrencyResourceConfigMultiple() string {
	return `
resource "emporix_currency" "czk" {
  code = "CZK"
  name = {
    en = "Czech Koruna"
  }
}

resource "emporix_currency" "huf" {
  code = "HUF"
  name = {
    en = "Hungarian Forint"
  }
}

resource "emporix_currency" "ron" {
  code = "RON"
  name = {
    en = "Romanian Leu"
  }
}
`
}

// testAccCheckCurrencyDestroy verifies that currencies have been deleted
func testAccCheckCurrencyDestroy(s *terraform.State) error {
	ctx := context.Background()

	// Get configured client
	client, err := getTestClient()
	if err != nil {
		return fmt.Errorf("failed to get test client: %w", err)
	}

	// Iterate through all resources in state
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "emporix_currency" {
			continue
		}

		code := rs.Primary.Attributes["code"]

		// Try to get the currency
		currency, err := client.GetCurrency(ctx, code)

		// If we get a 404 or the currency is nil, that's what we want
		if err != nil || currency == nil {
			continue
		}

		// If currency still exists, that's an error
		return fmt.Errorf("currency %s still exists after destroy", code)
	}

	return nil
}
