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

func TestAccPriceModelResource_generatedId(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPriceModelResourceConfig_noId(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("emporix_price_model.test", "id"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "name.en", "Generated Id Test"),
				),
			},
		},
	})
}

func TestAccPriceModelResource_changeTierType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPriceModelResourceConfig_basic("tf-acc-change-strategy"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tier_type", "BASIC"),
				),
			},
			{
				Config: testAccPriceModelResourceConfig_tiers("tf-acc-change-strategy", []int{0, 10, 50}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tier_type", "VOLUME"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "tier_definition.tiers.#", "3"),
				),
			},
		},
	})
}

func TestAccPriceModelResource_updateMeasurementUnit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPriceModelResourceConfig_basic("tf-acc-mu-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_model.test", "measurement_unit.quantity", "1"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "measurement_unit.unit_code", "pc"),
				),
			},
			{
				Config: testAccPriceModelResourceConfig_measurementUnit("tf-acc-mu-update", 5, "pc"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_model.test", "measurement_unit.quantity", "5"),
					resource.TestCheckResourceAttr("emporix_price_model.test", "measurement_unit.unit_code", "pc"),
				),
			},
		},
	})
}

func TestAccPriceModelResource_clearDescription(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPriceModelResourceConfig_optionalFields("tf-acc-clear-description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_model.test", "description.en", "Has all optional fields set"),
				),
			},
			{
				Config: testAccPriceModelResourceConfig_optionalFieldsNoDescription("tf-acc-clear-description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("emporix_price_model.test", "description"),
				),
			},
		},
	})
}

// TestAccPriceModelResource_tierValidationErrors covers every documented tier_definition
// validation rule as a single table, one subtest per rule. The first three regexes are the
// exact API wording, confirmed live; the rest match the generic validation-error envelope
// ("There is a validation error...") shared by all of them, since their exact wording hasn't
// been confirmed against the live API.
func TestAccPriceModelResource_tierValidationErrors(t *testing.T) {
	cases := []struct {
		name        string
		config      string
		expectedErr *regexp.Regexp
	}{
		{
			name:        "invalid_tier_type",
			config:      testAccPriceModelResourceConfig_invalidTierType("tf-acc-invalid"),
			expectedErr: regexp.MustCompile(`(?i)value must be one of`),
		},
		{
			name:        "mismatched_unit_codes",
			config:      testAccPriceModelResourceConfig_mismatchedUnitCodes("tf-acc-unit-mismatch"),
			expectedErr: regexp.MustCompile(`(?i)unit codes must have the same values`),
		},
		{
			name:        "non_multiple_tier_quantity",
			config:      testAccPriceModelResourceConfig_nonMultipleTierQuantity("tf-acc-non-multiple"),
			expectedErr: regexp.MustCompile(`(?i)must\s+be\s+a\s+multiple\s+of\s+the\s+measurement\s+unit`),
		},
		{
			name:        "duplicate_tier_quantity",
			config:      testAccPriceModelResourceConfig_tiers("tf-acc-dup-tier", []int{0, 10, 10}),
			expectedErr: regexp.MustCompile(`(?i)validation error`),
		},
		{
			name:        "tiers_not_ascending",
			config:      testAccPriceModelResourceConfig_tiers("tf-acc-tiers-order", []int{0, 50, 10}),
			expectedErr: regexp.MustCompile(`(?i)validation error`),
		},
		{
			name:        "first_tier_not_zero",
			config:      testAccPriceModelResourceConfig_tiers("tf-acc-first-tier", []int{10, 20}),
			expectedErr: regexp.MustCompile(`(?i)validation error`),
		},
		{
			name:        "basic_tier_nonzero",
			config:      testAccPriceModelResourceConfig_basicTiers("tf-acc-basic-nonzero", []int{5}),
			expectedErr: regexp.MustCompile(`(?i)validation error`),
		},
		{
			name:        "basic_multiple_tiers",
			config:      testAccPriceModelResourceConfig_basicTiers("tf-acc-basic-multi", []int{0, 10}),
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

func testAccPriceModelResourceConfig_noId() string {
	return `
resource "emporix_price_model" "test" {
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

func testAccPriceModelResourceConfig_measurementUnit(id string, quantity int, unitCode string) string {
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

// testAccPriceModelResourceConfig_optionalFieldsNoDescription matches _optionalFields but with
// description removed, so a test can isolate clearing description without also flipping default
// (which, being Optional+Computed with a static default, would revert to false if just omitted).
func testAccPriceModelResourceConfig_optionalFieldsNoDescription(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_model" "test" {
  id              = %[1]q
  includes_tax    = true
  includes_markup = true
  default         = true

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

// testAccPriceModelResourceConfig_basicTiers generates a BASIC price model with the given tier quantities
func testAccPriceModelResourceConfig_basicTiers(id string, quantities []int) string {
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

func testAccPriceModelResourceConfig_mismatchedUnitCodes(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_model" "test" {
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

func testAccPriceModelResourceConfig_nonMultipleTierQuantity(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_model" "test" {
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
