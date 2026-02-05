package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSchemaResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSchemaDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSchemaResourceConfig("test-schema-basic-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_schema.test", "id", "test-schema-basic-1"),
					resource.TestCheckResourceAttrSet("emporix_schema.test", "name.en"),
					resource.TestCheckResourceAttr("emporix_schema.test", "types.0", "PRODUCT"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.key", "customField"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.type", "TEXT"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "emporix_schema.test",
				ImportState:                          true,
				ImportStateId:                        "test-schema-basic-1",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
				// Ignore attributes because dynamic types have different representations
				// after import (all fields populated) vs config (only specified fields)
				ImportStateVerifyIgnore: []string{"attributes"},
			},
			// Update testing
			{
				Config: testAccSchemaResourceConfigUpdated("test-schema-basic-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_schema.test", "id", "test-schema-basic-1"),
					resource.TestCheckResourceAttr("emporix_schema.test", "name.en", "Updated Product Schema"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.key", "customField"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.1.key", "additionalField"),
				),
			},
		},
	})
}

func TestAccSchemaResource_requiresReplace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSchemaDestroy,
		Steps: []resource.TestStep{
			// Create with first ID
			{
				Config: testAccSchemaResourceConfig("test-schema-replace-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_schema.test", "id", "test-schema-replace-1"),
				),
			},
			// Change ID (should require replace)
			{
				Config: testAccSchemaResourceConfig("test-schema-replace-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_schema.test", "id", "test-schema-replace-2"),
				),
			},
		},
	})
}

func TestAccSchemaResource_complexAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSchemaDestroy,
		Steps: []resource.TestStep{
			// Create with various attribute types
			{
				Config: testAccSchemaResourceConfigComplex("test-schema-complex-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_schema.test", "id", "test-schema-complex-1"),
					// Text attribute
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.key", "textField"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.type", "TEXT"),
					// Number attribute
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.1.key", "numberField"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.1.type", "NUMBER"),
					// Enum attribute with values
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.2.key", "enumField"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.2.type", "ENUM"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.2.values.0.value", "option1"),
				),
			},
		},
	})
}

func TestAccSchemaResource_multipleTypes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSchemaDestroy,
		Steps: []resource.TestStep{
			// Create with multiple types
			{
				Config: testAccSchemaResourceConfigMultipleTypes("test-schema-multitypes-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_schema.test", "id", "test-schema-multitypes-1"),
					resource.TestCheckResourceAttr("emporix_schema.test", "types.0", "PRODUCT"),
					resource.TestCheckResourceAttr("emporix_schema.test", "types.1", "CATEGORY"),
				),
			},
		},
	})
}

func TestAccSchemaResource_multipleSchemas(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSchemaDestroy,
		Steps: []resource.TestStep{
			// Create multiple schemas
			{
				Config: testAccSchemaResourceConfigMultiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_schema.product", "id", "test-product-custom-schema"),
					resource.TestCheckResourceAttr("emporix_schema.customer", "id", "test-customer-custom-schema"),
					resource.TestCheckResourceAttr("emporix_schema.order", "id", "test-order-custom-schema"),
				),
			},
		},
	})
}

func TestAccSchemaResource_nestedObjects(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSchemaDestroy,
		Steps: []resource.TestStep{
			// Create schema with deeply nested OBJECT attributes
			{
				Config: testAccSchemaResourceConfigNestedObjects("test-schema-nested-objects-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_schema.test", "id", "test-schema-nested-objects-1"),
					// Top-level OBJECT attribute
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.key", "address"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.type", "OBJECT"),
					// First-level nested attribute (street)
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.0.key", "street"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.0.type", "TEXT"),
					// First-level nested OBJECT attribute (coordinates)
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.1.key", "coordinates"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.1.type", "OBJECT"),
					// Second-level nested attributes (latitude, longitude)
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.1.attributes.0.key", "latitude"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.1.attributes.0.type", "DECIMAL"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.1.attributes.1.key", "longitude"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.1.attributes.1.type", "DECIMAL"),
					// Nested ENUM within OBJECT
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.2.key", "addressType"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.2.type", "ENUM"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.2.values.0.value", "home"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.2.values.1.value", "work"),
					// Nested ARRAY within OBJECT
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.3.key", "tags"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.3.type", "ARRAY"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.3.array_type.type", "TEXT"),
				),
			},
			// ImportState testing for nested objects
			{
				ResourceName:                         "emporix_schema.test",
				ImportState:                          true,
				ImportStateId:                        "test-schema-nested-objects-1",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
				// Ignore attributes because dynamic types have different representations
				// after import (all fields populated) vs config (only specified fields)
				ImportStateVerifyIgnore: []string{"attributes"},
			},
		},
	})
}

func TestAccSchemaResource_fourLevelNesting(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSchemaDestroy,
		Steps: []resource.TestStep{
			// Create schema with 4 levels of OBJECT nesting
			{
				Config: testAccSchemaResourceConfigFourLevelNesting("test-schema-4level-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_schema.test", "id", "test-schema-4level-1"),
					// Level 1: organization
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.key", "organization"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.type", "OBJECT"),
					// Level 2: department
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.0.key", "department"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.0.type", "OBJECT"),
					// Level 3: team
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.0.attributes.0.key", "team"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.0.attributes.0.type", "OBJECT"),
					// Level 4: basic type attributes (deepest level)
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.0.attributes.0.attributes.0.key", "teamName"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.0.attributes.0.attributes.0.type", "TEXT"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.0.attributes.0.attributes.1.key", "memberName"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.0.attributes.0.attributes.1.type", "TEXT"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.0.attributes.0.attributes.2.key", "memberRole"),
					resource.TestCheckResourceAttr("emporix_schema.test", "attributes.0.attributes.0.attributes.0.attributes.2.type", "TEXT"),
				),
			},
			// ImportState testing for 4-level nesting
			{
				ResourceName:                         "emporix_schema.test",
				ImportState:                          true,
				ImportStateId:                        "test-schema-4level-1",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
				// Ignore attributes because dynamic types have different representations
				// after import (all fields populated) vs config (only specified fields)
				ImportStateVerifyIgnore: []string{"attributes"},
			},
		},
	})
}

// testAccSchemaResourceConfig generates a basic schema configuration
func testAccSchemaResourceConfig(id string) string {
	return fmt.Sprintf(`
resource "emporix_schema" "test" {
  id = %[1]q
  name = {
    en = "Product Schema"
  }
  types = ["PRODUCT"]

  attributes = [
    {
      key = "customField"
      name = {
        en = "Custom Field"
      }
      type = "TEXT"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
    }
  ]
}
`, id)
}

// testAccSchemaResourceConfigUpdated generates an updated schema configuration
func testAccSchemaResourceConfigUpdated(id string) string {
	return fmt.Sprintf(`
resource "emporix_schema" "test" {
  id = %[1]q
  name = {
    en = "Updated Product Schema"
    de = "Aktualisiertes Produktschema"
  }
  types = ["PRODUCT"]

  attributes = [
    {
      key = "customField"
      name = {
        en = "Custom Field"
      }
      type = "TEXT"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
    },
    {
      key = "additionalField"
      name = {
        en = "Additional Field"
      }
      type = "NUMBER"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
    }
  ]
}
`, id)
}

// testAccSchemaResourceConfigComplex generates a schema with complex attributes
func testAccSchemaResourceConfigComplex(id string) string {
	return fmt.Sprintf(`
resource "emporix_schema" "test" {
  id = %[1]q
  name = {
    en = "Complex Schema"
  }
  types = ["CUSTOM_ENTITY"]

  attributes = [
    {
      key = "textField"
      name = {
        en = "Text Field"
      }
      description = {
        en = "A simple text field"
      }
      type = "TEXT"
      metadata = {
        read_only  = false
        localized  = true
        required   = true
        nullable   = false
      }
    },
    {
      key = "numberField"
      name = {
        en = "Number Field"
      }
      type = "NUMBER"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
    },
    {
      key = "enumField"
      name = {
        en = "Enum Field"
      }
      type = "ENUM"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
      values = [
        {
          value = "option1"
        },
        {
          value = "option2"
        }
      ]
    }
  ]
}
`, id)
}

// testAccSchemaResourceConfigMultipleTypes generates a schema with multiple types
func testAccSchemaResourceConfigMultipleTypes(id string) string {
	return fmt.Sprintf(`
resource "emporix_schema" "test" {
  id = %[1]q
  name = {
    en = "Multi-Type Schema"
  }
  types = ["PRODUCT", "CATEGORY"]

  attributes = [
    {
      key = "sharedField"
      name = {
        en = "Shared Field"
      }
      type = "TEXT"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
    }
  ]
}
`, id)
}

// testAccSchemaResourceConfigNestedObjects generates a schema with deeply nested OBJECT attributes
func testAccSchemaResourceConfigNestedObjects(id string) string {
	return fmt.Sprintf(`
resource "emporix_schema" "test" {
  id = %[1]q
  name = {
    en = "Nested Objects Schema"
  }
  types = ["CUSTOMER"]

  attributes = [
    {
      key = "address"
      name = {
        en = "Address"
      }
      description = {
        en = "Customer address with nested structure"
      }
      type = "OBJECT"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
      attributes = [
        {
          key = "street"
          name = {
            en = "Street"
          }
          type = "TEXT"
          metadata = {
            read_only  = false
            localized  = false
            required   = true
            nullable   = false
          }
        },
        {
          key = "coordinates"
          name = {
            en = "Coordinates"
          }
          description = {
            en = "GPS coordinates"
          }
          type = "OBJECT"
          metadata = {
            read_only  = false
            localized  = false
            required   = false
            nullable   = true
          }
          attributes = [
            {
              key = "latitude"
              name = {
                en = "Latitude"
              }
              type = "DECIMAL"
              metadata = {
                read_only  = false
                localized  = false
                required   = true
                nullable   = false
              }
            },
            {
              key = "longitude"
              name = {
                en = "Longitude"
              }
              type = "DECIMAL"
              metadata = {
                read_only  = false
                localized  = false
                required   = true
                nullable   = false
              }
            }
          ]
        },
        {
          key = "addressType"
          name = {
            en = "Address Type"
          }
          type = "ENUM"
          metadata = {
            read_only  = false
            localized  = false
            required   = false
            nullable   = true
          }
          values = [
            {
              value = "home"
            },
            {
              value = "work"
            },
            {
              value = "other"
            }
          ]
        },
        {
          key = "tags"
          name = {
            en = "Tags"
          }
          type = "ARRAY"
          metadata = {
            read_only  = false
            localized  = false
            required   = false
            nullable   = true
          }
          array_type = {
            type      = "TEXT"
            localized = false
          }
        }
      ]
    }
  ]
}
`, id)
}

// testAccSchemaResourceConfigFourLevelNesting generates a schema with 4 levels of OBJECT nesting
func testAccSchemaResourceConfigFourLevelNesting(id string) string {
	return fmt.Sprintf(`
resource "emporix_schema" "test" {
  id = %[1]q
  name = {
    en = "Four Level Nesting Schema"
  }
  types = ["CUSTOM_ENTITY"]

  attributes = [
    {
      key = "organization"
      name = {
        en = "Organization"
      }
      type = "OBJECT"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
      attributes = [
        {
          key = "department"
          name = {
            en = "Department"
          }
          type = "OBJECT"
          metadata = {
            read_only  = false
            localized  = false
            required   = false
            nullable   = true
          }
          attributes = [
            {
              key = "team"
              name = {
                en = "Team"
              }
              type = "OBJECT"
              metadata = {
                read_only  = false
                localized  = false
                required   = false
                nullable   = true
              }
              attributes = [
                {
                  key = "teamName"
                  name = {
                    en = "Team Name"
                  }
                  type = "TEXT"
                  metadata = {
                    read_only  = false
                    localized  = false
                    required   = true
                    nullable   = false
                  }
                },
                {
                  key = "memberName"
                  name = {
                    en = "Member Name"
                  }
                  type = "TEXT"
                  metadata = {
                    read_only  = false
                    localized  = false
                    required   = false
                    nullable   = true
                  }
                },
                {
                  key = "memberRole"
                  name = {
                    en = "Member Role"
                  }
                  type = "TEXT"
                  metadata = {
                    read_only  = false
                    localized  = false
                    required   = false
                    nullable   = true
                  }
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}
`, id)
}

// testAccSchemaResourceConfigMultiple generates a configuration with multiple schemas
func testAccSchemaResourceConfigMultiple() string {
	return `
resource "emporix_schema" "product" {
  id = "test-product-custom-schema"
  name = {
    en = "Product Custom Schema"
  }
  types = ["PRODUCT"]

  attributes = [
    {
      key = "productCustom"
      name = {
        en = "Product Custom"
      }
      type = "TEXT"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
    }
  ]
}

resource "emporix_schema" "customer" {
  id = "test-customer-custom-schema"
  name = {
    en = "Customer Custom Schema"
  }
  types = ["CUSTOMER"]

  attributes = [
    {
      key = "customerCustom"
      name = {
        en = "Customer Custom"
      }
      type = "TEXT"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
    }
  ]
}

resource "emporix_schema" "order" {
  id = "test-order-custom-schema"
  name = {
    en = "Order Custom Schema"
  }
  types = ["ORDER"]

  attributes = [
    {
      key = "orderCustom"
      name = {
        en = "Order Custom"
      }
      type = "BOOLEAN"
      metadata = {
        read_only  = false
        localized  = false
        required   = false
        nullable   = true
      }
    }
  ]
}
`
}

// testAccCheckSchemaDestroy verifies that schemas have been deleted
func testAccCheckSchemaDestroy(s *terraform.State) error {
	ctx := context.Background()

	// Get configured client
	client, err := getTestClient()
	if err != nil {
		return fmt.Errorf("failed to get test client: %w", err)
	}

	// Iterate through all resources in state
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "emporix_schema" {
			continue
		}

		id := rs.Primary.Attributes["id"]

		// Try to get the schema
		_, err := client.GetSchema(ctx, id)

		// If not found, resource was successfully destroyed
		if IsNotFound(err) {
			continue
		}

		// If other error, fail the test
		if err != nil {
			return fmt.Errorf("unexpected error checking schema: %w", err)
		}

		// If no error, schema still exists
		return fmt.Errorf("schema %s still exists after destroy", id)
	}

	return nil
}
