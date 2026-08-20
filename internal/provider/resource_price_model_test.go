package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccPriceModelResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPriceModelResourceConfig_basic("tf-acc-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_model.test", "id", "tf-acc-basic"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "includes_tax", "true"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "default", "false"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "name.en", "Standard Pricing"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tier_type", "BASIC"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.#", "1"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.0.min_quantity.quantity", "0"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.0.min_quantity.unit_code", "pc"),
					resource.TestCheckResourceAttrSet("emporix_price_model.test", "tier_definition.tiers.0.id"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "emporix_price_model.test",
				ImportState:                          true,
				ImportStateId:                        "tf-acc-basic",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
			},
		},
	})
}

func TestAccPriceModelResource_optionalFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPriceModelResourceConfig_optionalFields("tf-acc-optional"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_model.test", "default", "true"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "includes_markup", "true"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "description.en", "Has all optional fields set"),
				),
			},
		},
	})
}

func TestAccPriceModelResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelDestroy,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccPriceModelResourceConfig_basic("tf-acc-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_model.test", "includes_tax", "true"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "name.en", "Standard Pricing"),
				),
			},
			// Update name/description/includes_tax
			{
				Config: testAccPriceModelResourceConfig_updated("tf-acc-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_model.test", "includes_tax", "false"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "name.en", "Standard Pricing Updated"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "name.de", "Standardpreis aktualisiert"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "description.en", "Updated description"),
				),
			},
		},
	})
}

func TestAccPriceModelResource_volumeTiers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPriceModelResourceConfig_volume("tf-acc-volume"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tier_type", "VOLUME"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.#", "3"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.0.min_quantity.quantity", "0"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.1.min_quantity.quantity", "10"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.2.min_quantity.quantity", "50"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "measurement_unit.quantity", "1"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "measurement_unit.unit_code", "pc"),
				),
			},
		},
	})
}

func TestAccPriceModelResource_tiered(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPriceModelResourceConfig_tieredStrategy("tf-acc-tiered"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tier_type", "TIERED"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.#", "2"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.0.min_quantity.quantity", "0"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.0.min_quantity.unit_code", "kg"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.1.min_quantity.quantity", "100"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "measurement_unit.unit_code", "kg"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "emporix_price_model.test",
				ImportState:                          true,
				ImportStateId:                        "tf-acc-tiered",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
			},
		},
	})
}

func TestAccPriceModelResource_updateTiersAppend(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelDestroy,
		Steps: []resource.TestStep{
			// Create with 3 tiers: 0, 10, 50
			{
				Config: testAccPriceModelResourceConfig_tiers("tf-acc-tiers-append", []int{0, 10, 50}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.#", "3"),
				),
			},
			// Append a new, strictly-larger tier at the end
			{
				Config: testAccPriceModelResourceConfig_tiers("tf-acc-tiers-append", []int{0, 10, 50, 100}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.#", "4"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.0.min_quantity.quantity", "0"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.1.min_quantity.quantity", "10"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.2.min_quantity.quantity", "50"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.3.min_quantity.quantity", "100"),
					resource.TestCheckResourceAttrSet("emporix_price_model.test", "tier_definition.tiers.3.id"),
				),
			},
		},
	})
}

func TestAccPriceModelResource_updateTiers_insertMiddle(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelDestroy,
		Steps: []resource.TestStep{
			// Create with 3 tiers: 0, 10, 50
			{
				Config: testAccPriceModelResourceConfig_tiers("tf-acc-tiers", []int{0, 10, 50}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.#", "3"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.0.min_quantity.quantity", "0"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.1.min_quantity.quantity", "10"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.2.min_quantity.quantity", "50"),
					resource.TestCheckResourceAttrSet("emporix_price_model.test", "tier_definition.tiers.0.id"),
					resource.TestCheckResourceAttrSet("emporix_price_model.test", "tier_definition.tiers.1.id"),
					resource.TestCheckResourceAttrSet("emporix_price_model.test", "tier_definition.tiers.2.id"),
				),
			},
			// Insert a tier in the middle (5, between 0 and 10)
			{
				Config: testAccPriceModelResourceConfig_tiers("tf-acc-tiers", []int{0, 5, 10, 50}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.#", "4"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.0.min_quantity.quantity", "0"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.1.min_quantity.quantity", "5"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.2.min_quantity.quantity", "10"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.3.min_quantity.quantity", "50"),
					resource.TestCheckResourceAttrSet("emporix_price_model.test", "tier_definition.tiers.0.id"),
					resource.TestCheckResourceAttrSet("emporix_price_model.test", "tier_definition.tiers.1.id"),
					resource.TestCheckResourceAttrSet("emporix_price_model.test", "tier_definition.tiers.2.id"),
					resource.TestCheckResourceAttrSet("emporix_price_model.test", "tier_definition.tiers.3.id"),
				),
			},
			// Remove the inserted tier again - shifts indices back down.
			{
				Config: testAccPriceModelResourceConfig_tiers("tf-acc-tiers", []int{0, 10, 50}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.#", "3"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.0.min_quantity.quantity", "0"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.1.min_quantity.quantity", "10"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.2.min_quantity.quantity", "50"),
				),
			},
		},
	})
}

func TestAccPriceModelResource_forceDelete(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPriceModelResourceConfig_forceDelete("tf-acc-force-delete"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_model.test", "force_delete", "true"),
				),
			},
		},
	})
}

func TestAccPriceModelResource_requiresReplace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPriceModelResourceConfig_basic("tf-acc-replace-a"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_model.test", "id", "tf-acc-replace-a"),
				),
			},
			// Changing id must force replacement (RequiresReplace plan modifier)
			{
				Config: testAccPriceModelResourceConfig_basic("tf-acc-replace-b"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_model.test", "id", "tf-acc-replace-b"),
				),
			},
		},
	})
}

func TestAccPriceModelResource_invalidTierType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPriceModelResourceConfig_invalidTierType("tf-acc-invalid"),
				ExpectError: regexp.MustCompile(`(?i)value must be one of`),
			},
		},
	})
}

func testAccPriceModelResourceConfig_basic(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_model" "test" {
  id           = %[1]q
  includes_tax = true

  name = {
    en = "Standard Pricing"
  }

  tier_definition = {
    tier_type = "BASIC"
    tiers = [
      {
        min_quantity = {
          quantity  = 0
          unit_code = "pc"
        }
      }
    ]
  }

  measurement_unit = {
    quantity  = 1
    unit_code = "pc"
  }
}
`, id)
}

func testAccPriceModelResourceConfig_optionalFields(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_model" "test" {
  id              = %[1]q
  includes_tax    = true
  includes_markup = true
  default         = true

  name = {
    en = "Optional Fields Test"
  }

  description = {
    en = "Has all optional fields set"
  }

  tier_definition = {
    tier_type = "BASIC"
    tiers = [
      {
        min_quantity = {
          quantity  = 0
          unit_code = "pc"
        }
      }
    ]
  }

  measurement_unit = {
    quantity  = 1
    unit_code = "pc"
  }
}
`, id)
}

func testAccPriceModelResourceConfig_updated(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_model" "test" {
  id           = %[1]q
  includes_tax = false

  name = {
    en = "Standard Pricing Updated"
    de = "Standardpreis aktualisiert"
  }

  description = {
    en = "Updated description"
  }

  tier_definition = {
    tier_type = "BASIC"
    tiers = [
      {
        min_quantity = {
          quantity  = 0
          unit_code = "pc"
        }
      }
    ]
  }

  measurement_unit = {
    quantity  = 1
    unit_code = "pc"
  }
}
`, id)
}

func testAccPriceModelResourceConfig_volume(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_model" "test" {
  id           = %[1]q
  includes_tax = true

  name = {
    en = "Volume Pricing"
  }

  tier_definition = {
    tier_type = "VOLUME"
    tiers = [
      {
        min_quantity = {
          quantity  = 0
          unit_code = "pc"
        }
      },
      {
        min_quantity = {
          quantity  = 10
          unit_code = "pc"
        }
      },
      {
        min_quantity = {
          quantity  = 50
          unit_code = "pc"
        }
      }
    ]
  }

  measurement_unit = {
    quantity  = 1
    unit_code = "pc"
  }
}
`, id)
}

func testAccPriceModelResourceConfig_tieredStrategy(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_model" "test" {
  id           = %[1]q
  includes_tax = false

  name = {
    en = "Tiered Pricing"
  }

  tier_definition = {
    tier_type = "TIERED"
    tiers = [
      {
        min_quantity = {
          quantity  = 0
          unit_code = "kg"
        }
      },
      {
        min_quantity = {
          quantity  = 100
          unit_code = "kg"
        }
      }
    ]
  }

  measurement_unit = {
    quantity  = 1
    unit_code = "kg"
  }
}
`, id)
}

// testAccPriceModelResourceConfig_tiers generates a VOLUME price model with the given tier quantities
func testAccPriceModelResourceConfig_tiers(id string, quantities []int) string {
	tiers := make([]string, len(quantities))
	for i, q := range quantities {
		tiers[i] = fmt.Sprintf(`
      {
        min_quantity = {
          quantity  = %d
          unit_code = "pc"
        }
      }`, q)
	}

	return fmt.Sprintf(`
resource "emporix_price_model" "test" {
  id           = %[1]q
  includes_tax = true

  name = {
    en = "Volume Pricing"
  }

  tier_definition = {
    tier_type = "VOLUME"
    tiers = [%[2]s
    ]
  }

  measurement_unit = {
    quantity  = 1
    unit_code = "pc"
  }
}
`, id, strings.Join(tiers, ","))
}

// testAccPriceModelResourceConfig_forceDelete generates a basic price model with force_delete = true
func testAccPriceModelResourceConfig_forceDelete(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_model" "test" {
  id           = %[1]q
  includes_tax = true
  force_delete = true

  name = {
    en = "Force Delete Test"
  }

  tier_definition = {
    tier_type = "BASIC"
    tiers = [
      {
        min_quantity = {
          quantity  = 0
          unit_code = "pc"
        }
      }
    ]
  }

  measurement_unit = {
    quantity  = 1
    unit_code = "pc"
  }
}
`, id)
}

func testAccPriceModelResourceConfig_invalidTierType(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_model" "test" {
  id           = %[1]q
  includes_tax = true

  name = {
    en = "Invalid"
  }

  tier_definition = {
    tier_type = "INVALID"
    tiers = [
      {
        min_quantity = {
          quantity  = 0
          unit_code = "pc"
        }
      }
    ]
  }

  measurement_unit = {
    quantity  = 1
    unit_code = "pc"
  }
}
`, id)
}

// testAccCheckPriceModelDestroy verifies that price models have been deleted
func testAccCheckPriceModelDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := getTestClient()
	if err != nil {
		return fmt.Errorf("failed to get test client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "emporix_price_model" {
			continue
		}

		id := rs.Primary.Attributes["id"]

		maxRetries := 10
		for i := 0; i < maxRetries; i++ {
			_, err := client.GetPriceModel(ctx, id)

			if IsNotFound(err) {
				break
			}

			if err != nil {
				return fmt.Errorf("unexpected error checking price model: %w", err)
			}

			if i == maxRetries-1 {
				return fmt.Errorf("price model %s still exists after destroy (tried %d times)", id, maxRetries)
			}

			time.Sleep(time.Duration(100*(1<<uint(i))) * time.Millisecond)
		}
	}

	return nil
}
