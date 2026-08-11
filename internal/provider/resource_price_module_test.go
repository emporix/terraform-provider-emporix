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

func TestAccPriceModuleResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModuleDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPriceModuleResourceConfig_basic("tf-acc-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_module.test", "id", "tf-acc-basic"),
					resource.TestCheckResourceAttr("emporix_price_module.test", "includes_tax", "true"),
					resource.TestCheckResourceAttr("emporix_price_module.test", "default", "false"),
					resource.TestCheckResourceAttr("emporix_price_module.test", "name.en", "Standard Pricing"),
					resource.TestCheckResourceAttr("emporix_price_module.test", "tier_definition.tier_type", "BASIC"),
					resource.TestCheckResourceAttr("emporix_price_module.test", "tier_definition.tiers.#", "1"),
					resource.TestCheckResourceAttr("emporix_price_module.test", "tier_definition.tiers.0.min_quantity.quantity", "0"),
					resource.TestCheckResourceAttr("emporix_price_module.test", "tier_definition.tiers.0.min_quantity.unit_code", "piece"),
					resource.TestCheckResourceAttrSet("emporix_price_module.test", "tier_definition.tiers.0.id"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "emporix_price_module.test",
				ImportState:                          true,
				ImportStateId:                        "tf-acc-basic",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
			},
		},
	})
}

func TestAccPriceModuleResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModuleDestroy,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccPriceModuleResourceConfig_basic("tf-acc-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_module.test", "includes_tax", "true"),
					resource.TestCheckResourceAttr("emporix_price_module.test", "name.en", "Standard Pricing"),
				),
			},
			// Update name/description/includes_tax
			{
				Config: testAccPriceModuleResourceConfig_updated("tf-acc-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_module.test", "includes_tax", "false"),
					resource.TestCheckResourceAttr("emporix_price_module.test", "name.en", "Standard Pricing Updated"),
					resource.TestCheckResourceAttr("emporix_price_module.test", "name.de", "Standardpreis aktualisiert"),
					resource.TestCheckResourceAttr("emporix_price_module.test", "description.en", "Updated description"),
				),
			},
		},
	})
}

func TestAccPriceModuleResource_volumeTiers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPriceModuleResourceConfig_volume("tf-acc-volume"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_module.test", "tier_definition.tier_type", "VOLUME"),
					resource.TestCheckResourceAttr("emporix_price_module.test", "tier_definition.tiers.#", "3"),
					resource.TestCheckResourceAttr("emporix_price_module.test", "tier_definition.tiers.0.min_quantity.quantity", "0"),
					resource.TestCheckResourceAttr("emporix_price_module.test", "tier_definition.tiers.1.min_quantity.quantity", "10"),
					resource.TestCheckResourceAttr("emporix_price_module.test", "tier_definition.tiers.2.min_quantity.quantity", "50"),
					resource.TestCheckResourceAttr("emporix_price_module.test", "measurement_unit.quantity", "1"),
					resource.TestCheckResourceAttr("emporix_price_module.test", "measurement_unit.unit_code", "piece"),
				),
			},
		},
	})
}

func TestAccPriceModuleResource_requiresReplace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPriceModuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPriceModuleResourceConfig_basic("tf-acc-replace-a"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_module.test", "id", "tf-acc-replace-a"),
				),
			},
			// Changing id must force replacement (RequiresReplace plan modifier)
			{
				Config: testAccPriceModuleResourceConfig_basic("tf-acc-replace-b"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_price_module.test", "id", "tf-acc-replace-b"),
				),
			},
		},
	})
}

func TestAccPriceModuleResource_invalidTierType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPriceModuleResourceConfig_invalidTierType("tf-acc-invalid"),
				ExpectError: regexp.MustCompile(`(?i)value must be one of`),
			},
		},
	})
}

func testAccPriceModuleResourceConfig_basic(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_module" "test" {
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
          unit_code = "piece"
        }
      }
    ]
  }
}
`, id)
}

func testAccPriceModuleResourceConfig_updated(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_module" "test" {
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
          unit_code = "piece"
        }
      }
    ]
  }
}
`, id)
}

func testAccPriceModuleResourceConfig_volume(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_module" "test" {
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
          unit_code = "piece"
        }
      },
      {
        min_quantity = {
          quantity  = 10
          unit_code = "piece"
        }
      },
      {
        min_quantity = {
          quantity  = 50
          unit_code = "piece"
        }
      }
    ]
  }

  measurement_unit = {
    quantity  = 1
    unit_code = "piece"
  }
}
`, id)
}

func testAccPriceModuleResourceConfig_invalidTierType(id string) string {
	return fmt.Sprintf(`
resource "emporix_price_module" "test" {
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
          unit_code = "piece"
        }
      }
    ]
  }
}
`, id)
}

// testAccCheckPriceModuleDestroy verifies that price models have been deleted
func testAccCheckPriceModuleDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := getTestClient()
	if err != nil {
		return fmt.Errorf("failed to get test client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "emporix_price_module" {
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
