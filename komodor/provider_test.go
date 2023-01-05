package komodor

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProvider *schema.Provider
var testAccProviders map[string]*schema.Provider
var testAccProviderFactories map[string]func() (*schema.Provider, error)

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"komodor": testAccProvider,
	}
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"komodor": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("KOMODOR_API_KEY"); v == "" {
		t.Fatal("KOMODOR_API_KEY must be set for acceptance tests")
	}

	// err := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	// if err != nil {
	// 	t.Fatal(err)
	// }
}
