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

// TestAccPriceModelResource_optionalFields exercises default, includes_markup, and
// description on a create - every other test relies on their schema defaults (false / null),
// so the "!IsNull()" branches in priceModelToAPI/priceModelFromAPI for these three fields
// have otherwise never actually run with a non-default value.
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

// TestAccPriceModelResource_tiered exercises the TIERED strategy end-to-end (create, read,
// import). BASIC/VOLUME are already covered elsewhere; this closes the one remaining
// pricing strategy so all three documented tier_type values have actual create+import
// coverage, not just schema validation.
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
			// ImportState testing - first time any multi-tier + measurement_unit
			// combination gets verified through import, not just BASIC's single tier.
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

// TestAccPriceModelResource_updateTiersAppend exercises the tier-addition pattern the
// Emporix admin UI actually supports: appending a new tier whose quantity is strictly
// larger than every existing tier (the UI reportedly won't let you type a smaller value
// than the current last tier at all). No existing tier's list position changes here - this
// isolates whether "append a larger tier" alone hits the general read-after-write revert
// bug, independent of the mid-list-insertion question TestAccPriceModelResource_updateTiers_insertMiddle covers.
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
			// Append a new, strictly-larger tier at the end (100) - every existing tier
			// keeps its original list position; nothing shifts.
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

// TestAccPriceModelResource_updateTiers_insertMiddle exercises inserting and removing a
// tier in the *middle* of an existing VOLUME price model's tier list - something the
// Emporix admin UI reportedly does not allow at all (it only lets you append a strictly
// larger tier). This may not be a supported operation independent of the general
// read-after-write revert bug; comparing this test's result against
// TestAccPriceModelResource_updateTiersAppend's is what tells the two apart. Since each
// tier's "id" is Computed and correlated by list position, inserting a tier in the middle
// also shifts every subsequent tier's index - this verifies that shift doesn't corrupt
// state or fail the update on top of whatever the insertion itself does.
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
			// Insert a tier in the middle (5, between 0 and 10) - shifts the old tiers at
			// index 1 and 2 to index 2 and 3.
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

// TestAccPriceModelResource_forceDelete exercises the force_delete=true path on Delete,
// verifying the API actually accepts the forceDelete query param and returns a status
// code our client handles (worry raised in review: async force-delete might return 202
// instead of the 204 our client currently requires).
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
			// No further steps - resource.Test's implicit final destroy (asserted by
			// CheckDestroy) is what actually exercises force_delete=true on DELETE.
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

// testAccPriceModelResourceConfig_tiers generates a VOLUME price model with an arbitrary,
// ordered list of tier quantities - used to exercise inserting/removing tiers on update.
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

// testAccPriceModelResourceConfig_forceDelete generates a basic price model with
// force_delete = true, so that resource.Test's implicit teardown exercises the
// forceDelete=true DELETE code path.
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
