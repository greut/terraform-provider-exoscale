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

func TestAccAffinityGroup(t *testing.T) {
	ag := new(egoscale.AffinityGroup)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAffinityGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAffinityGroupCreate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAffinityGroupExists("exoscale_affinity.ag", ag),
					testAccCheckAffinityGroup(ag),
					testAccCheckAffinityGroupAttributes(map[string]schema.SchemaValidateFunc{
						"name":        ValidateString("terraform-test-affinity"),
						"description": ValidateString("Terraform Acceptance Test"),
						"type":        ValidateString("host anti-affinity"),
					}),
				),
			},
		},
	})
}

func testAccCheckAffinityGroupExists(n string, ag *egoscale.AffinityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Affinity Group ID is set")
		}

		id, err := egoscale.ParseUUID(rs.Primary.ID)
		if err != nil {
			return err
		}

		client := GetComputeClient(testAccProvider.Meta())

		ag.ID = id
		resp, err := client.Get(ag)
		if err != nil {
			return err
		}

		return Copy(ag, resp.(*egoscale.AffinityGroup))
	}
}

func testAccCheckAffinityGroup(ag *egoscale.AffinityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ag.ID == nil {
			return fmt.Errorf("affinity group is nil")
		}

		return nil
	}
}

func testAccCheckAffinityGroupAttributes(expected map[string]schema.SchemaValidateFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "exoscale_affinity" {
				continue
			}

			return testResourceAttributes(expected, rs.Primary.Attributes)
		}

		return errors.New("affinity resource not found in the state")
	}
}

func testAccCheckAffinityGroupDestroy(s *terraform.State) error {
	client := GetComputeClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "exoscale_affinity" {
			continue
		}

		id, err := egoscale.ParseUUID(rs.Primary.ID)
		if err != nil {
			return err
		}

		key := &egoscale.AffinityGroup{ID: id}
		_, err = client.Get(key)
		if err != nil {
			if r, ok := err.(*egoscale.ErrorResponse); ok {
				if r.ErrorCode == egoscale.ParamError {
					return nil
				}
			}
			return err
		}
	}
	return fmt.Errorf("AffinityGroup: still exists")
}

var testAccAffinityGroupCreate = `
resource "exoscale_affinity" "ag" {
  name = "terraform-test-affinity"
  description = "Terraform Acceptance Test"
}
`
