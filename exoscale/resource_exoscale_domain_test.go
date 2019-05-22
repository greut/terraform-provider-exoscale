package exoscale

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/exoscale/egoscale"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

const (
	testDomain = "terraform-test.exo"
)

func TestAccDomain(t *testing.T) {
	domain := new(egoscale.DNSDomain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDNSDomainCreate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists("exoscale_domain.exo", domain),
					testAccCheckDomain(domain),
					testAccCheckDomainAttributes(map[string]schema.SchemaValidateFunc{
						"name":       ValidateString(testDomain),
						"state":      ValidateString("hosted"),
						"auto_renew": ValidateString("false"),
						"expires_on": ValidateString(""),
						"token":      ValidateRegexp("^[0-9a-f]+$"),
					}),
				),
			},
		},
	})
}

func testAccCheckDomainExists(n string, domain *egoscale.DNSDomain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No domain ID is set")
		}

		client := GetDNSClient(testAccProvider.Meta())
		d, err := client.GetDomain(context.TODO(), rs.Primary.ID)
		if err != nil {
			return err
		}

		*domain = *d

		return nil
	}
}

func testAccCheckDomain(domain *egoscale.DNSDomain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(domain.Token) != 32 {
			return fmt.Errorf("DNS Domain: token length doesn't match")
		}

		return nil
	}
}

func testAccCheckDomainAttributes(expected map[string]schema.SchemaValidateFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "exoscale_domain" {
				continue
			}

			return testResourceAttributes(expected, rs.Primary.Attributes)
		}

		return errors.New("domain resource not found in the state")
	}
}

func testAccCheckDomainDestroy(s *terraform.State) error {
	client := GetDNSClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "exoscale_domain" {
			continue
		}

		d, err := client.GetDomain(context.TODO(), rs.Primary.Attributes["name"])
		if err != nil {
			if _, ok := err.(*egoscale.DNSErrorResponse); ok {
				return nil
			}
			return err
		}
		if d == nil {
			return nil
		}
		return fmt.Errorf("DNS Domain: still exists")
	}
	return nil
}

var testAccDNSDomainCreate = fmt.Sprintf(`
resource "exoscale_domain" "exo" {
  name = "%s"
}
`,
	testDomain)
