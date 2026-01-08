package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTenantConfigurationResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTenantConfigurationDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTenantConfigurationResourceConfig("test_config_1", `"test_value"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_tenant_configuration.test", "key", "test_config_1"),
					resource.TestCheckResourceAttr("emporix_tenant_configuration.test", "value", `"test_value"`),
					resource.TestCheckResourceAttr("emporix_tenant_configuration.test", "secured", "false"),
					resource.TestCheckResourceAttrSet("emporix_tenant_configuration.test", "version"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "emporix_tenant_configuration.test",
				ImportState:                          true,
				ImportStateId:                        "test_config_1",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
			// Update testing
			{
				Config: testAccTenantConfigurationResourceConfig("test_config_1", `"updated_value"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_tenant_configuration.test", "key", "test_config_1"),
					resource.TestCheckResourceAttr("emporix_tenant_configuration.test", "value", `"updated_value"`),
				),
			},
		},
	})
}

func TestAccTenantConfigurationResource_object(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTenantConfigurationDestroy,
		Steps: []resource.TestStep{
			// Create with object value
			{
				Config: testAccTenantConfigurationResourceConfig("test_config_obj", `{setting = "value", enabled = true}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_tenant_configuration.test", "key", "test_config_obj"),
					resource.TestCheckResourceAttrSet("emporix_tenant_configuration.test", "value"),
				),
			},
		},
	})
}

func TestAccTenantConfigurationResource_secured(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTenantConfigurationDestroy,
		Steps: []resource.TestStep{
			// Create with secured flag
			{
				Config: testAccTenantConfigurationResourceConfigSecured("test_config_secure", `"secret_value"`, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_tenant_configuration.test", "key", "test_config_secure"),
					resource.TestCheckResourceAttr("emporix_tenant_configuration.test", "secured", "true"),
				),
			},
		},
	})
}

func TestAccTenantConfigurationResource_requiresReplace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTenantConfigurationDestroy,
		Steps: []resource.TestStep{
			// Create with key1
			{
				Config: testAccTenantConfigurationResourceConfig("test_config_key1", `"value1"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_tenant_configuration.test", "key", "test_config_key1"),
				),
			},
			// Change key (should require replace)
			{
				Config: testAccTenantConfigurationResourceConfig("test_config_key2", `"value2"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_tenant_configuration.test", "key", "test_config_key2"),
				),
			},
		},
	})
}

// testAccTenantConfigurationResourceConfig generates a tenant configuration
func testAccTenantConfigurationResourceConfig(key, value string) string {
	return fmt.Sprintf(`
resource "emporix_tenant_configuration" "test" {
  key   = %[1]q
  value = jsonencode(%[2]s)
}
`, key, value)
}

// testAccTenantConfigurationResourceConfigSecured generates a tenant configuration with secured flag
func testAccTenantConfigurationResourceConfigSecured(key, value string, secured bool) string {
	return fmt.Sprintf(`
resource "emporix_tenant_configuration" "test" {
  key     = %[1]q
  value   = jsonencode(%[2]s)
  secured = %[3]t
}
`, key, value, secured)
}

// testAccCheckTenantConfigurationDestroy verifies that tenant configurations have been deleted
func testAccCheckTenantConfigurationDestroy(s *terraform.State) error {
	ctx := context.Background()

	// Get configured client
	client, err := getTestClient()
	if err != nil {
		return fmt.Errorf("failed to get test client: %w", err)
	}

	// Iterate through all resources in state
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "emporix_tenant_configuration" {
			continue
		}

		key := rs.Primary.Attributes["key"]

		// Try to get the configuration
		_, err := client.GetTenantConfiguration(ctx, key)

		// If not found, resource was successfully destroyed
		if IsNotFound(err) {
			continue
		}

		// If other error, fail the test
		if err != nil {
			return fmt.Errorf("unexpected error checking tenant configuration: %w", err)
		}

		// If no error, configuration still exists
		return fmt.Errorf("tenant configuration %s still exists after destroy", key)
	}

	return nil
}
