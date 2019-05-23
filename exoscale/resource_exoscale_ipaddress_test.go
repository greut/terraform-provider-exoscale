package exoscale

import (
	"errors"
	"fmt"
	"testing"

	"github.com/exoscale/egoscale"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

const (
	testEIPHealthcheckMode1              = "http"
	testEIPHealthcheckPort1        int64 = 80
	testEIPHealthcheckPath1              = "/health"
	testEIPHealthcheckInterval1    int64 = 10
	testEIPHealthcheckTimeout1     int64 = 5
	testEIPHealthcheckStrikesOk1   int64 = 1
	testEIPHealthcheckStrikesFail1       = 2
	testEIPHealthcheckMode2              = "http"
	testEIPHealthcheckPort2        int64 = 8000
	testEIPHealthcheckPath2              = "/healthz"
	testEIPHealthcheckInterval2    int64 = 5
	testEIPHealthcheckTimeout2     int64 = 2
	testEIPHealthcheckStrikesOk2   int64 = 2
	testEIPHealthcheckStrikesFail2 int64 = 3
)

func TestAccElasticIP(t *testing.T) {
	eip := new(egoscale.IPAddress)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckElasticIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccElasticIPCreate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckElasticIPExists("exoscale_ipaddress.eip", eip),
					testAccCheckElasticIPCreate(eip),
					testAccCheckElasticIPAttributes(map[string]schema.SchemaValidateFunc{
						"healthcheck_mode":         ValidateString(testEIPHealthcheckMode1),
						"healthcheck_port":         ValidateString(fmt.Sprint(testEIPHealthcheckPort1)),
						"healthcheck_path":         ValidateString(testEIPHealthcheckPath1),
						"healthcheck_interval":     ValidateString(fmt.Sprint(testEIPHealthcheckInterval1)),
						"healthcheck_timeout":      ValidateString(fmt.Sprint(testEIPHealthcheckTimeout1)),
						"healthcheck_strikes_ok":   ValidateString(fmt.Sprint(testEIPHealthcheckStrikesOk1)),
						"healthcheck_strikes_fail": ValidateString(fmt.Sprint(testEIPHealthcheckStrikesFail1)),
					}),
				),
			},
			{
				Config: testAccElasticIPUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckElasticIPExists("exoscale_ipaddress.eip", eip),
					testAccCheckElasticIPUpdate(eip),
					testAccCheckElasticIPAttributes(map[string]schema.SchemaValidateFunc{
						"healthcheck_mode":         ValidateString(testEIPHealthcheckMode2),
						"healthcheck_port":         ValidateString(fmt.Sprint(testEIPHealthcheckPort2)),
						"healthcheck_path":         ValidateString(testEIPHealthcheckPath2),
						"healthcheck_interval":     ValidateString(fmt.Sprint(testEIPHealthcheckInterval2)),
						"healthcheck_timeout":      ValidateString(fmt.Sprint(testEIPHealthcheckTimeout2)),
						"healthcheck_strikes_ok":   ValidateString(fmt.Sprint(testEIPHealthcheckStrikesOk2)),
						"healthcheck_strikes_fail": ValidateString(fmt.Sprint(testEIPHealthcheckStrikesFail2)),
					}),
				),
			},
		},
	})
}

func testAccCheckElasticIPExists(n string, eip *egoscale.IPAddress) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No elastic IP ID is set")
		}

		id, err := egoscale.ParseUUID(rs.Primary.ID)
		if err != nil {
			return err
		}

		client := GetComputeClient(testAccProvider.Meta())
		eip.ID = id
		resp, err := client.Get(eip)
		if err != nil {
			return err
		}

		return Copy(eip, resp.(*egoscale.IPAddress))
	}
}

func testAccCheckElasticIPCreate(eip *egoscale.IPAddress) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if eip.IPAddress == nil {
			return fmt.Errorf("EIP address is nil")
		}

		if eip.Healthcheck == nil {
			return fmt.Errorf("EIP healthcheck is nil")
		}
		if eip.Healthcheck.Mode != testEIPHealthcheckMode1 {
			return fmt.Errorf("expected EIP healthcheck mode %v, got %v",
				testEIPHealthcheckMode1,
				eip.Healthcheck.Mode)
		}
		if eip.Healthcheck.Port != testEIPHealthcheckPort1 {
			return fmt.Errorf("expected EIP healthcheck port %v, got %v",
				testEIPHealthcheckPort1,
				eip.Healthcheck.Port)
		}
		if eip.Healthcheck.Path != testEIPHealthcheckPath1 {
			return fmt.Errorf("expected EIP healthcheck path %v, got %v",
				testEIPHealthcheckPath1,
				eip.Healthcheck.Path)
		}
		if eip.Healthcheck.Interval != testEIPHealthcheckInterval1 {
			return fmt.Errorf("expected EIP healthcheck interval %v, got %v",
				testEIPHealthcheckInterval1,
				eip.Healthcheck.Interval)
		}
		if eip.Healthcheck.Timeout != testEIPHealthcheckTimeout1 {
			return fmt.Errorf("expected EIP healthcheck timeout %v, got %v",
				testEIPHealthcheckTimeout1,
				eip.Healthcheck.Timeout)
		}
		if eip.Healthcheck.StrikesOk != testEIPHealthcheckStrikesOk1 {
			return fmt.Errorf("expected EIP healthcheck strikes-ok %v, got %v",
				testEIPHealthcheckStrikesOk1,
				eip.Healthcheck.StrikesOk)
		}
		if eip.Healthcheck.StrikesFail != testEIPHealthcheckStrikesFail1 {
			return fmt.Errorf("expected EIP healthcheck strikes-fail %v, got %v",
				testEIPHealthcheckStrikesFail1,
				eip.Healthcheck.StrikesFail)
		}

		return nil
	}
}

func testAccCheckElasticIPUpdate(eip *egoscale.IPAddress) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if eip.IPAddress == nil {
			return fmt.Errorf("EIP address is nil")
		}

		if eip.Healthcheck == nil {
			return fmt.Errorf("EIP healthcheck is nil")
		}
		if eip.Healthcheck.Mode != testEIPHealthcheckMode2 {
			return fmt.Errorf("expected EIP healthcheck mode %v, got %v",
				testEIPHealthcheckMode2,
				eip.Healthcheck.Mode)
		}
		if eip.Healthcheck.Port != testEIPHealthcheckPort2 {
			return fmt.Errorf("expected EIP healthcheck port %v, got %v",
				testEIPHealthcheckPort2,
				eip.Healthcheck.Port)
		}
		if eip.Healthcheck.Path != testEIPHealthcheckPath2 {
			return fmt.Errorf("expected EIP healthcheck path %v, got %v",
				testEIPHealthcheckPath2,
				eip.Healthcheck.Path)
		}
		if eip.Healthcheck.Interval != testEIPHealthcheckInterval2 {
			return fmt.Errorf("expected EIP healthcheck interval %v, got %v",
				testEIPHealthcheckInterval2,
				eip.Healthcheck.Interval)
		}
		if eip.Healthcheck.Timeout != testEIPHealthcheckTimeout2 {
			return fmt.Errorf("expected EIP healthcheck timeout %v, got %v",
				testEIPHealthcheckTimeout2,
				eip.Healthcheck.Timeout)
		}
		if eip.Healthcheck.StrikesOk != testEIPHealthcheckStrikesOk2 {
			return fmt.Errorf("expected EIP healthcheck strikes-ok %v, got %v",
				testEIPHealthcheckStrikesOk2,
				eip.Healthcheck.StrikesOk)
		}
		if eip.Healthcheck.StrikesFail != testEIPHealthcheckStrikesFail2 {
			return fmt.Errorf("expected EIP healthcheck strikes-fail %v, got %v",
				testEIPHealthcheckStrikesFail2,
				eip.Healthcheck.StrikesFail)
		}

		return nil
	}
}

func testAccCheckElasticIPAttributes(expected map[string]schema.SchemaValidateFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "exoscale_ipaddress" {
				continue
			}

			return checkResourceAttributes(expected, rs.Primary.Attributes)
		}

		return errors.New("ipaddress resource not found in the state")
	}
}

func testAccCheckElasticIPDestroy(s *terraform.State) error {
	client := GetComputeClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "exoscale_ipaddress" {
			continue
		}

		id, err := egoscale.ParseUUID(rs.Primary.ID)
		if err != nil {
			return err
		}

		key := &egoscale.IPAddress{
			ID:        id,
			IsElastic: true,
		}
		_, err = client.Get(key)
		if err != nil {
			if r, ok := err.(*egoscale.ErrorResponse); ok {
				if r.ErrorCode == egoscale.ParamError {
					return nil
				}
			}
			return err
		}
		return fmt.Errorf("ipAddress: %#v still exists", key)
	}
	return nil
}

var testAccElasticIPCreate = fmt.Sprintf(`
resource "exoscale_ipaddress" "eip" {
  zone = %q
  healthcheck_mode = "%s"
  healthcheck_port = %d
  healthcheck_path = "%s"
  healthcheck_interval = %d
  healthcheck_timeout = %d
  healthcheck_strikes_ok = %d
  healthcheck_strikes_fail = %d
  tags = {
    test = "acceptance"
  }
}
`,
	defaultExoscaleZone,
	testEIPHealthcheckMode1,
	testEIPHealthcheckPort1,
	testEIPHealthcheckPath1,
	testEIPHealthcheckInterval1,
	testEIPHealthcheckTimeout1,
	testEIPHealthcheckStrikesOk1,
	testEIPHealthcheckStrikesFail1,
)

var testAccElasticIPUpdate = fmt.Sprintf(`
resource "exoscale_ipaddress" "eip" {
  zone = %q
  healthcheck_mode = "%s"
  healthcheck_port = %d
  healthcheck_path = "%s"
  healthcheck_interval = %d
  healthcheck_timeout = %d
  healthcheck_strikes_ok = %d
  healthcheck_strikes_fail = %d
}
`,
	defaultExoscaleZone,
	testEIPHealthcheckMode2,
	testEIPHealthcheckPort2,
	testEIPHealthcheckPath2,
	testEIPHealthcheckInterval2,
	testEIPHealthcheckTimeout2,
	testEIPHealthcheckStrikesOk2,
	testEIPHealthcheckStrikesFail2,
)
