package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccWebhookResource_basic(t *testing.T) {
	// Enable force deletion for webhook tests
	os.Setenv("EMPORIX_WEBHOOK_FORCE_DELETE", "true")
	t.Cleanup(func() {
		os.Unsetenv("EMPORIX_WEBHOOK_FORCE_DELETE")
	})

	url := testWebhookDestinationURL(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccWebhookResourceConfig("test_webhook_1", `"HTTP"`, fmt.Sprintf("%q", url), true, nil, nil),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_webhook.test", "code", "test_webhook_1"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "provider_type", "HTTP"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "destination_url", url),
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
				Config: testAccWebhookResourceConfig("test_webhook_1", `"HTTP"`, fmt.Sprintf("%q", url), true, nil, nil),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_webhook.test", "code", "test_webhook_1"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "active", "true"),
				),
			},
		},
	})
}

// TestAccWebhookResource_svixProvider exercises the SVIX provider. Commented out: it
// needs a real Svix account API key (EMPORIX_WEBHOOK_SVIX_API_KEY) - a placeholder value
// fails with a 500 "Client 'svix': Unauthorized" since Emporix authenticates against
// Svix's API with it. No destination_url: SVIX's config only accepts apiKey.
//
// func TestAccWebhookResource_svixProvider(t *testing.T) {
// 	svixApiKey := os.Getenv("EMPORIX_WEBHOOK_SVIX_API_KEY")
// 	if svixApiKey == "" {
// 		t.Skip("EMPORIX_WEBHOOK_SVIX_API_KEY must be set to a real Svix account API key " +
// 			"to test the SVIX provider - Emporix authenticates against Svix's API with it, " +
// 			"so no placeholder value can substitute for a real one")
// 	}
//
// 	os.Setenv("EMPORIX_WEBHOOK_FORCE_DELETE", "true")
// 	t.Cleanup(func() {
// 		os.Unsetenv("EMPORIX_WEBHOOK_FORCE_DELETE")
// 	})
//
// 	resource.Test(t, resource.TestCase{
// 		PreCheck:                 func() { testAccPreCheck(t) },
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 		CheckDestroy:             testAccCheckWebhookDestroy,
// 		Steps: []resource.TestStep{
// 			// Create with SVIX provider
// 			{
// 				Config: fmt.Sprintf(`
// resource "emporix_webhook" "test" {
//   code          = "test_webhook_svix"
//   provider_type = "SVIX"
//   active        = true
//
//   secret_key = %q
// }
// `, svixApiKey),
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					resource.TestCheckResourceAttr("emporix_webhook.test", "code", "test_webhook_svix"),
// 					resource.TestCheckResourceAttr("emporix_webhook.test", "provider_type", "SVIX"),
// 					resource.TestCheckNoResourceAttr("emporix_webhook.test", "destination_url"),
// 					resource.TestCheckResourceAttr("emporix_webhook.test", "secret_key_exists", "true"),
// 				),
// 			},
// 		},
// 	})
// }

func TestAccWebhookResource_withSecretKey(t *testing.T) {
	os.Setenv("EMPORIX_WEBHOOK_FORCE_DELETE", "true")
	t.Cleanup(func() {
		os.Unsetenv("EMPORIX_WEBHOOK_FORCE_DELETE")
	})

	url := testWebhookDestinationURL(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			// Create with secret key (HTTP provider)
			{
				Config: testAccWebhookResourceConfigWithSecretKey("test_webhook_secret", `"HTTP"`, fmt.Sprintf("%q", url), true, `"my-secret-key"`),
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

	url := testWebhookDestinationURL(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			// Create with headers
			{
				Config: testAccWebhookResourceConfigWithHeaders("test_webhook_headers", `"HTTP"`, fmt.Sprintf("%q", url), true, map[string]string{"X-Custom-Header": "custom-value", "X-Api-Key": "api-key-123"}),
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

	url := testWebhookDestinationURL(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			// Create with events configuration
			{
				Config: testAccWebhookResourceConfigWithEvents("test_webhook_events", `"HTTP"`, fmt.Sprintf("%q", url), true,
					[]testEventConfig{
						{EventType: "order.created", DestinationUrl: url},
						{EventType: "customer.created", DestinationUrl: url},
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

// TestAccWebhookResource_eventsConfigurationFromVariable is a regression test for a bug where
// assigning events_configuration from a Terraform variable (whose declared type omits the
// Computed attributes destination_url/secret_key/subscribed) made Terraform widen the value to
// the resource schema's type and mark the whole list Unknown, crashing ValidateConfig. The same
// literal value written directly in the resource block never went through that widening, so the
// bug only showed up when the value came through a variable/module input. PlanOnly is enough
// here: the crash happened during ValidateConfig, before any API call. ExpectNonEmptyPlan is
// needed since this step never applies, so the plan is a genuine "+create", not empty.
func TestAccWebhookResource_eventsConfigurationFromVariable(t *testing.T) {
	url := testWebhookDestinationURL(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
variable "events_configuration" {
  type = list(object({
    event_type = string
    headers    = map(string)
  }))
  default = []
}

resource "emporix_webhook" "test" {
  code             = "test_webhook_events_from_var"
  provider_type    = "HTTP"
  destination_url  = %q
  active           = true

  events_configuration = var.events_configuration
}
`, url),
				ConfigVariables: map[string]config.Variable{
					"events_configuration": config.ListVariable(
						config.ObjectVariable(map[string]config.Variable{
							"event_type": config.StringVariable("order.created"),
							"headers": config.MapVariable(map[string]config.Variable{
								"X-Event-Group": config.StringVariable("orders"),
							}),
						}),
					),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWebhookResource_eventsConfigurationLifecycle(t *testing.T) {
	os.Setenv("EMPORIX_WEBHOOK_FORCE_DELETE", "true")
	t.Cleanup(func() {
		os.Unsetenv("EMPORIX_WEBHOOK_FORCE_DELETE")
	})

	code := "test_webhook_events_lifecycle"
	url := testWebhookDestinationURL(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			// Step 1: create with 3 event subscriptions.
			{
				Config: testAccWebhookResourceConfigWithEvents(code, `"HTTP"`, fmt.Sprintf("%q", url), true,
					[]testEventConfig{
						{EventType: "order.created", DestinationUrl: url},
						{EventType: "customer.created", DestinationUrl: url},
						{EventType: "product.updated", DestinationUrl: url},
					}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.#", "3"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.0.event_type", "order.created"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.1.event_type", "customer.created"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.2.event_type", "product.updated"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.0.subscribed", "true"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.1.subscribed", "true"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.2.subscribed", "true"),
					testAccCheckWebhookEventsConfigurationCount(code, 3),
					testAccCheckEventSubscriptionStatus(
						[]string{"order.created", "customer.created", "product.updated"},
						nil,
					),
				),
			},
			// Step 2: drop "product.updated" - confirm it's actually removed from the
			// API-side configuration and its subscription was unsubscribed, while the
			// other two remain untouched.
			{
				Config: testAccWebhookResourceConfigWithEvents(code, `"HTTP"`, fmt.Sprintf("%q", url), true,
					[]testEventConfig{
						{EventType: "order.created", DestinationUrl: url},
						{EventType: "customer.created", DestinationUrl: url},
					}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.#", "2"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.0.event_type", "order.created"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.1.event_type", "customer.created"),
					testAccCheckWebhookEventsConfigurationCount(code, 2),
					testAccCheckEventSubscriptionStatus(
						[]string{"order.created", "customer.created"},
						[]string{"product.updated"},
					),
				),
			},
			// Step 3: remove events_configuration entirely - confirm the API-side
			// configuration is fully cleared and every event was unsubscribed.
			{
				Config: testAccWebhookResourceConfig(code, `"HTTP"`, fmt.Sprintf("%q", url), true, nil, nil),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("emporix_webhook.test", "events_configuration.0.event_type"),
					testAccCheckWebhookEventsConfigurationCount(code, 0),
					testAccCheckEventSubscriptionStatus(
						nil,
						[]string{"order.created", "customer.created", "product.updated"},
					),
				),
			},
		},
	})
}

func TestAccWebhookResource_eventDestinationUrlFallback(t *testing.T) {
	os.Setenv("EMPORIX_WEBHOOK_FORCE_DELETE", "true")
	t.Cleanup(func() {
		os.Unsetenv("EMPORIX_WEBHOOK_FORCE_DELETE")
	})

	code := "test_webhook_dest_url_fallback"
	parentUrl := testWebhookDestinationURL(t)
	overrideUrl := parentUrl + "?target=2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			// Step 1: one event omits destination_url entirely and must fall back to the
			// parent destination_url, both in state and on the API side. A second event
			// with an explicit override must keep its own value.
			{
				Config: testAccWebhookResourceConfigWithEvents(code, `"HTTP"`, fmt.Sprintf("%q", parentUrl), true,
					[]testEventConfig{
						{EventType: "order.created", DestinationUrl: ""},
						{EventType: "customer.created", DestinationUrl: overrideUrl},
					}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.0.destination_url", parentUrl),
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.1.destination_url", overrideUrl),
					testAccCheckEventDestinationUrl(code, "order.created", parentUrl),
					testAccCheckEventDestinationUrl(code, "customer.created", overrideUrl),
				),
			},
			// Step 2: an event that explicitly sets destination_url = "" must also fall
			// back to the parent, instead of sending a blank value to the API (which the
			// API rejects with "destinationUrl must not be blank").
			{
				Config: fmt.Sprintf(`
resource "emporix_webhook" "test" {
  code             = %q
  provider_type    = "HTTP"
  destination_url  = %q
  active           = true

  events_configuration = [
    {
      event_type      = "order.created"
      destination_url = ""
    }
  ]
}
`, code, parentUrl),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.0.destination_url", parentUrl),
					testAccCheckEventDestinationUrl(code, "order.created", parentUrl),
				),
			},
		},
	})
}

func TestAccWebhookResource_subscribedToggle(t *testing.T) {
	os.Setenv("EMPORIX_WEBHOOK_FORCE_DELETE", "true")
	t.Cleanup(func() {
		os.Unsetenv("EMPORIX_WEBHOOK_FORCE_DELETE")
	})

	code := "test_webhook_subscribed_toggle"
	url := testWebhookDestinationURL(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			// Step 1: create with 2 events, both implicitly subscribed (default true).
			{
				Config: testAccWebhookResourceConfigWithEvents(code, `"HTTP"`, fmt.Sprintf("%q", url), true,
					[]testEventConfig{
						{EventType: "order.created", DestinationUrl: url},
						{EventType: "customer.created", DestinationUrl: url},
					}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.#", "2"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.0.subscribed", "true"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.1.subscribed", "true"),
					testAccCheckWebhookEventsConfigurationCount(code, 2),
					testAccCheckEventSubscriptionStatus(
						[]string{"order.created", "customer.created"},
						nil,
					),
				),
			},
			// Step 2: set subscribed = false on "order.created" while keeping its
			// destination_url configured. This must unsubscribe the event on the API
			// side (via UNSUBSCRIBE) WITHOUT removing its events_configuration entry -
			// unlike dropping the event entirely, the override configuration stays intact.
			{
				Config: testAccWebhookResourceConfigWithEvents(code, `"HTTP"`, fmt.Sprintf("%q", url), true,
					[]testEventConfig{
						{EventType: "order.created", DestinationUrl: url, Subscribed: boolPtr(false)},
						{EventType: "customer.created", DestinationUrl: url},
					}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.#", "2"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.0.subscribed", "false"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.1.subscribed", "true"),
					testAccCheckWebhookEventsConfigurationCount(code, 2),
					testAccCheckEventDestinationUrl(code, "order.created", url),
					testAccCheckEventSubscriptionStatus(
						[]string{"customer.created"},
						[]string{"order.created"},
					),
				),
			},
			// Step 3: re-subscribe "order.created" by setting subscribed back to true.
			{
				Config: testAccWebhookResourceConfigWithEvents(code, `"HTTP"`, fmt.Sprintf("%q", url), true,
					[]testEventConfig{
						{EventType: "order.created", DestinationUrl: url, Subscribed: boolPtr(true)},
						{EventType: "customer.created", DestinationUrl: url},
					}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.#", "2"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.0.subscribed", "true"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.1.subscribed", "true"),
					testAccCheckWebhookEventsConfigurationCount(code, 2),
					testAccCheckEventSubscriptionStatus(
						[]string{"order.created", "customer.created"},
						nil,
					),
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

	url := testWebhookDestinationURL(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			// Create with code1
			{
				Config: testAccWebhookResourceConfig("test_webhook_code1", `"HTTP"`, fmt.Sprintf("%q", url), true, nil, nil),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_webhook.test", "code", "test_webhook_code1"),
				),
			},
			// Change code (should require replace)
			{
				Config: testAccWebhookResourceConfig("test_webhook_code2", `"HTTP"`, fmt.Sprintf("%q", url), true, nil, nil),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_webhook.test", "code", "test_webhook_code2"),
				),
			},
		},
	})
}

// Two entries sharing an event_type must get distinct ids and update independently.
func TestAccWebhookResource_multiTargetSameEventType(t *testing.T) {
	os.Setenv("EMPORIX_WEBHOOK_FORCE_DELETE", "true")
	t.Cleanup(func() {
		os.Unsetenv("EMPORIX_WEBHOOK_FORCE_DELETE")
	})

	code := "test_webhook_multi_target"
	url := testWebhookDestinationURL(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			// Step 1: two entries for the same event_type, routed to different URLs.
			{
				Config: testAccWebhookResourceConfigWithEvents(code, `"HTTP"`, fmt.Sprintf("%q", url), true,
					[]testEventConfig{
						{EventType: "order.created", DestinationUrl: url},
						{EventType: "order.created", DestinationUrl: url + "?target=2"},
					}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.#", "2"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.0.event_type", "order.created"),
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.1.event_type", "order.created"),
					resource.TestCheckResourceAttrSet("emporix_webhook.test", "events_configuration.0.id"),
					resource.TestCheckResourceAttrSet("emporix_webhook.test", "events_configuration.1.id"),
					testAccCheckDistinctEventEntryIds(code, "order.created", 2),
				),
			},
			// Step 2: change only the second entry's URL - both must still exist independently.
			{
				Config: testAccWebhookResourceConfigWithEvents(code, `"HTTP"`, fmt.Sprintf("%q", url), true,
					[]testEventConfig{
						{EventType: "order.created", DestinationUrl: url},
						{EventType: "order.created", DestinationUrl: url + "?target=3"},
					}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("emporix_webhook.test", "events_configuration.#", "2"),
					testAccCheckDistinctEventEntryIds(code, "order.created", 2),
					testAccCheckEventEntryUrls(code, "order.created", []string{url, url + "?target=3"}),
				),
			},
		},
	})
}

func boolPtr(b bool) *bool {
	return &b
}

// testEventConfig is a helper type for test config generation (NOT the API model)
type testEventConfig struct {
	EventType      string
	DestinationUrl string
	Subscribed     *bool
}

// testAccWebhookResourceConfig generates a webhook resource config
func testAccWebhookResourceConfig(code, provider, destinationUrl string, active bool, secretKey *string, headers map[string]string) string {
	secretKeyPart := ""
	if secretKey != nil {
		secretKeyPart = fmt.Sprintf(`secret_key = %q`, *secretKey)
	}

	headersPart := ""
	if len(headers) > 0 {
		headersPart = "headers = {\n"
		for k, v := range headers {
			headersPart += fmt.Sprintf("    %q = %q\n", k, v)
		}
		headersPart += "  }"
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
			destinationUrlLine := ""
			if event.DestinationUrl != "" {
				destinationUrlLine = fmt.Sprintf("\n      destination_url = %q", event.DestinationUrl)
			}
			subscribedLine := ""
			if event.Subscribed != nil {
				subscribedLine = fmt.Sprintf("\n      subscribed = %t", *event.Subscribed)
			}
			eventsBlock += fmt.Sprintf(`
    {
      event_type = %q%s%s
    }`, event.EventType, destinationUrlLine, subscribedLine)
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

// testAccCheckWebhookEventsConfigurationCount verifies, directly against the API (not Terraform state)
func testAccCheckWebhookEventsConfigurationCount(code string, want int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getTestClient()
		if err != nil {
			return fmt.Errorf("failed to get test client: %w", err)
		}

		webhook, err := client.GetWebhook(context.Background(), code)
		if err != nil {
			return fmt.Errorf("failed to get webhook %q: %w", code, err)
		}

		got := 0
		if webhook.Configuration != nil {
			got = len(webhook.Configuration.EventsConfiguration)
		}
		if got != want {
			return fmt.Errorf("webhook %q: expected %d events_configuration entries on the API side, got %d", code, want, got)
		}
		return nil
	}
}

// testAccCheckEventSubscriptionStatus verifies, directly against the API, that the given
// event types are actually SUBSCRIBED and that the given event types are NOT SUBSCRIBED
// (either absent from the list, or reported with a non-SUBSCRIBED status).
func testAccCheckEventSubscriptionStatus(wantSubscribed, wantNotSubscribed []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getTestClient()
		if err != nil {
			return fmt.Errorf("failed to get test client: %w", err)
		}

		entries, err := client.ListEventSubscriptions(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list event subscriptions: %w", err)
		}

		statusByType := make(map[string]string, len(entries))
		for _, e := range entries {
			statusByType[e.Event.Type] = e.Subscription
		}

		for _, eventType := range wantSubscribed {
			if statusByType[eventType] != "SUBSCRIBED" {
				return fmt.Errorf("expected event %q to be SUBSCRIBED, got status %q", eventType, statusByType[eventType])
			}
		}
		for _, eventType := range wantNotSubscribed {
			if status, ok := statusByType[eventType]; ok && status == "SUBSCRIBED" {
				return fmt.Errorf("expected event %q to not be subscribed, but it is SUBSCRIBED", eventType)
			}
		}
		return nil
	}
}

// testAccCheckEventDestinationUrl verifies, directly against the API (not Terraform state)
func testAccCheckEventDestinationUrl(code, eventType, want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getTestClient()
		if err != nil {
			return fmt.Errorf("failed to get test client: %w", err)
		}

		webhook, err := client.GetWebhook(context.Background(), code)
		if err != nil {
			return fmt.Errorf("failed to get webhook %q: %w", code, err)
		}

		if webhook.Configuration == nil {
			return fmt.Errorf("webhook %q: configuration is nil", code)
		}

		for _, event := range webhook.Configuration.EventsConfiguration {
			if event.EventType == eventType {
				if event.DestinationUrl != want {
					return fmt.Errorf("webhook %q: event %q destinationUrl = %q, want %q", code, eventType, event.DestinationUrl, want)
				}
				return nil
			}
		}
		return fmt.Errorf("webhook %q: event %q not found in API configuration", code, eventType)
	}
}

// Verifies wantCount entries exist for eventType, each with a non-empty, unique id.
func testAccCheckDistinctEventEntryIds(code, eventType string, wantCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getTestClient()
		if err != nil {
			return fmt.Errorf("failed to get test client: %w", err)
		}

		webhook, err := client.GetWebhook(context.Background(), code)
		if err != nil {
			return fmt.Errorf("failed to get webhook %q: %w", code, err)
		}
		if webhook.Configuration == nil {
			return fmt.Errorf("webhook %q: configuration is nil", code)
		}

		seen := make(map[string]struct{})
		count := 0
		for _, event := range webhook.Configuration.EventsConfiguration {
			if event.EventType != eventType {
				continue
			}
			count++
			if event.Id == "" {
				return fmt.Errorf("webhook %q: entry for %q has no id", code, eventType)
			}
			if _, dup := seen[event.Id]; dup {
				return fmt.Errorf("webhook %q: duplicate id %q found among %q entries", code, event.Id, eventType)
			}
			seen[event.Id] = struct{}{}
		}
		if count != wantCount {
			return fmt.Errorf("webhook %q: found %d entries for %q, want %d", code, count, eventType, wantCount)
		}
		return nil
	}
}

// Verifies the set of destinationUrls across entries for eventType matches wantUrls (order-independent).
func testAccCheckEventEntryUrls(code, eventType string, wantUrls []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := getTestClient()
		if err != nil {
			return fmt.Errorf("failed to get test client: %w", err)
		}

		webhook, err := client.GetWebhook(context.Background(), code)
		if err != nil {
			return fmt.Errorf("failed to get webhook %q: %w", code, err)
		}
		if webhook.Configuration == nil {
			return fmt.Errorf("webhook %q: configuration is nil", code)
		}

		want := make(map[string]struct{}, len(wantUrls))
		for _, u := range wantUrls {
			want[u] = struct{}{}
		}

		got := make(map[string]struct{})
		for _, event := range webhook.Configuration.EventsConfiguration {
			if event.EventType == eventType {
				got[event.DestinationUrl] = struct{}{}
			}
		}

		if len(got) != len(want) {
			return fmt.Errorf("webhook %q: event %q destination URLs = %v, want %v", code, eventType, keys(got), wantUrls)
		}
		for u := range want {
			if _, ok := got[u]; !ok {
				return fmt.Errorf("webhook %q: event %q missing expected destination URL %q (got %v)", code, eventType, u, keys(got))
			}
		}
		return nil
	}
}

func keys(m map[string]struct{}) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}

// Must be reachable: the API HEAD/OPTIONS-checks destination_url and rejects placeholders.
func testWebhookDestinationURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("EMPORIX_WEBHOOK_TEST_URL")
	if url == "" {
		t.Fatal("EMPORIX_WEBHOOK_TEST_URL must be set to a real, reachable URL (responds " +
			"successfully to HEAD or OPTIONS) for webhook acceptance tests")
	}
	return url
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
