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
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "zone_id", "zone-test"),
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
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "slots.0.shipping_method", "standard-shipping"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "slots.0.capacity", "50"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "slots.1.shipping_method", "express-shipping"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "slots.1.capacity", "30"),
				),
			},
		},
	})
}

func TestAccDeliveryTimeResource_nextDay(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryTimeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryTimeResourceConfig_nextDay(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("emporix_delivery_time.test", "id"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "name", "monday-delivery"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "zone_id", "zone-test"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "delivery_day_shift", "1"),
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

func TestAccDeliveryTimeResource_specificDate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryTimeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryTimeResourceConfig_specificDate(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("emporix_delivery_time.test", "id"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "name", "christmas-2024"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "day.date", "2024-12-25T11:00:00.000Z"),
				),
			},
		},
	})
}

func TestAccDeliveryTimeResource_dateRange(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDeliveryTimeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryTimeResourceConfig_dateRange(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("emporix_delivery_time.test", "id"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "name", "summer-2024"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "day.date_from", "2024-06-01T10:00:00.000Z"),
					resource.TestCheckResourceAttr("emporix_delivery_time.test", "day.date_to", "2024-08-31T10:00:00.000Z"),
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
resource "emporix_shipping_zone" "test" {
  id   = "zone-test"
  site = "main"

  name = {
    en = "Test Zone"
  }

  ship_to = [
    { country = "PL" }
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
        currency = "PLN"
      }
      cost = {
        amount   = 15.00
        currency = "PLN"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.test]
}

resource "emporix_delivery_time" "test" {
  name               = "friday-slots"
  site_code          = "main"
  is_delivery_day    = true
  zone_id            = emporix_shipping_zone.test.id
  time_zone_id       = "Europe/Warsaw"
  delivery_day_shift = 0

  day = {
    weekday = "FRIDAY"
  }

  slots = [
    {
      shipping_method = emporix_shipping_method.test.id
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

  depends_on = [
    emporix_shipping_zone.test,
    emporix_shipping_method.test
  ]
}
`
}

func testAccDeliveryTimeResourceConfig_updated() string {
	return `
resource "emporix_shipping_zone" "test" {
  id   = "zone-test"
  site = "main"

  name = {
    en = "Test Zone"
  }

  ship_to = [
    { country = "PL" }
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
        currency = "PLN"
      }
      cost = {
        amount   = 15.00
        currency = "PLN"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.test]
}

resource "emporix_delivery_time" "test" {
  name               = "friday-slots"
  site_code          = "main"
  is_delivery_day    = true
  zone_id            = emporix_shipping_zone.test.id
  time_zone_id       = "Europe/Warsaw"
  delivery_day_shift = 0

  day = {
    weekday = "FRIDAY"
  }

  slots = [
    {
      shipping_method = emporix_shipping_method.test.id
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

  depends_on = [
    emporix_shipping_zone.test,
    emporix_shipping_method.test
  ]
}
`
}

func testAccDeliveryTimeResourceConfig_multipleSlots() string {
	return `
resource "emporix_shipping_zone" "test" {
  id   = "zone-test"
  site = "main"

  name = {
    en = "Test Zone"
  }

  ship_to = [
    { country = "PL" }
  ]
}

resource "emporix_shipping_method" "standard" {
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
        currency = "PLN"
      }
      cost = {
        amount   = 15.00
        currency = "PLN"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.test]
}

resource "emporix_shipping_method" "express" {
  id      = "express-shipping"
  site    = "main"
  zone_id = emporix_shipping_zone.test.id

  name = {
    en = "Express Shipping"
  }

  active = true

  fees = [
    {
      min_order_value = {
        amount   = 0
        currency = "PLN"
      }
      cost = {
        amount   = 25.00
        currency = "PLN"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.test]
}

resource "emporix_delivery_time" "test" {
  name               = "saturday-slots"
  site_code          = "main"
  is_delivery_day    = true
  zone_id            = emporix_shipping_zone.test.id
  time_zone_id       = "Europe/Warsaw"
  delivery_day_shift = 0

  day = {
    weekday = "SATURDAY"
  }

  slots = [
    {
      shipping_method = emporix_shipping_method.standard.id
      capacity        = 50

      delivery_time_range = {
        time_from = "10:00:00"
        time_to   = "12:00:00"
      }

      cut_off_time = {
        time                = "2023-06-13T06:00:00.000Z"
        delivery_cycle_name = "morning"
      }
    },
    {
      shipping_method = emporix_shipping_method.express.id
      capacity        = 30

      delivery_time_range = {
        time_from = "14:00:00"
        time_to   = "16:00:00"
      }

      cut_off_time = {
        time                = "2023-06-13T11:00:00.000Z"
        delivery_cycle_name = "afternoon"
      }
    }
  ]

  depends_on = [
    emporix_shipping_zone.test,
    emporix_shipping_method.standard,
    emporix_shipping_method.express
  ]
}
`
}

func testAccDeliveryTimeResourceConfig_nextDay() string {
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
        amount   = 9.99
        currency = "USD"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.test]
}

resource "emporix_delivery_time" "test" {
  name               = "monday-delivery"
  site_code          = "main"
  is_delivery_day    = true
  zone_id            = emporix_shipping_zone.test.id
  time_zone_id       = "America/New_York"
  delivery_day_shift = 1

  day = {
    weekday = "MONDAY"
  }

  slots = [
    {
      shipping_method = emporix_shipping_method.test.id
      capacity        = 200

      delivery_time_range = {
        time_from = "08:00:00"
        time_to   = "18:00:00"
      }

      cut_off_time = {
        time                = "2023-06-11T15:00:00.000Z"
        delivery_cycle_name = "next-day"
      }
    }
  ]

  depends_on = [
    emporix_shipping_zone.test,
    emporix_shipping_method.test
  ]
}
`
}

func testAccDeliveryTimeResourceConfig_specificDate() string {
	return `
resource "emporix_shipping_zone" "test" {
  id   = "zone-test"
  site = "main"

  name = {
    en = "Test Zone"
  }

  ship_to = [
    { country = "PL" }
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
        currency = "PLN"
      }
      cost = {
        amount   = 15.00
        currency = "PLN"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.test]
}

resource "emporix_delivery_time" "test" {
  name               = "christmas-2024"
  site_code          = "main"
  is_delivery_day    = true
  zone_id            = emporix_shipping_zone.test.id
  time_zone_id       = "Europe/Warsaw"
  delivery_day_shift = 0

  day = {
    date = "2024-12-25T11:00:00.000Z"
  }

  slots = [
    {
      shipping_method = emporix_shipping_method.test.id
      capacity        = 20

      delivery_time_range = {
        time_from = "09:00:00"
        time_to   = "12:00:00"
      }

      cut_off_time = {
        time                = "2024-12-24T06:00:00.000Z"
        delivery_cycle_name = "christmas"
      }
    }
  ]

  depends_on = [
    emporix_shipping_zone.test,
    emporix_shipping_method.test
  ]
}
`
}

func testAccDeliveryTimeResourceConfig_dateRange() string {
	return `
resource "emporix_shipping_zone" "test" {
  id   = "zone-test"
  site = "main"

  name = {
    en = "Test Zone"
  }

  ship_to = [
    { country = "PL" }
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
        currency = "PLN"
      }
      cost = {
        amount   = 15.00
        currency = "PLN"
      }
    }
  ]

  depends_on = [emporix_shipping_zone.test]
}

resource "emporix_delivery_time" "test" {
  name               = "summer-2024"
  site_code          = "main"
  is_delivery_day    = true
  zone_id            = emporix_shipping_zone.test.id
  time_zone_id       = "Europe/Warsaw"
  delivery_day_shift = 0

  day = {
    date_from = "2024-06-01T10:00:00.000Z"
    date_to   = "2024-08-31T10:00:00.000Z"
  }

  slots = [
    {
      shipping_method = emporix_shipping_method.test.id
      capacity        = 100

      delivery_time_range = {
        time_from = "08:00:00"
        time_to   = "20:00:00"
      }

      cut_off_time = {
        time                = "2024-06-01T06:00:00.000Z"
        delivery_cycle_name = "summer"
      }
    }
  ]

  depends_on = [
    emporix_shipping_zone.test,
    emporix_shipping_method.test
  ]
}
`
}
