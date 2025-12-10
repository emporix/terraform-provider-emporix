package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"emporix": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates that required environment variables are set before
// running acceptance tests.
func testAccPreCheck(t *testing.T) {
	// Check TF_ACC is set
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for acceptance tests")
	}

	// Check required Emporix environment variables
	requiredEnvVars := []string{
		"EMPORIX_TENANT",
		"EMPORIX_CLIENT_ID",
		"EMPORIX_CLIENT_SECRET",
	}

	for _, envVar := range requiredEnvVars {
		if v := os.Getenv(envVar); v == "" {
			t.Fatalf("%s must be set for acceptance tests", envVar)
		}
	}
}

// testAccProviderConfig returns a basic provider configuration for use in tests.
// It reads credentials from environment variables.
func testAccProviderConfig() string {
	return `
provider "emporix" {
  # Credentials read from environment variables:
  # EMPORIX_TENANT
  # EMPORIX_CLIENT_ID
  # EMPORIX_CLIENT_SECRET
}
`
}
