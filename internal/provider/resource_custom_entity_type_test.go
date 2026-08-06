package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccCustomEntityTypeResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCustomEntityTypeDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCustomEntityTypeResourceConfig("TEST_DOCUMENT_BASIC"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_custom_entity_type.test", "id", "TEST_DOCUMENT_BASIC"),
					resource.TestCheckResourceAttrSet("emporix_custom_entity_type.test", "name.en"),
					resource.TestCheckResourceAttrSet("emporix_custom_entity_type.test", "version"),
					resource.TestCheckResourceAttrSet("emporix_custom_entity_type.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "emporix_custom_entity_type.test",
				ImportState:                          true,
				ImportStateId:                        "TEST_DOCUMENT_BASIC",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
			},
			// Update testing
			{
				Config: testAccCustomEntityTypeResourceConfigUpdated("TEST_DOCUMENT_BASIC"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_custom_entity_type.test", "id", "TEST_DOCUMENT_BASIC"),
					resource.TestCheckResourceAttr("emporix_custom_entity_type.test", "name.en", "Updated Document Type"),
				),
			},
		},
	})
}

func TestAccCustomEntityTypeResource_requiresReplace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCustomEntityTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomEntityTypeResourceConfig("TEST_DOCUMENT_REPLACE_1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_custom_entity_type.test", "id", "TEST_DOCUMENT_REPLACE_1"),
				),
			},
			{
				Config: testAccCustomEntityTypeResourceConfig("TEST_DOCUMENT_REPLACE_2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_custom_entity_type.test", "id", "TEST_DOCUMENT_REPLACE_2"),
				),
			},
		},
	})
}

// testAccCustomEntityTypeResourceConfig generates a basic custom entity type configuration
func testAccCustomEntityTypeResourceConfig(id string) string {
	return fmt.Sprintf(`
resource "emporix_custom_entity_type" "test" {
  id = %[1]q
  name = {
    en = "Document Type"
  }
}
`, id)
}

// testAccCustomEntityTypeResourceConfigUpdated generates an updated custom entity type configuration
func testAccCustomEntityTypeResourceConfigUpdated(id string) string {
	return fmt.Sprintf(`
resource "emporix_custom_entity_type" "test" {
  id = %[1]q
  name = {
    en = "Updated Document Type"
    de = "Aktualisierter Dokumenttyp"
  }
}
`, id)
}

// testAccCheckCustomEntityTypeDestroy verifies that custom entity types have been deleted
func testAccCheckCustomEntityTypeDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := getTestClient()
	if err != nil {
		return fmt.Errorf("failed to get test client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "emporix_custom_entity_type" {
			continue
		}

		id := rs.Primary.Attributes["id"]

		_, err := client.GetCustomEntityType(ctx, id)

		if IsNotFound(err) {
			continue
		}

		if err != nil {
			return fmt.Errorf("unexpected error checking custom entity type: %w", err)
		}

		return fmt.Errorf("custom entity type %s still exists after destroy", id)
	}

	return nil
}
