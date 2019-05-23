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

func TestAccCompute(t *testing.T) {
	sg := new(egoscale.SecurityGroup)
	vm := new(egoscale.VirtualMachine)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeCreate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeExists("exoscale_compute.vm", vm),
					testAccCheckCompute(vm),
					testAccCheckComputeAttributes(map[string]schema.SchemaValidateFunc{
						"template":     ValidateString(defaultExoscaleTemplate),
						"display_name": ValidateString("terraform-test-compute1"),
						"size":         ValidateString("Micro"),
						"disk_size":    ValidateString("12"),
						"key_pair":     ValidateString("terraform-test-keypair"),
						"tags.test":    ValidateString("terraform"),
					}),
				),
			},
			{
				Config: testAccComputeUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists("exoscale_security_group.sg", sg),
					testAccCheckComputeExists("exoscale_compute.vm", vm),
					testAccCheckCompute(vm),
					testAccCheckComputeAttributes(map[string]schema.SchemaValidateFunc{
						"template":          ValidateString(defaultExoscaleTemplate),
						"display_name":      ValidateString("terraform-test-compute2"),
						"size":              ValidateString("Small"),
						"disk_size":         ValidateString("18"),
						"key_pair":          ValidateString("terraform-test-keypair"),
						"security_groups.#": ValidateString("2"),
						"ip6":               ValidateString("true"),
						"user_data":         ValidateString("#cloud-config\npackage_upgrade: true\n"),
					}),
				),
			},
		},
	})
}

func testAccCheckComputeExists(n string, vm *egoscale.VirtualMachine) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Compute ID is set")
		}

		id, err := egoscale.ParseUUID(rs.Primary.ID)
		if err != nil {
			return err
		}

		client := GetComputeClient(testAccProvider.Meta())

		resp, err := client.Get(&egoscale.VirtualMachine{
			ID: id,
		})
		if err != nil {
			return err
		}

		return Copy(vm, resp.(*egoscale.VirtualMachine))
	}
}

func testAccCheckCompute(vm *egoscale.VirtualMachine) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if vm.ID == nil {
			return fmt.Errorf("compute is nil")
		}

		return nil
	}
}

func testAccCheckComputeAttributes(expected map[string]schema.SchemaValidateFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "exoscale_compute" {
				continue
			}

			return checkResourceAttributes(expected, rs.Primary.Attributes)
		}

		return errors.New("compute resource not found in the state")
	}
}

func testAccCheckComputeDestroy(s *terraform.State) error {
	client := GetComputeClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "exoscale_compute" {
			continue
		}

		id, err := egoscale.ParseUUID(rs.Primary.ID)
		if err != nil {
			return err
		}

		vm := &egoscale.VirtualMachine{ID: id}
		_, err = client.Get(vm)
		if err != nil {
			if r, ok := err.(*egoscale.ErrorResponse); ok {
				if r.ErrorCode == egoscale.ParamError {
					return nil
				}
			}
			return err
		}
	}
	return fmt.Errorf("compute still exists")
}

var testAccComputeCreate = fmt.Sprintf(`
resource "exoscale_ssh_keypair" "key" {
  name = "terraform-test-keypair"
}

resource "exoscale_compute" "vm" {
  template = %q
  zone = %q
  display_name = "terraform-test-compute1"
  size = "Micro"
  disk_size = "12"
  key_pair = "${exoscale_ssh_keypair.key.name}"

  tags = {
    test = "terraform"
  }

  timeouts {
    create = "10m"
  }
}
`,
	defaultExoscaleTemplate,
	defaultExoscaleZone,
)

var testAccComputeUpdate = fmt.Sprintf(`
resource "exoscale_ssh_keypair" "key" {
  name = "terraform-test-keypair"
}

resource "exoscale_security_group" "sg" {
  name = "terraform-test-security-group"
}

resource "exoscale_compute" "vm" {
  template = %q
  zone = %q
  display_name = "terraform-test-compute2"
  size = "Small"
  disk_size = "18"
  key_pair = "${exoscale_ssh_keypair.key.name}"

  user_data = <<EOF
#cloud-config
package_upgrade: true
EOF

  security_groups = ["default", "terraform-test-security-group"]

  ip6 = true

  timeouts {
    delete = "30m"
  }

  # Ensure SG exists before we reference it
  depends_on = ["exoscale_security_group.sg"]
}
`,
	defaultExoscaleTemplate,
	defaultExoscaleZone,
)
