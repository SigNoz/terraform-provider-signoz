package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"signoz": providerserver.NewProtocol6WithError(New("test", "test", "test")()),
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("SIGNOZ_ACCESS_TOKEN"); v == "" {
		t.Skip("SIGNOZ_ACCESS_TOKEN must be set for acceptance tests")
	}
	if v := os.Getenv("SIGNOZ_ENDPOINT"); v == "" {
		t.Skip("SIGNOZ_ENDPOINT must be set for acceptance tests")
	}
}
