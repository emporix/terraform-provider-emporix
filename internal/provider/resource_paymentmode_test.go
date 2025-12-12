package provider

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccPaymentModeResource_basic(t *testing.T) {
	code := fmt.Sprintf("test-pm-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPaymentModeDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPaymentModeResourceConfig(code, true, "INVOICE"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_paymentmode.test", "code", code),
					resource.TestCheckResourceAttr("emporix_paymentmode.test", "active", "true"),
					resource.TestCheckResourceAttr("emporix_paymentmode.test", "payment_provider", "INVOICE"),
					resource.TestCheckResourceAttrSet("emporix_paymentmode.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "emporix_paymentmode.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccPaymentModeResourceConfig(code, false, "INVOICE"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_paymentmode.test", "code", code),
					resource.TestCheckResourceAttr("emporix_paymentmode.test", "active", "false"),
				),
			},
		},
	})
}

func TestAccPaymentModeResource_cashOnDelivery(t *testing.T) {
	code := fmt.Sprintf("test-cod-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPaymentModeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPaymentModeResourceConfig(code, true, "CASH_ON_DELIVERY"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_paymentmode.test", "code", code),
					resource.TestCheckResourceAttr("emporix_paymentmode.test", "payment_provider", "CASH_ON_DELIVERY"),
				),
			},
		},
	})
}

func TestAccPaymentModeResource_withConfiguration(t *testing.T) {
	code := fmt.Sprintf("test-config-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPaymentModeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPaymentModeResourceConfigWithSettings(code),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_paymentmode.test", "code", code),
					resource.TestCheckResourceAttr("emporix_paymentmode.test", "configuration.key1", "value1"),
					resource.TestCheckResourceAttr("emporix_paymentmode.test", "configuration.key2", "value2"),
				),
			},
			// Update configuration
			{
				Config: testAccPaymentModeResourceConfigWithSettingsUpdated(code),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_paymentmode.test", "code", code),
					resource.TestCheckResourceAttr("emporix_paymentmode.test", "configuration.key1", "updated1"),
					resource.TestCheckResourceAttr("emporix_paymentmode.test", "configuration.key3", "value3"),
				),
			},
		},
	})
}

func TestAccPaymentModeResource_requiresReplace(t *testing.T) {
	code1 := fmt.Sprintf("test-replace1-%d", time.Now().Unix())
	code2 := fmt.Sprintf("test-replace2-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPaymentModeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPaymentModeResourceConfig(code1, true, "INVOICE"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_paymentmode.test", "code", code1),
				),
			},
			// Change code (should require replace)
			{
				Config: testAccPaymentModeResourceConfig(code2, true, "INVOICE"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_paymentmode.test", "code", code2),
				),
			},
		},
	})
}

func TestAccPaymentModeResource_multiplePaymentModes(t *testing.T) {
	timestamp := time.Now().Unix()
	code1 := fmt.Sprintf("test-multi1-%d", timestamp)
	code2 := fmt.Sprintf("test-multi2-%d", timestamp)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPaymentModeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPaymentModeResourceConfigMultiple(code1, code2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_paymentmode.invoice", "code", code1),
					resource.TestCheckResourceAttr("emporix_paymentmode.invoice", "payment_provider", "INVOICE"),
					resource.TestCheckResourceAttr("emporix_paymentmode.cod", "code", code2),
					resource.TestCheckResourceAttr("emporix_paymentmode.cod", "payment_provider", "CASH_ON_DELIVERY"),
				),
			},
		},
	})
}

// testAccPaymentModeResourceConfig generates a basic payment mode resource configuration
func testAccPaymentModeResourceConfig(code string, active bool, provider string) string {
	return fmt.Sprintf(`
resource "emporix_paymentmode" "test" {
  code             = %[1]q
  active           = %[2]t
  payment_provider = %[3]q
}
`, code, active, provider)
}

// testAccPaymentModeResourceConfigWithSettings generates a payment mode with configuration
func testAccPaymentModeResourceConfigWithSettings(code string) string {
	return fmt.Sprintf(`
resource "emporix_paymentmode" "test" {
  code             = %[1]q
  active           = true
  payment_provider = "INVOICE"

  configuration = {
    key1 = "value1"
    key2 = "value2"
  }
}
`, code)
}

// testAccPaymentModeResourceConfigWithSettingsUpdated generates an updated payment mode configuration
func testAccPaymentModeResourceConfigWithSettingsUpdated(code string) string {
	return fmt.Sprintf(`
resource "emporix_paymentmode" "test" {
  code             = %[1]q
  active           = true
  payment_provider = "INVOICE"

  configuration = {
    key1 = "updated1"
    key3 = "value3"
  }
}
`, code)
}

// testAccPaymentModeResourceConfigMultiple generates multiple payment mode resources
func testAccPaymentModeResourceConfigMultiple(code1, code2 string) string {
	return fmt.Sprintf(`
resource "emporix_paymentmode" "invoice" {
  code             = %[1]q
  active           = true
  payment_provider = "INVOICE"
}

resource "emporix_paymentmode" "cod" {
  code             = %[2]q
  active           = true
  payment_provider = "CASH_ON_DELIVERY"
}
`, code1, code2)
}

// testAccCheckPaymentModeDestroy verifies that payment modes have been deleted
func testAccCheckPaymentModeDestroy(s *terraform.State) error {
	ctx := context.Background()

	// Get configured client
	client, err := getTestClient()
	if err != nil {
		return fmt.Errorf("failed to get test client: %w", err)
	}

	// Iterate through all resources in state
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "emporix_paymentmode" {
			continue
		}

		id := rs.Primary.ID

		// Try to get the payment mode
		paymentMode, err := client.GetPaymentMode(ctx, id)

		// If we get a 404 or the payment mode is nil, that's what we want
		if err != nil || paymentMode == nil {
			continue
		}

		// If payment mode still exists, that's an error
		return fmt.Errorf("payment mode %s still exists after destroy", id)
	}

	return nil
}
