package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccShippingZoneResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckShippingZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShippingZoneResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "id", "test-zone-1"),
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "site", "main"),
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "name.en", "Test Zone 1"),
					// First zone becomes default automatically - don't check default value
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "ship_to.#", "1"),
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "ship_to.0.country", "DE"),
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "ship_to.0.postal_code", "70190"),
				),
			},
		},
	})
}

func TestAccShippingZoneResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckShippingZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShippingZoneResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "name.en", "Test Zone 1"),
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "ship_to.#", "1"),
				),
			},
			{
				Config: testAccShippingZoneResourceConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "name.en", "Updated Zone 1"),
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "ship_to.#", "2"),
					// Destinations are sorted: AT, then DE
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "ship_to.0.country", "AT"),
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "ship_to.0.postal_code", "1010"),
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "ship_to.1.country", "DE"),
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "ship_to.1.postal_code", "70190"),
				),
			},
		},
	})
}

func TestAccShippingZoneResource_default(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckShippingZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShippingZoneResourceConfig_default(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "id", "test-zone-default"),
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "default", "true"),
				),
			},
		},
	})
}

func TestAccShippingZoneResource_translations(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckShippingZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShippingZoneResourceConfig_translations(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "id", "test-zone-i18n"),
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "name.en", "English Zone"),
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "name.de", "Deutsche Zone"),
				),
			},
		},
	})
}

func TestAccShippingZoneResource_multipleDestinations(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckShippingZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShippingZoneResourceConfig_multipleDestinations(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "ship_to.#", "3"),
					// Destinations are sorted: AT, DE, FR
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "ship_to.0.country", "AT"),
					// AT has no postal_code, so it should be null (not checked)
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "ship_to.1.country", "DE"),
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "ship_to.1.postal_code", "70*"),
					resource.TestCheckResourceAttr("emporix_shipping_zone.test", "ship_to.2.country", "FR"),
					// FR has no postal_code, so it should be null (not checked)
				),
			},
		},
	})
}

func testAccCheckShippingZoneDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := getTestClient()
	if err != nil {
		return fmt.Errorf("failed to get test client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "emporix_shipping_zone" {
			continue
		}

		zoneID := rs.Primary.ID
		site := rs.Primary.Attributes["site"]

		_, err := client.GetShippingZone(ctx, site, zoneID)

		if IsNotFound(err) {
			continue
		}

		if err != nil {
			return fmt.Errorf("unexpected error checking shipping zone: %w", err)
		}

		return fmt.Errorf("shipping zone %s still exists after destroy", zoneID)
	}

	return nil
}

func testAccShippingZoneResourceConfig_basic() string {
	return `
resource "emporix_shipping_zone" "test" {
  id   = "test-zone-1"
  site = "main"
  name = {
    en = "Test Zone 1"
  }

  ship_to = [
    {
      country     = "DE"
      postal_code = "70190"
    }
  ]
}
`
}

func testAccShippingZoneResourceConfig_updated() string {
	return `
resource "emporix_shipping_zone" "test" {
  id   = "test-zone-1"
  site = "main"
  name = {
    en = "Updated Zone 1"
  }

  ship_to = [
    {
      country     = "AT"
      postal_code = "1010"
    },
    {
      country     = "DE"
      postal_code = "70190"
    }
  ]
}
`
}

func testAccShippingZoneResourceConfig_default() string {
	return `
resource "emporix_shipping_zone" "test" {
  id   = "test-zone-default"
  site = "main"
  name = {
    en = "Default Zone"
  }
  default = true

  ship_to = [
    {
      country = "DE"
    }
  ]
}
`
}

func testAccShippingZoneResourceConfig_translations() string {
	return `
resource "emporix_shipping_zone" "test" {
  id   = "test-zone-i18n"
  site = "main"
  name = {
    en = "English Zone"
    de = "Deutsche Zone"
  }

  ship_to = [
    {
      country = "DE"
    }
  ]
}
`
}

func testAccShippingZoneResourceConfig_multipleDestinations() string {
	return `
resource "emporix_shipping_zone" "test" {
  id   = "test-zone-multi"
  site = "main"
  name = {
    en = "Multi Destination Zone"
  }

  ship_to = [
    {
      country = "AT"
    },
    {
      country     = "DE"
      postal_code = "70*"
    },
    {
      country = "FR"
    }
  ]
}
`
}
