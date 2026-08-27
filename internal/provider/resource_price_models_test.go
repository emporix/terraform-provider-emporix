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

// uniqueTestID appends a nanosecond timestamp so a leftover from an interrupted prior run
// (whose CheckDestroy never got to clean it up) can't collide with the current run.
func uniqueTestID(base string) string {
	return fmt.Sprintf("%s-%d", base, time.Now().UnixNano())
}

func TestAccPriceModelsResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelsDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPriceModelsResourceConfig_basic("tf-acc-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_models.test", "id", "tf-acc-basic"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "includes_tax", "true"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "default", "false"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "name.en", "Standard Pricing"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tier_type", "BASIC"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.#", "1"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.0.min_quantity.quantity", "0"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.0.min_quantity.unit_code", "pc"),
					resource.TestCheckResourceAttrSet("emporix_price_models.test", "tier_definition.tiers.0.id"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "emporix_price_models.test",
				ImportState:                          true,
				ImportStateId:                        "tf-acc-basic",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
			},
		},
	})
}

func TestAccPriceModelsResource_optionalFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPriceModelsResourceConfig_optionalFields(uniqueTestID("tf-acc-optional")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_models.test", "includes_markup", "true"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "description.en", "Has all optional fields set"),
				),
			},
		},
	})
}

func TestAccPriceModelsResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelsDestroy,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccPriceModelsResourceConfig_basic("tf-acc-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_models.test", "includes_tax", "true"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "name.en", "Standard Pricing"),
				),
			},
			// Update name/description/includes_tax
			{
				Config: testAccPriceModelsResourceConfig_updated("tf-acc-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_models.test", "includes_tax", "false"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "name.en", "Standard Pricing Updated"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "name.de", "Standardpreis aktualisiert"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "description.en", "Updated description"),
				),
			},
		},
	})
}

func TestAccPriceModelsResource_volumeTiers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPriceModelsResourceConfig_volume("tf-acc-volume"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tier_type", "VOLUME"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.#", "3"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.0.min_quantity.quantity", "0"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.1.min_quantity.quantity", "10"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.2.min_quantity.quantity", "50"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "measurement_unit.quantity", "1"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "measurement_unit.unit_code", "pc"),
				),
			},
		},
	})
}

func TestAccPriceModelsResource_tiered(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPriceModelsResourceConfig_tieredStrategy("tf-acc-tiered"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tier_type", "TIERED"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.#", "2"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.0.min_quantity.quantity", "0"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.0.min_quantity.unit_code", "kg"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.1.min_quantity.quantity", "100"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "measurement_unit.unit_code", "kg"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "emporix_price_models.test",
				ImportState:                          true,
				ImportStateId:                        "tf-acc-tiered",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
			},
		},
	})
}

func TestAccPriceModelsResource_updateTiersAppend(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelsDestroy,
		Steps: []resource.TestStep{
			// Create with 3 tiers: 0, 10, 50
			{
				Config: testAccPriceModelsResourceConfig_tiers("tf-acc-tiers-append", []int{0, 10, 50}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.#", "3"),
				),
			},
			// Append a new, strictly-larger tier at the end
			{
				Config: testAccPriceModelsResourceConfig_tiers("tf-acc-tiers-append", []int{0, 10, 50, 100}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.#", "4"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.0.min_quantity.quantity", "0"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.1.min_quantity.quantity", "10"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.2.min_quantity.quantity", "50"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.3.min_quantity.quantity", "100"),
					resource.TestCheckResourceAttrSet("emporix_price_models.test", "tier_definition.tiers.3.id"),
				),
			},
		},
	})
}

func TestAccPriceModelsResource_updateTiers_insertMiddle(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelsDestroy,
		Steps: []resource.TestStep{
			// Create with 3 tiers: 0, 10, 50
			{
				Config: testAccPriceModelsResourceConfig_tiers("tf-acc-tiers", []int{0, 10, 50}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.#", "3"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.0.min_quantity.quantity", "0"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.1.min_quantity.quantity", "10"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.2.min_quantity.quantity", "50"),
					resource.TestCheckResourceAttrSet("emporix_price_models.test", "tier_definition.tiers.0.id"),
					resource.TestCheckResourceAttrSet("emporix_price_models.test", "tier_definition.tiers.1.id"),
					resource.TestCheckResourceAttrSet("emporix_price_models.test", "tier_definition.tiers.2.id"),
				),
			},
			// Insert a tier in the middle (5, between 0 and 10)
			{
				Config: testAccPriceModelsResourceConfig_tiers("tf-acc-tiers", []int{0, 5, 10, 50}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.#", "4"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.0.min_quantity.quantity", "0"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.1.min_quantity.quantity", "5"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.2.min_quantity.quantity", "10"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.3.min_quantity.quantity", "50"),
					resource.TestCheckResourceAttrSet("emporix_price_models.test", "tier_definition.tiers.0.id"),
					resource.TestCheckResourceAttrSet("emporix_price_models.test", "tier_definition.tiers.1.id"),
					resource.TestCheckResourceAttrSet("emporix_price_models.test", "tier_definition.tiers.2.id"),
					resource.TestCheckResourceAttrSet("emporix_price_models.test", "tier_definition.tiers.3.id"),
				),
			},
			// Remove the inserted tier again - shifts indices back down.
			{
				Config: testAccPriceModelsResourceConfig_tiers("tf-acc-tiers", []int{0, 10, 50}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.#", "3"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.0.min_quantity.quantity", "0"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.1.min_quantity.quantity", "10"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.2.min_quantity.quantity", "50"),
				),
			},
		},
	})
}

func TestAccPriceModelsResource_forceDelete(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPriceModelsResourceConfig_forceDelete("tf-acc-force-delete"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_models.test", "force_delete", "true"),
				),
			},
		},
	})
}

func TestAccPriceModelsResource_requiresReplace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPriceModelsResourceConfig_basic("tf-acc-replace-a"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_models.test", "id", "tf-acc-replace-a"),
				),
			},
			// Changing id must force replacement (RequiresReplace plan modifier)
			{
				Config: testAccPriceModelsResourceConfig_basic("tf-acc-replace-b"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_models.test", "id", "tf-acc-replace-b"),
				),
			},
		},
	})
}

func TestAccPriceModelsResource_generatedId(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPriceModelsResourceConfig_noId(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("emporix_price_models.test", "id"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "name.en", "Generated Id Test"),
				),
			},
		},
	})
}

func TestAccPriceModelsResource_changeTierType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPriceModelsResourceConfig_basic("tf-acc-change-strategy"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tier_type", "BASIC"),
				),
			},
			{
				Config: testAccPriceModelsResourceConfig_tiers("tf-acc-change-strategy", []int{0, 10, 50}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tier_type", "VOLUME"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "tier_definition.tiers.#", "3"),
				),
			},
		},
	})
}

func TestAccPriceModelsResource_updateMeasurementUnit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPriceModelsResourceConfig_basic("tf-acc-mu-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_models.test", "measurement_unit.quantity", "1"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "measurement_unit.unit_code", "pc"),
				),
			},
			{
				Config: testAccPriceModelsResourceConfig_measurementUnit("tf-acc-mu-update", 5, "pc"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_models.test", "measurement_unit.quantity", "5"),
					resource.TestCheckResourceAttr("emporix_price_models.test", "measurement_unit.unit_code", "pc"),
				),
			},
		},
	})
}

func TestAccPriceModelsResource_clearDescription(t *testing.T) {
	id := uniqueTestID("tf-acc-clear-description")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPriceModelsResourceConfig_optionalFields(id),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_models.test", "description.en", "Has all optional fields set"),
				),
			},
			{
				Config: testAccPriceModelsResourceConfig_optionalFieldsNoDescription(id),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("emporix_price_models.test", "description"),
				),
			},
		},
	})
}

// TestAccPriceModelsResource_tierValidationErrors covers every documented tier_definition
// validation rule as a single table, one subtest per rule.
func TestAccPriceModelsResource_tierValidationErrors(t *testing.T) {
	cases := []struct {
		name        string
		config      string
		expectedErr *regexp.Regexp
	}{
		{
			name:        "invalid_tier_type",
			config:      testAccPriceModelsResourceConfig_invalidTierType("tf-acc-invalid"),
			expectedErr: regexp.MustCompile(`(?i)value must be one of`),
		},
		{
			name:        "mismatched_unit_codes",
			config:      testAccPriceModelsResourceConfig_mismatchedUnitCodes("tf-acc-unit-mismatch"),
			expectedErr: regexp.MustCompile(`(?i)unit codes must have the same values`),
		},
		{
			name:        "non_multiple_tier_quantity",
			config:      testAccPriceModelsResourceConfig_nonMultipleTierQuantity("tf-acc-non-multiple"),
			expectedErr: regexp.MustCompile(`(?i)must\s+be\s+a\s+multiple\s+of\s+the\s+measurement\s+unit`),
		},
		{
			name:        "duplicate_tier_quantity",
			config:      testAccPriceModelsResourceConfig_tiers("tf-acc-dup-tier", []int{0, 10, 10}),
			expectedErr: regexp.MustCompile(`(?i)validation error`),
		},
		{
			name:        "tiers_not_ascending",
			config:      testAccPriceModelsResourceConfig_tiers("tf-acc-tiers-order", []int{0, 50, 10}),
			expectedErr: regexp.MustCompile(`(?i)validation error`),
		},
		{
			name:        "first_tier_not_zero",
			config:      testAccPriceModelsResourceConfig_tiers("tf-acc-first-tier", []int{10, 20}),
			expectedErr: regexp.MustCompile(`(?i)validation error`),
		},
		{
			name:        "basic_tier_nonzero",
			config:      testAccPriceModelsResourceConfig_basicTiers("tf-acc-basic-nonzero", []int{5}),
			expectedErr: regexp.MustCompile(`(?i)validation error`),
		},
		{
			name:        "basic_multiple_tiers",
			config:      testAccPriceModelsResourceConfig_basicTiers("tf-acc-basic-multi", []int{0, 10}),
			expectedErr: regexp.MustCompile(`(?i)validation error`),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      tc.config,
						ExpectError: tc.expectedErr,
					},
				},
			})
		})
	}
}

func testAccPriceModelsResourceConfig_basic(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_models" "test" {
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

func testAccPriceModelsResourceConfig_noId() string {
	return `
resource "emporix_price_models" "test" {
  includes_tax = true

  name = {
    en = "Generated Id Test"
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
`
}

func testAccPriceModelsResourceConfig_measurementUnit(id string, quantity int, unitCode string) string {
	return fmt.Sprintf(`
resource "emporix_price_models" "test" {
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
          unit_code = %[3]q
        }
      }
    ]
  }

  measurement_unit = {
    quantity  = %[2]d
    unit_code = %[3]q
  }
}
`, id, quantity, unitCode)
}

func testAccPriceModelsResourceConfig_optionalFields(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_models" "test" {
  id              = %[1]q
  includes_tax    = true
  includes_markup = true

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

// testAccPriceModelsResourceConfig_optionalFieldsNoDescription matches _optionalFields but with
// description removed, so a test can isolate clearing description without also flipping default
// (which, being Optional+Computed with a static default, would revert to false if just omitted).
func testAccPriceModelsResourceConfig_optionalFieldsNoDescription(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_models" "test" {
  id              = %[1]q
  includes_tax    = true
  includes_markup = true

  name = {
    en = "Optional Fields Test"
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

func testAccPriceModelsResourceConfig_updated(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_models" "test" {
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

func testAccPriceModelsResourceConfig_volume(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_models" "test" {
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

func testAccPriceModelsResourceConfig_tieredStrategy(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_models" "test" {
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

// testAccPriceModelsResourceConfig_tiers generates a VOLUME price model with the given tier quantities
func testAccPriceModelsResourceConfig_tiers(id string, quantities []int) string {
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
resource "emporix_price_models" "test" {
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

// testAccPriceModelsResourceConfig_basicTiers generates a BASIC price model with the given tier quantities
func testAccPriceModelsResourceConfig_basicTiers(id string, quantities []int) string {
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
resource "emporix_price_models" "test" {
  id           = %[1]q
  includes_tax = true

  name = {
    en = "Basic Pricing"
  }

  tier_definition = {
    tier_type = "BASIC"
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

// testAccPriceModelsResourceConfig_forceDelete generates a basic price model with force_delete = true
func testAccPriceModelsResourceConfig_forceDelete(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_models" "test" {
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

func testAccPriceModelsResourceConfig_invalidTierType(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_models" "test" {
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

func testAccPriceModelsResourceConfig_mismatchedUnitCodes(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_models" "test" {
  id           = %[1]q
  includes_tax = true

  name = {
    en = "Mismatched Unit Codes"
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
    unit_code = "kg"
  }
}
`, id)
}

func testAccPriceModelsResourceConfig_nonMultipleTierQuantity(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_models" "test" {
  id           = %[1]q
  includes_tax = true

  name = {
    en = "Non-Multiple Tier Quantity"
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
          quantity  = 7
          unit_code = "pc"
        }
      }
    ]
  }

  measurement_unit = {
    quantity  = 10
    unit_code = "pc"
  }
}
`, id)
}

// testAccCheckPriceModelsDestroy verifies that price models have been deleted
func testAccCheckPriceModelsDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := getTestClient()
	if err != nil {
		return fmt.Errorf("failed to get test client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "emporix_price_models" {
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
