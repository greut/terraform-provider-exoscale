package exoscale

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/exoscale/egoscale"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDomainRecord(t *testing.T) {
	domain := new(egoscale.DNSDomain)
	record := new(egoscale.DNSRecord)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDNSRecordDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDNSRecordCreate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSDomainExists("exoscale_domain.exo", domain),
					testAccCheckDNSRecordExists("exoscale_domain_record.mx", domain, record),
					testAccCheckDNSRecord(record),
					testAccCheckDNSRecordAttributes(map[string]string{
						"name":        "mail1",
						"record_type": "MX",
						"content":     "mta1",
						"prio":        "10",
						"ttl":         "10",
					}),
				),
			},
			resource.TestStep{
				Config: testAccDNSRecordUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSDomainExists("exoscale_domain.exo", domain),
					testAccCheckDNSRecordExists("exoscale_domain_record.mx", domain, record),
					testAccCheckDNSRecord(record),
					testAccCheckDNSRecordAttributes(map[string]string{
						"name":        "mail2",
						"record_type": "MX",
						"content":     "mta2",
						"prio":        "20",
						"ttl":         "20",
					}),
				),
			},
		},
	})
}

func testAccCheckDNSRecordExists(n string, domain *egoscale.DNSDomain, record *egoscale.DNSRecord) resource.TestCheckFunc {
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

func testAccCheckDNSRecord(record *egoscale.DNSRecord) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if record.TTL == 0 {
			return fmt.Errorf("DNS Domain Record: TTL is zero")
		}

		return nil
	}
}

func testAccCheckDNSRecordAttributes(expected map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "exoscale_domain_record" {
				continue
			}

			return testResourceAttributes(expected, rs.Primary.Attributes)
		}

		return fmt.Errorf("Could not find domain record")
	}
}

func testAccCheckDNSRecordDestroy(s *terraform.State) error {
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

var testAccDNSRecordCreate = `
resource "exoscale_domain" "exo" {
  name = "acceptance.exo"
}

resource "exoscale_domain_record" "mx" {
  domain      = "${exoscale_domain.exo.id}"
  name        = "mail1"
  record_type = "MX"
  content     = "mta1"
  prio        = 10
  ttl         = 10
}
`

var testAccDNSRecordUpdate = `
resource "exoscale_domain" "exo" {
  name = "acceptance.exo"
}

resource "exoscale_domain_record" "mx" {
  domain      = "${exoscale_domain.exo.id}"
  name        = "mail2"
  record_type = "MX"
  content     = "mta2"
  ttl         = 20
  prio        = 20
}
`
