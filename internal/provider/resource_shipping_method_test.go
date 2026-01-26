package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccShippingMethodResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckShippingMethodDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccShippingMethodResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "id", "standard-shipping"),
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "site", "main"),
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "zone_id", "zone-test"),
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "name.en", "Standard Shipping"),
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "active", "true"),
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "fees.#", "1"),
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "fees.0.min_order_value.amount", "0"),
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "fees.0.min_order_value.currency", "USD"),
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "fees.0.cost.amount", "5.99"),
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "fees.0.cost.currency", "USD"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "emporix_shipping_method.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "main:zone-test:standard-shipping",
			},
			// Update testing
			{
				Config: testAccShippingMethodResourceConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "name.en", "Standard Shipping Updated"),
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "active", "false"),
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "fees.0.cost.amount", "7.99"),
				),
			},
		},
	})
}

func TestAccShippingMethodResource_multipleFees(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckShippingMethodDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShippingMethodResourceConfig_multipleFees(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "id", "express-shipping"),
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "fees.#", "3"),
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "fees.0.min_order_value.amount", "0"),
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "fees.0.cost.amount", "15"),
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "fees.1.min_order_value.amount", "50"),
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "fees.1.cost.amount", "10"),
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "fees.2.min_order_value.amount", "100"),
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "fees.2.cost.amount", "0"),
				),
			},
		},
	})
}

func TestAccShippingMethodResource_withMaxOrderValue(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckShippingMethodDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShippingMethodResourceConfig_maxOrderValue(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "id", "budget-shipping"),
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "max_order_value.amount", "500"),
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "max_order_value.currency", "USD"),
					resource.TestCheckResourceAttr("emporix_shipping_method.test", "shipping_tax_code", "TAX-SHIPPING-01"),
				),
			},
		},
	})
}

func testAccCheckShippingMethodResourceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		client, err := getTestClient()
		if err != nil {
			return fmt.Errorf("failed to get test client: %w", err)
		}

		site := rs.Primary.Attributes["site"]
		zoneID := rs.Primary.Attributes["zone_id"]
		id := rs.Primary.ID

		_, err = client.GetShippingMethod(context.Background(), site, zoneID, id)
		if err != nil {
			return fmt.Errorf("Shipping method not found: %s", err)
		}

		return nil
	}
}

func testAccCheckShippingMethodDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := getTestClient()
	if err != nil {
		return fmt.Errorf("failed to get test client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "emporix_shipping_method" {
			continue
		}

		site := rs.Primary.Attributes["site"]
		zoneID := rs.Primary.Attributes["zone_id"]
		id := rs.Primary.ID

		_, err := client.GetShippingMethod(ctx, site, zoneID, id)

		if IsNotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Shipping method %s still exists", id)
	}

	return nil
}

func testAccShippingMethodResourceConfig_basic() string {
	return `
resource "emporix_shipping_zone" "test" {
  id   = "zone-test"
  site = "main"

  name = {
    en = "Test Zone"
  }

  ship_to = [
    { country = "US" }
  ]
}

resource "emporix_shipping_method" "test" {
  id      = "standard-shipping"
  site    = "main"
  zone_id = emporix_shipping_zone.test.id

  name = {
    en = "Standard Shipping"
  }

  active = true

  fees = [
    {
      min_order_value = {
        amount   = 0
        currency = "USD"
      }
      cost = {
        amount   = 5.99
        currency = "USD"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.test]
}
`
}

func testAccShippingMethodResourceConfig_updated() string {
	return `
resource "emporix_shipping_zone" "test" {
  id   = "zone-test"
  site = "main"

  name = {
    en = "Test Zone"
  }

  ship_to = [
    { country = "US" }
  ]
}

resource "emporix_shipping_method" "test" {
  id      = "standard-shipping"
  site    = "main"
  zone_id = emporix_shipping_zone.test.id

  name = {
    en = "Standard Shipping Updated"
  }

  active = false

  fees = [
    {
      min_order_value = {
        amount   = 0
        currency = "USD"
      }
      cost = {
        amount   = 7.99
        currency = "USD"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.test]
}
`
}

func testAccShippingMethodResourceConfig_multipleFees() string {
	return `
resource "emporix_shipping_zone" "test" {
  id   = "zone-test"
  site = "main"

  name = {
    en = "Test Zone"
  }

  ship_to = [
    { country = "US" }
  ]
}

resource "emporix_shipping_method" "test" {
  id      = "express-shipping"
  site    = "main"
  zone_id = emporix_shipping_zone.test.id

  name = {
    en = "Express Shipping"
    de = "Expressversand"
  }

  active = true

  fees = [
    {
      min_order_value = {
        amount   = 0
        currency = "USD"
      }
      cost = {
        amount   = 15
        currency = "USD"
      }
    },
    {
      min_order_value = {
        amount   = 50
        currency = "USD"
      }
      cost = {
        amount   = 10
        currency = "USD"
      }
    },
    {
      min_order_value = {
        amount   = 100
        currency = "USD"
      }
      cost = {
        amount   = 0
        currency = "USD"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.test]
}
`
}

func testAccShippingMethodResourceConfig_maxOrderValue() string {
	return `
resource "emporix_shipping_zone" "test" {
  id   = "zone-test"
  site = "main"

  name = {
    en = "Test Zone"
  }

  ship_to = [
    { country = "US" }
  ]
}

resource "emporix_shipping_method" "test" {
  id      = "budget-shipping"
  site    = "main"
  zone_id = emporix_shipping_zone.test.id

  name = {
    en = "Budget Shipping"
  }

  active = true

  max_order_value = {
    amount   = 500
    currency = "USD"
  }

  fees = [
    {
      min_order_value = {
        amount   = 0
        currency = "USD"
      }
      cost = {
        amount   = 3.99
        currency = "USD"
      }
    }
  ]

  shipping_tax_code = "TAX-SHIPPING-01"

  depends_on = [emporix_shipping_zone.test]
}
`
}
