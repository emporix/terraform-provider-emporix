package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccWebhookResource_basic(t *testing.T) {
	// Enable force deletion for webhook tests
	os.Setenv("EMPORIX_WEBHOOK_FORCE_DELETE", "true")
	t.Cleanup(func() {
		os.Unsetenv("EMPORIX_WEBHOOK_FORCE_DELETE")
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccWebhookResourceConfig("test_webhook_1", `"HTTP"`, `"<URL>"`, true, nil, nil),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_webhook.test", "code", "test_webhook_1"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "provider_type", "HTTP"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "destination_url", "<URL>"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "active", "true"),
					resource.TestCheckResourceAttrSet("emporix_webhook.test", "version"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "emporix_webhook.test",
				ImportState:                          true,
				ImportStateId:                        "test_webhook_1",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "code",
			},
			// Update testing - change active to true
			{
				Config: testAccWebhookResourceConfig("test_webhook_1", `"HTTP"`, `"<URL>"`, true, nil, nil),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_webhook.test", "code", "test_webhook_1"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "active", "true"),
				),
			},
		},
	})
}

func TestAccWebhookResource_svixProvider(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			// Create with SVIX provider
			{
				Config: testAccWebhookResourceConfig("test_webhook_svix", `"svix"`, `"https://my-app.svix.com"`, false, nil, nil),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_webhook.test", "code", "test_webhook_svix"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "provider_type", "SVIX"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "destination_url", "https://my-app.svix.com"),
				),
			},
		},
	})
}

func TestAccWebhookResource_withSecretKey(t *testing.T) {
	os.Setenv("EMPORIX_WEBHOOK_FORCE_DELETE", "true")
	t.Cleanup(func() {
		os.Unsetenv("EMPORIX_WEBHOOK_FORCE_DELETE")
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			// Create with secret key (HTTP provider)
			{
				Config: testAccWebhookResourceConfigWithSecretKey("test_webhook_secret", `"HTTP"`, `"<URL>"`, true, `"my-secret-key"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_webhook.test", "code", "test_webhook_secret"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "provider_type", "HTTP"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "secret_key_exists", "true"),
				),
			},
		},
	})
}

func TestAccWebhookResource_withHeaders(t *testing.T) {
	os.Setenv("EMPORIX_WEBHOOK_FORCE_DELETE", "true")
	t.Cleanup(func() {
		os.Unsetenv("EMPORIX_WEBHOOK_FORCE_DELETE")
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			// Create with headers
			{
				Config: testAccWebhookResourceConfigWithHeaders("test_webhook_headers", `"HTTP"`, `"<URL>"`, true, map[string]string{"X-Custom-Header": "custom-value", "X-Api-Key": "api-key-123"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_webhook.test", "code", "test_webhook_headers"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "provider_type", "HTTP"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "headers.%", "2"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "headers.X-Custom-Header", "custom-value"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "headers.X-Api-Key", "api-key-123"),
				),
			},
		},
	})
}

func TestAccWebhookResource_withEventsConfiguration(t *testing.T) {
	os.Setenv("EMPORIX_WEBHOOK_FORCE_DELETE", "true")
	t.Cleanup(func() {
		os.Unsetenv("EMPORIX_WEBHOOK_FORCE_DELETE")
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			// Create with events configuration
			{
				Config: testAccWebhookResourceConfigWithEvents("test_webhook_events", `"HTTP"`, `"<URL>"`, true,
					[]testEventConfig{
						{EventType: "order.created", DestinationUrl: "<URL>"},
						{EventType: "customer.created", DestinationUrl: "<URL>"},
					}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_webhook.test", "code", "test_webhook_events"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "provider_type", "HTTP"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.0.event_type", "order.created"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.1.event_type", "customer.created"),
				),
			},
		},
	})
}

func TestAccWebhookResource_requiresReplace(t *testing.T) {
	os.Setenv("EMPORIX_WEBHOOK_FORCE_DELETE", "true")
	t.Cleanup(func() {
		os.Unsetenv("EMPORIX_WEBHOOK_FORCE_DELETE")
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			// Create with code1
			{
				Config: testAccWebhookResourceConfig("test_webhook_code1", `"HTTP"`, `"<URL>"`, true, nil, nil),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_webhook.test", "code", "test_webhook_code1"),
				),
			},
			// Change code (should require replace)
			{
				Config: testAccWebhookResourceConfig("test_webhook_code2", `"HTTP"`, `"<URL>"`, true, nil, nil),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_webhook.test", "code", "test_webhook_code2"),
				),
			},
		},
	})
}

// testEventConfig is a helper type for test config generation (NOT the API model)
type testEventConfig struct {
	EventType      string
	DestinationUrl string
}

// testAccWebhookResourceConfig generates a webhook resource config
func testAccWebhookResourceConfig(code, provider, destinationUrl string, active bool, secretKey *string, headers map[string]string) string {
	secretKeyPart := ""
	if secretKey != nil {
		secretKeyPart = fmt.Sprintf(`secret_key = %q`, *secretKey)
	}

	headersPart := ""
	if len(headers) > 0 {
		headersPart = `headers = {`
		for k, v := range headers {
			headersPart += fmt.Sprintf(`%s = %q,`, k, v)
		}
		headersPart += `}`
	}

	return fmt.Sprintf(`
resource "emporix_webhook" "test" {
  code          = %q
  provider_type      = %s
  destination_url = %s
  active        = %t
  %s
  %s
}
`, code, provider, destinationUrl, active, secretKeyPart, headersPart)
}

// testAccWebhookResourceConfigWithSecretKey generates a webhook resource config with secret key
func testAccWebhookResourceConfigWithSecretKey(code, provider, destinationUrl string, active bool, secretKey string) string {
	return testAccWebhookResourceConfig(code, provider, destinationUrl, active, &secretKey, nil)
}

// testAccWebhookResourceConfigWithHeaders generates a webhook resource config with headers
func testAccWebhookResourceConfigWithHeaders(code, provider, destinationUrl string, active bool, headers map[string]string) string {
	return testAccWebhookResourceConfig(code, provider, destinationUrl, active, nil, headers)
}

// testAccWebhookResourceConfigWithEvents generates a webhook resource config with events configuration
func testAccWebhookResourceConfigWithEvents(code, provider, destinationUrl string, active bool, events []testEventConfig) string {
	eventsBlock := ``
	if len(events) > 0 {
		eventsBlock = `events_configuration = [`
		for i, event := range events {
			if i > 0 {
				eventsBlock += `,`
			}
			if event.DestinationUrl != "" {
				eventsBlock += fmt.Sprintf(`
    {
      event_type      = %q
      destination_url = %q
    }`, event.EventType, event.DestinationUrl)
			} else {
				eventsBlock += fmt.Sprintf(`
    {
      event_type = %q
    }`, event.EventType)
			}
		}
		eventsBlock += `]`
	}

	return fmt.Sprintf(`
resource "emporix_webhook" "test" {
  code          = %q
  provider_type      = %s
  destination_url = %s
  active        = %t
  %s
}
`, code, provider, destinationUrl, active, eventsBlock)
}

// testAccCheckWebhookDestroy verifies that webhooks have been destroyed
func testAccCheckWebhookDestroy(s *terraform.State) error {
	ctx := context.Background()

	// Get configured client
	client, err := getTestClient()
	if err != nil {
		return fmt.Errorf("failed to get test client: %w", err)
	}

	// Iterate through all resources in state
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "emporix_webhook" {
			continue
		}

		code := rs.Primary.Attributes["code"]

		// Try to get the webhook configuration
		_, err := client.GetWebhook(ctx, code)

		// If not found, resource was successfully destroyed
		if IsNotFound(err) {
			continue
		}

		// If other error, fail the test
		if err != nil {
			return fmt.Errorf("unexpected error checking webhook: %w", err)
		}

		// If no error, webhook still exists - try force delete
		_ = client.DeleteWebhook(ctx, code)
		return fmt.Errorf("webhook %s still exists after destroy", code)
	}

	return nil
}
