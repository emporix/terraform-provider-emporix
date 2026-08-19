package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"
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

// TestAccCustomEntityTypeResource_reservedID verifies the API's reserved-word constraint
// (AVAILABILITY and LOCATION are reserved by the platform) is caught at plan time.
func TestAccCustomEntityTypeResource_reservedID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCustomEntityTypeResourceConfig("AVAILABILITY"),
				ExpectError: regexp.MustCompile(`(?i)AVAILABILITY`),
			},
		},
	})
}

// TestAccCustomEntityTypeResource_invalidIDFormat verifies the id regex (must start with an
// uppercase letter, contain only uppercase letters/digits/underscores) is caught at plan time.
func TestAccCustomEntityTypeResource_invalidIDFormat(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCustomEntityTypeResourceConfig("lowercase_id"),
				ExpectError: regexp.MustCompile(`(?i)must start with an uppercase letter`),
			},
		},
	})
}

// TestAccCustomEntityTypeResource_deleteBlockedWhileReferenced verifies a type can't be deleted
// while an instance still references it. Uses the client directly instead of resource.Test:
// Terraform's own dependency graph would always destroy the instance before the type, so this
// state can't be reached through ordinary declarative config.
func TestAccCustomEntityTypeResource_deleteBlockedWhileReferenced(t *testing.T) {
	testAccPreCheck(t)

	client, err := getTestClient()
	if err != nil {
		t.Fatalf("failed to get test client: %v", err)
	}

	ctx := context.Background()
	typeID := "TEST_CE_DELETE_BLOCKED"

	if _, err := client.CreateCustomEntityType(ctx, &CustomEntityTypeCreate{
		ID:   typeID,
		Name: map[string]string{"en": "Delete Blocked Test Type"},
	}); err != nil {
		t.Fatalf("failed to create custom entity type: %v", err)
	}

	instance, err := client.CreateCustomEntityInstance(ctx, typeID, &CustomEntityInstanceCreate{
		Name: map[string]string{"en": "Blocking Instance"},
	})
	if err != nil {
		_ = client.DeleteCustomEntityType(ctx, typeID)
		t.Fatalf("failed to create custom entity instance: %v", err)
	}

	// The type must not be deletable while the instance above still references it.
	err = client.DeleteCustomEntityType(ctx, typeID)
	if err == nil {
		_ = client.DeleteCustomEntityInstance(ctx, typeID, instance.ID)
		_ = client.DeleteCustomEntityType(ctx, typeID)
		t.Fatal("expected deleting the type to fail while an instance still references it, but it succeeded")
	}
	if !strings.Contains(err.Error(), "cannot be deleted while schemas or instances still reference them") {
		t.Errorf("expected a 'still referenced' error, got: %v", err)
	}

	// Clean up the instance, then confirm the type becomes deletable once nothing references it.
	if err := client.DeleteCustomEntityInstance(ctx, typeID, instance.ID); err != nil {
		t.Fatalf("failed to delete blocking instance during cleanup: %v", err)
	}
	if err := client.DeleteCustomEntityType(ctx, typeID); err != nil {
		t.Errorf("expected deleting the type to succeed once the instance was removed, got: %v", err)
	}
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
