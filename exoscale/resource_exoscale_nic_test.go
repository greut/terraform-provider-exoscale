package exoscale

import (
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/exoscale/egoscale"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccNIC(t *testing.T) {
	vm := new(egoscale.VirtualMachine)
	nw := new(egoscale.Network)
	nic := new(egoscale.Nic)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNICDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNICCreate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeExists("exoscale_compute.vm", vm),
					testAccCheckNetworkExists("exoscale_network.net", nw),
					testAccCheckNICExists("exoscale_nic.nic", vm, nic),
					testAccCheckNIC(nic, net.ParseIP("10.0.0.1")),
					testAccCheckNICAttributes(map[string]schema.SchemaValidateFunc{
						"mac_address": ValidateRegexp("^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$"),
						"ip_address":  ValidateString("10.0.0.1"),
					}),
				),
			}, {
				Config: testAccNICUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeExists("exoscale_compute.vm", vm),
					testAccCheckNetworkExists("exoscale_network.net", nw),
					testAccCheckNICExists("exoscale_nic.nic", vm, nic),
					testAccCheckNIC(nic, net.ParseIP("10.0.0.3")),
					testAccCheckNICAttributes(map[string]schema.SchemaValidateFunc{
						"mac_address": ValidateRegexp("^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$"),
						"ip_address":  ValidateString("10.0.0.3"),
					}),
				),
			},
		},
	})
}

func testAccCheckNICExists(n string, vm *egoscale.VirtualMachine, nic *egoscale.Nic) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no NIC ID is set")
		}

		id, err := egoscale.ParseUUID(rs.Primary.ID)
		if err != nil {
			return err
		}

		client := GetComputeClient(testAccProvider.Meta())
		nic.VirtualMachineID = vm.ID
		nic.ID = id
		resp, err := client.Get(nic)
		if err != nil {
			return err
		}

		return Copy(nic, resp.(*egoscale.Nic))
	}
}

func testAccCheckNIC(nic *egoscale.Nic, ipAddress net.IP) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if nic.MACAddress == nil {
			return fmt.Errorf("NIC is nil")
		}

		if !nic.IPAddress.Equal(ipAddress) {
			return fmt.Errorf("NIC has bad IP address, got %s, want %s", nic.IPAddress, ipAddress)
		}

		return nil
	}
}

func testAccCheckNICAttributes(expected map[string]schema.SchemaValidateFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "exoscale_nic" {
				continue
			}

			return checkResourceAttributes(expected, rs.Primary.Attributes)
		}

		return errors.New("nic resource not found in the state")
	}
}

func testAccCheckNICDestroy(s *terraform.State) error {
	client := GetComputeClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "exoscale_nic" {
			continue
		}

		vmID, err := egoscale.ParseUUID(rs.Primary.Attributes["compute_id"])
		if err != nil {
			return err
		}

		nic := &egoscale.Nic{VirtualMachineID: vmID}
		_, err = client.Get(nic)
		if err != nil {
			if r, ok := err.(*egoscale.ErrorResponse); ok {
				if r.ErrorCode == egoscale.ParamError {
					return nil
				}
			}
			return err
		}
	}
	return fmt.Errorf("NIC still exists")
}

var testAccNICCreate = fmt.Sprintf(`
resource "exoscale_ssh_keypair" "key" {
  name = "terraform-test-keypair"
}

resource "exoscale_compute" "vm" {
  display_name = "terraform-test-compute"
  template = %q
  zone = %q
  size = "Micro"
  disk_size = "12"
  key_pair = "${exoscale_ssh_keypair.key.name}"

  timeouts {
    create = "10m"
    delete = "30m"
  }
}

resource "exoscale_network" "net" {
  name = "terraform-test-network"
  display_text = "Terraform Acceptance Test"
  zone = %q
  network_offering = %q

  start_ip = "10.0.0.1"
  end_ip = "10.0.0.1"
  netmask = "255.255.255.252"
}

resource "exoscale_nic" "nic" {
  compute_id = "${exoscale_compute.vm.id}"
  network_id = "${exoscale_network.net.id}"

  ip_address = "10.0.0.1"
}
`,
	defaultExoscaleTemplate,
	defaultExoscaleZone,
	defaultExoscaleZone,
	defaultExoscaleNetworkOffering,
)

var testAccNICUpdate = fmt.Sprintf(`
resource "exoscale_ssh_keypair" "key" {
  name = "terraform-test-keypair"
}

resource "exoscale_compute" "vm" {
  display_name = "terraform-test-compute"
  template = %q
  zone = %q
  size = "Micro"
  disk_size = "12"
  key_pair = "${exoscale_ssh_keypair.key.name}"

  timeouts {
    create = "10m"
    delete = "30m"
  }
}

resource "exoscale_network" "net" {
  name = "terraform-test-network"
  display_text = "Terraform Acceptance Test"
  zone = %q
  network_offering = %q

  start_ip = "10.0.0.1"
  end_ip = "10.0.0.1"
  netmask = "255.255.255.248"
}

resource "exoscale_nic" "nic" {
  compute_id = "${exoscale_compute.vm.id}"
  network_id = "${exoscale_network.net.id}"

  ip_address = "10.0.0.3"
}
`,
	defaultExoscaleTemplate,
	defaultExoscaleZone,
	defaultExoscaleZone,
	defaultExoscaleNetworkOffering,
)
