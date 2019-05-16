package exoscale

import (
	"fmt"
	"net"
	"testing"

	"github.com/exoscale/egoscale"
	"github.com/hashicorp/terraform/helper/resource"
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
					testAccCheckNICAttributes(nic, net.ParseIP("10.0.0.1")),
					testAccCheckNICCreateAttributes(),
				),
			}, {
				Config: testAccNICUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeExists("exoscale_compute.vm", vm),
					testAccCheckNetworkExists("exoscale_network.net", nw),
					testAccCheckNICExists("exoscale_nic.nic", vm, nic),
					testAccCheckNICAttributes(nic, net.ParseIP("10.0.0.3")),
					testAccCheckNICCreateAttributes(),
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

func testAccCheckNICAttributes(nic *egoscale.Nic, ipAddress net.IP) resource.TestCheckFunc {
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

func testAccCheckNICCreateAttributes() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "exoscale_nic" {
				continue
			}
			_, err := net.ParseMAC(rs.Primary.Attributes["mac_address"])
			if err != nil {
				return fmt.Errorf("Bad MAC address %s", err)
			}

			return nil
		}

		return fmt.Errorf("could not find NIC MAC address")
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
