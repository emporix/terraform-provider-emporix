package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccCustomEntityResource_basic(t *testing.T) {
	typeID := "TEST_CUSTOM_ENTITY_BASIC"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCustomEntityDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCustomEntityResourceConfig(typeID, `{"foo":"bar"}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("emporix_custom_entity.test", "id"),
					resource.TestCheckResourceAttr("emporix_custom_entity.test", "type", typeID),
					resource.TestCheckResourceAttr("emporix_custom_entity.test", "name.en", "Test Instance"),
					resource.TestCheckResourceAttrSet("emporix_custom_entity.test", "version"),
					resource.TestCheckResourceAttrSet("emporix_custom_entity.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "emporix_custom_entity.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccCustomEntityImportStateIdFunc("emporix_custom_entity.test"),
				// mixins JSON key order isn't guaranteed to round-trip byte-identically
				ImportStateVerifyIgnore: []string{"mixins"},
			},
			// Update testing
			{
				Config: testAccCustomEntityResourceConfigUpdated(typeID, `{"foo":"baz"}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_custom_entity.test", "name.en", "Updated Test Instance"),
				),
			},
		},
	})
}

func TestAccCustomEntityResource_withOwner(t *testing.T) {
	typeID := "TEST_CUSTOM_ENTITY_OWNER"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCustomEntityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomEntityResourceConfigWithOwner(typeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("emporix_custom_entity.owned", "id"),
					resource.TestCheckResourceAttr("emporix_custom_entity.owned", "owner.type", "EMPLOYEE"),
					resource.TestCheckResourceAttr("emporix_custom_entity.owned", "owner.user_id", "test-employee-1"),
				),
			},
		},
	})
}

// testAccCustomEntityImportStateIdFunc builds the "type:id" import identifier from state,
// since the instance id is server-generated and unknown ahead of time.
func testAccCustomEntityImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["type"], rs.Primary.Attributes["id"]), nil
	}
}

// testAccCustomEntityResourceConfig generates a custom entity type plus a basic custom entity instance.
func testAccCustomEntityResourceConfig(typeID, mixinsJSON string) string {
	return fmt.Sprintf(`
resource "emporix_custom_entity_type" "test" {
  id = %[1]q
  name = {
    en = "Custom Entity Test Type"
  }
}

resource "emporix_custom_entity" "test" {
  type = emporix_custom_entity_type.test.id
  name = {
    en = "Test Instance"
  }
  mixins = %[2]q
}
`, typeID, mixinsJSON)
}

// testAccCustomEntityResourceConfigUpdated generates an updated name/mixins for the basic instance.
func testAccCustomEntityResourceConfigUpdated(typeID, mixinsJSON string) string {
	return fmt.Sprintf(`
resource "emporix_custom_entity_type" "test" {
  id = %[1]q
  name = {
    en = "Custom Entity Test Type"
  }
}

resource "emporix_custom_entity" "test" {
  type = emporix_custom_entity_type.test.id
  name = {
    en = "Updated Test Instance"
  }
  mixins = %[2]q
}
`, typeID, mixinsJSON)
}

// testAccCustomEntityResourceConfigWithOwner generates a custom entity instance with an owner set.
func testAccCustomEntityResourceConfigWithOwner(typeID string) string {
	return fmt.Sprintf(`
resource "emporix_custom_entity_type" "test" {
  id = %[1]q
  name = {
    en = "Custom Entity Owner Test Type"
  }
}

resource "emporix_custom_entity" "owned" {
  type = emporix_custom_entity_type.test.id
  name = {
    en = "Owned Instance"
  }
  owner = {
    type    = "EMPLOYEE"
    user_id = "test-employee-1"
  }
}
`, typeID)
}

// testAccCheckCustomEntityDestroy verifies that custom entity instances have been deleted
func testAccCheckCustomEntityDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := getTestClient()
	if err != nil {
		return fmt.Errorf("failed to get test client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "emporix_custom_entity" {
			continue
		}

		entityType := rs.Primary.Attributes["type"]
		id := rs.Primary.Attributes["id"]

		_, err := client.GetCustomEntity(ctx, entityType, id)

		if IsNotFound(err) {
			continue
		}

		if err != nil {
			return fmt.Errorf("unexpected error checking custom entity: %w", err)
		}

		return fmt.Errorf("custom entity %s:%s still exists after destroy", entityType, id)
	}

	return nil
}
