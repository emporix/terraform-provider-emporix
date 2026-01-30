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
