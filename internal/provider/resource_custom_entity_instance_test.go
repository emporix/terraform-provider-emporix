package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccCustomEntityInstanceResource_basic(t *testing.T) {
	typeID := "TEST_CUSTOM_ENTITY_BASIC"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCustomEntityInstanceDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing. ExpectNonEmptyPlan: modified_at/version can't carry
			// UseStateForUnknown, so a post-apply replan shows a diff even with no config changes.
			{
				Config:             testAccCustomEntityInstanceResourceConfig(typeID, `{}`),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("emporix_custom_entity_instance.test", "id"),
					resource.TestCheckResourceAttr("emporix_custom_entity_instance.test", "type", typeID),
					resource.TestCheckResourceAttr("emporix_custom_entity_instance.test", "name.en", "Test Instance"),
					resource.TestCheckResourceAttrSet("emporix_custom_entity_instance.test", "version"),
					resource.TestCheckResourceAttrSet("emporix_custom_entity_instance.test", "created_at"),
					resource.TestCheckResourceAttr("emporix_custom_entity_instance.test", "media.#", "0"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "emporix_custom_entity_instance.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccCustomEntityInstanceImportStateIdFunc("emporix_custom_entity_instance.test"),
				// mixins JSON key order isn't guaranteed to round-trip byte-identically
				ImportStateVerifyIgnore: []string{"mixins"},
			},
			// Update testing
			{
				Config:             testAccCustomEntityInstanceResourceConfigUpdated(typeID, `{}`),
				ExpectNonEmptyPlan: true, // see the note on the create step above
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_custom_entity_instance.test", "name.en", "Updated Test Instance"),
				),
			},
		},
	})
}

// TestAccCustomEntityInstanceResource_ownerMissingUserID verifies owner.user_id is required at plan time.
func TestAccCustomEntityInstanceResource_ownerMissingUserID(t *testing.T) {
	typeID := "TEST_CE_OWNER_NO_USERID"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCustomEntityInstanceResourceConfigOwnerBlock(typeID, `type = "EMPLOYEE"`),
				ExpectError: regexp.MustCompile(`(?i)user_id`),
			},
		},
	})
}

// TestAccCustomEntityInstanceResource_ownerServiceTypeRejected verifies owner.type rejects "SERVICE"
// at plan time - it's response-only and can't be set explicitly.
func TestAccCustomEntityInstanceResource_ownerServiceTypeRejected(t *testing.T) {
	typeID := "TEST_CE_OWNER_SERVICE"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCustomEntityInstanceResourceConfigOwnerBlock(typeID, `type = "SERVICE"`+"\n    user_id = \"svc-1\""),
				ExpectError: regexp.MustCompile(`(?i)value must be one of`),
			},
		},
	})
}

// TestAccCustomEntityInstanceResource_ownerLegalEntityRequiresCustomer verifies legal_entity_id is
// rejected when owner.type isn't CUSTOMER.
func TestAccCustomEntityInstanceResource_ownerLegalEntityRequiresCustomer(t *testing.T) {
	typeID := "TEST_CE_OWNER_LEGAL_ENTITY"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomEntityInstanceResourceConfigOwnerBlock(typeID,
					`type = "EMPLOYEE"`+"\n    user_id = \"employee-1\"\n    legal_entity_id = \"le-1\""),
				ExpectError: regexp.MustCompile(`legal_entity_id can only be provided`),
			},
		},
	})
}

// TestAccCustomEntityInstanceResource_typeRequiresReplace verifies changing "type" forces replacement
// rather than an in-place update, per the RequiresReplace plan modifier on that attribute.
func TestAccCustomEntityInstanceResource_typeRequiresReplace(t *testing.T) {
	typeA := "TEST_CE_REPLACE_TYPE_A"
	typeB := "TEST_CE_REPLACE_TYPE_B"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCustomEntityInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config:             testAccCustomEntityInstanceResourceConfigMovableType(typeA, typeB, typeA),
				ExpectNonEmptyPlan: true, // modified_at/version can't carry UseStateForUnknown
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_custom_entity_instance.movable", "type", typeA),
				),
			},
			{
				Config:             testAccCustomEntityInstanceResourceConfigMovableType(typeA, typeB, typeB),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_custom_entity_instance.movable", "type", typeB),
				),
			},
		},
	})
}

// TestAccCustomEntityInstanceResource_mixinsNestingRequired verifies mixins fields must nest
// under a top-level key equal to the governing emporix_schema's own `id`: a flat object (e.g.
// {"note": "hello"}) is rejected even when "note" is declared, and the same field nested under
// the schema's id succeeds.
func TestAccCustomEntityInstanceResource_mixinsNestingRequired(t *testing.T) {
	typeID := "TEST_CE_MIXIN_NESTING"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCustomEntityInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomEntityInstanceResourceConfigMixins(typeID, false),
				// Terraform's diagnostic renderer word-wraps long detail text with real
				// newlines, so tolerate arbitrary whitespace (including newlines) between
				// words rather than matching a literal space.
				ExpectError: regexp.MustCompile(`(?s)No\s+matching\s+schema\s+found`),
			},
			{
				Config:             testAccCustomEntityInstanceResourceConfigMixins(typeID, true),
				ExpectNonEmptyPlan: true, // modified_at/version can't carry UseStateForUnknown
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("emporix_custom_entity_instance.test", "id"),
				),
			},
		},
	})
}

// testAccCustomEntityInstanceImportStateIdFunc builds the "type:id" import identifier from state,
// since the instance id is server-generated and unknown ahead of time.
func testAccCustomEntityInstanceImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["type"], rs.Primary.Attributes["id"]), nil
	}
}

// testAccCustomEntityInstanceResourceConfig generates a custom entity type plus a basic custom entity instance.
func testAccCustomEntityInstanceResourceConfig(typeID, mixinsJSON string) string {
	return fmt.Sprintf(`
resource "emporix_custom_entity_type" "test" {
  id = %[1]q
  name = {
    en = "Custom Entity Test Type"
  }
}

resource "emporix_custom_entity_instance" "test" {
  type = emporix_custom_entity_type.test.id
  name = {
    en = "Test Instance"
  }
  mixins = %[2]q
}
`, typeID, mixinsJSON)
}

// testAccCustomEntityInstanceResourceConfigUpdated generates an updated name/mixins for the basic instance.
func testAccCustomEntityInstanceResourceConfigUpdated(typeID, mixinsJSON string) string {
	return fmt.Sprintf(`
resource "emporix_custom_entity_type" "test" {
  id = %[1]q
  name = {
    en = "Custom Entity Test Type"
  }
}

resource "emporix_custom_entity_instance" "test" {
  type = emporix_custom_entity_type.test.id
  name = {
    en = "Updated Test Instance"
  }
  mixins = %[2]q
}
`, typeID, mixinsJSON)
}

// testAccCustomEntityInstanceResourceConfigOwnerBlock generates a custom entity type plus an instance
// with an arbitrary owner block body, for testing owner-related validation failures.
func testAccCustomEntityInstanceResourceConfigOwnerBlock(typeID, ownerBody string) string {
	return fmt.Sprintf(`
resource "emporix_custom_entity_type" "test" {
  id = %[1]q
  name = {
    en = "Owner Validation Test Type"
  }
}

resource "emporix_custom_entity_instance" "test" {
  type = emporix_custom_entity_type.test.id
  name = {
    en = "Owner Validation Test Instance"
  }
  owner = {
    %[2]s
  }
}
`, typeID, ownerBody)
}

// testAccCustomEntityInstanceResourceConfigMovableType generates two custom entity types and a single
// instance whose "type" points at activeType - used to verify changing "type" forces replacement.
func testAccCustomEntityInstanceResourceConfigMovableType(typeA, typeB, activeType string) string {
	return fmt.Sprintf(`
resource "emporix_custom_entity_type" "a" {
  id = %[1]q
  name = {
    en = "Movable Type A"
  }
}

resource "emporix_custom_entity_type" "b" {
  id = %[2]q
  name = {
    en = "Movable Type B"
  }
}

resource "emporix_custom_entity_instance" "movable" {
  type = %[3]q
  name = {
    en = "Movable Instance"
  }

  depends_on = [emporix_custom_entity_type.a, emporix_custom_entity_type.b]
}
`, typeA, typeB, activeType)
}

// testAccCustomEntityInstanceResourceConfigMixins generates a type, a schema declaring "note",
// and an instance setting mixins.note either flat (nested=false, rejected) or nested under the
// schema's own id "test-fields-<typeID>" (nested=true, the structure the API expects).
func testAccCustomEntityInstanceResourceConfigMixins(typeID string, nested bool) string {
	mixinsBlock := `note = "hello"`
	if nested {
		mixinsBlock = fmt.Sprintf(`"test-fields-%s" = { note = "hello" }`, typeID)
	}

	return fmt.Sprintf(`
resource "emporix_custom_entity_type" "test" {
  id = %[1]q
  name = {
    en = "Mixin Structure Test Type"
  }
}

resource "emporix_schema" "test_fields" {
  id = "test-fields-%[1]s"
  name = {
    en = "Test Fields"
  }
  types = [emporix_custom_entity_type.test.id]

  attributes = [
    {
      key = "note"
      name = {
        en = "Note"
      }
      type = "TEXT"
      metadata = {
        read_only = false
        localized = false
        required  = false
        nullable  = true
      }
    }
  ]
}

resource "emporix_custom_entity_instance" "test" {
  type = emporix_custom_entity_type.test.id
  name = {
    en = "Mixin Structure Instance"
  }
  mixins = jsonencode({
    %[2]s
  })

  depends_on = [emporix_schema.test_fields]
}
`, typeID, mixinsBlock)
}

// testAccCheckCustomEntityInstanceDestroy verifies that custom entity instances have been deleted
func testAccCheckCustomEntityInstanceDestroy(s *terraform.State) error {
	ctx := context.Background()

	client, err := getTestClient()
	if err != nil {
		return fmt.Errorf("failed to get test client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "emporix_custom_entity_instance" {
			continue
		}

		entityType := rs.Primary.Attributes["type"]
		id := rs.Primary.Attributes["id"]

		_, err := client.GetCustomEntityInstance(ctx, entityType, id)

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
