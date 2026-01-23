package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDeliveryTimeResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryTimeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryTimeResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("emporix_delivery_time.test", "id"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "name", "friday-slots"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "site_code", "main"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "is_delivery_day", "true"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "zone_id", "zone1"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "time_zone_id", "Europe/Warsaw"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "day.weekday", "FRIDAY"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "slots.#", "1"),
				),
			},
		},
	})
}

func TestAccDeliveryTimeResource_multipleSlots(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryTimeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryTimeResourceConfig_multipleSlots(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("emporix_delivery_time.test", "id"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "name", "saturday-slots"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "slots.#", "2"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "slots.0.shipping_method", "standard"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "slots.0.capacity", "50"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "slots.1.shipping_method", "express"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "slots.1.capacity", "30"),
				),
			},
		},
	})
}

func TestAccDeliveryTimeResource_allZones(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryTimeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryTimeResourceConfig_allZones(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("emporix_delivery_time.test", "id"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "name", "all-zones-delivery"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "is_for_all_zones", "true"),
				),
			},
		},
	})
}

func TestAccDeliveryTimeResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryTimeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryTimeResourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "slots.#", "1"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "slots.0.capacity", "100"),
				),
			},
			{
				Config: testAccDeliveryTimeResourceConfig_updated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "slots.#", "1"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "slots.0.capacity", "150"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "slots.0.delivery_time_range.time_from", "09:00:00"),
				),
			},
		},
	})
}

func testAccCheckDeliveryTimeDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := getTestClient()
	if err != nil {
		return fmt.Errorf("failed to get test client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "emporix_delivery_time" {
			continue
		}

		id := rs.Primary.ID

		_, err := client.GetDeliveryTime(ctx, id)

		if IsNotFound(err) {
			continue
		}

		if err != nil {
			return fmt.Errorf("unexpected error checking delivery time: %w", err)
		}

		return fmt.Errorf("delivery time %s still exists after destroy", id)
	}

	return nil
}

func testAccDeliveryTimeResourceConfig_basic() string {
	return `
resource "emporix_delivery_time" "test" {
  name               = "friday-slots"
  site_code          = "main"
  is_delivery_day    = true
  zone_id            = "zone1"
  time_zone_id       = "Europe/Warsaw"
  is_for_all_zones   = false
  delivery_day_shift = 0

  day = {
    weekday = "FRIDAY"
  }

  slots = [
    {
      shipping_method = "standard"
      capacity        = 100

      delivery_time_range = {
        time_from = "10:00:00"
        time_to   = "12:00:00"
      }

      cut_off_time = {
        time                = "2023-06-12T06:00:00.000Z"
        delivery_cycle_name = "morning"
      }
    }
  ]
}
`
}

func testAccDeliveryTimeResourceConfig_updated() string {
	return `
resource "emporix_delivery_time" "test" {
  name               = "friday-slots"
  site_code          = "main"
  is_delivery_day    = true
  zone_id            = "zone1"
  time_zone_id       = "Europe/Warsaw"
  is_for_all_zones   = false
  delivery_day_shift = 0

  day = {
    weekday = "FRIDAY"
  }

  slots = [
    {
      shipping_method = "standard"
      capacity        = 150

      delivery_time_range = {
        time_from = "09:00:00"
        time_to   = "12:00:00"
      }

      cut_off_time = {
        time                = "2023-06-12T06:00:00.000Z"
        delivery_cycle_name = "morning"
      }
    }
  ]
}
`
}

func testAccDeliveryTimeResourceConfig_multipleSlots() string {
	return `
resource "emporix_delivery_time" "test" {
  name               = "saturday-slots"
  site_code          = "main"
  is_delivery_day    = true
  zone_id            = "zone1"
  time_zone_id       = "Europe/Warsaw"
  delivery_day_shift = 0

  day = {
    weekday = "SATURDAY"
  }

  slots = [
    {
      shipping_method = "standard"
      capacity        = 50

      delivery_time_range = {
        time_from = "10:00:00"
        time_to   = "12:00:00"
      }

      cut_off_time = {
        time                = "2023-06-13T08:00:00.000Z"
        delivery_cycle_name = "morning"
      }
    },
    {
      shipping_method = "express"
      capacity        = 30

      delivery_time_range = {
        time_from = "14:00:00"
        time_to   = "16:00:00"
      }

      cut_off_time = {
        time                = "2023-06-13T12:00:00.000Z"
        delivery_cycle_name = "afternoon"
      }
    }
  ]
}
`
}

func testAccDeliveryTimeResourceConfig_allZones() string {
	return `
resource "emporix_delivery_time" "test" {
  name               = "all-zones-delivery"
  site_code          = "main"
  is_delivery_day    = true
  is_for_all_zones   = true
  time_zone_id       = "America/New_York"
  delivery_day_shift = 1

  day = {
    weekday = "MONDAY"
  }

  slots = [
    {
      shipping_method = "standard"
      capacity        = 200

      delivery_time_range = {
        time_from = "08:00:00"
        time_to   = "18:00:00"
      }
    }
  ]
}
`
}
