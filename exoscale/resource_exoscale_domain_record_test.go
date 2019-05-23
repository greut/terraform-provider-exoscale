package exoscale

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/exoscale/egoscale"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDomainRecord(t *testing.T) {
	domain := new(egoscale.DNSDomain)
	record := new(egoscale.DNSRecord)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDomainRecordDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDomainRecordCreate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists("exoscale_domain.exo", domain),
					testAccCheckDomainRecordExists("exoscale_domain_record.mx", domain, record),
					testAccCheckDomainRecord(record),
					testAccCheckDomainRecordAttributes(map[string]schema.SchemaValidateFunc{
						"name":        ValidateString("mail1"),
						"record_type": ValidateString("MX"),
						"content":     ValidateString("mta1"),
						"prio":        ValidateString("10"),
						"ttl":         ValidateString("10"),
					}),
				),
			},
			resource.TestStep{
				Config: testAccDomainRecordUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists("exoscale_domain.exo", domain),
					testAccCheckDomainRecordExists("exoscale_domain_record.mx", domain, record),
					testAccCheckDomainRecord(record),
					testAccCheckDomainRecordAttributes(map[string]schema.SchemaValidateFunc{
						"name":        ValidateString("mail2"),
						"record_type": ValidateString("MX"),
						"content":     ValidateString("mta2"),
						"prio":        ValidateString("20"),
						"ttl":         ValidateString("20"),
					}),
				),
			},
		},
	})
}

func testAccCheckDomainRecordExists(n string, domain *egoscale.DNSDomain, record *egoscale.DNSRecord) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no domain ID is set")
		}

		id, _ := strconv.ParseInt(rs.Primary.ID, 10, 64)

		client := GetDNSClient(testAccProvider.Meta())
		r, err := client.GetRecord(context.TODO(), domain.Name, id)
		if err != nil {
			return err
		}

		*record = *r

		return nil
	}
}

func testAccCheckDomainRecord(record *egoscale.DNSRecord) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if record.TTL == 0 {
			return fmt.Errorf("DNS Domain Record: TTL is zero")
		}

		return nil
	}
}

func testAccCheckDomainRecordAttributes(expected map[string]schema.SchemaValidateFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "exoscale_domain_record" {
				continue
			}

			return checkResourceAttributes(expected, rs.Primary.Attributes)
		}

		return errors.New("domain_record resource not found in the state")
	}
}

func testAccCheckDomainRecordDestroy(s *terraform.State) error {
	client := GetDNSClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "exoscale_domain_record" {
			continue
		}

		id, _ := strconv.ParseInt(rs.Primary.ID, 10, 64)
		d, err := client.GetRecord(context.TODO(), rs.Primary.Attributes["domain"], id)
		if err != nil {
			if _, ok := err.(*egoscale.DNSErrorResponse); ok {
				return nil
			}
			return err
		}
		if d == nil {
			return nil
		}
		return fmt.Errorf("domain record still exists")
	}
	return nil
}

var testAccDomainRecordCreate = fmt.Sprintf(`
resource "exoscale_domain" "exo" {
  name = "%s"
}

resource "exoscale_domain_record" "mx" {
  domain      = "${exoscale_domain.exo.id}"
  name        = "mail1"
  record_type = "MX"
  content     = "mta1"
  prio        = 10
  ttl         = 10
}
`,
	testDomain)

var testAccDomainRecordUpdate = fmt.Sprintf(`
resource "exoscale_domain" "exo" {
  name = "%s"
}

resource "exoscale_domain_record" "mx" {
  domain      = "${exoscale_domain.exo.id}"
  name        = "mail2"
  record_type = "MX"
  content     = "mta2"
  ttl         = 20
  prio        = 20
}
`,
	testDomain)
