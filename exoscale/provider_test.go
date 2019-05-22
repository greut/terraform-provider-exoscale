package exoscale

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"exoscale": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	key := os.Getenv("EXOSCALE_API_KEY")
	secret := os.Getenv("EXOSCALE_API_SECRET")
	if key == "" || secret == "" {
		t.Fatal("EXOSCALE_API_KEY and EXOSCALE_API_SECRET must be set for acceptance tests")
	}
}

// testResourceAttributes compares a map of resource attributes against a map
// of expected resource attributes and performs validation on the values.
func testResourceAttributes(want map[string]schema.SchemaValidateFunc, got map[string]string) error {
	for attr, validateFunc := range want {
		v, ok := got[attr]
		if !ok {
			return fmt.Errorf("expected attribute %q not found in map", attr)
		} else if _, es := validateFunc(v, ""); len(es) > 0 {
			for i := range es {
				errors := make([]string, len(es))
				for _, e := range es {
					errors[i] = e.Error()
				}
				return fmt.Errorf("invalid value for attribute %q:\n%s\n", attr, strings.Join(errors, "\n"))
			}
		}
	}

	return nil
}

var defaultExoscaleZone = "ch-gva-2"
var defaultExoscaleTemplate = "Linux Ubuntu 18.04 LTS 64-bit"
var defaultExoscaleNetworkOffering = "PrivNet"
