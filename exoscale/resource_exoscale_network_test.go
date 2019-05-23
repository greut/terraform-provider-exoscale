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

func TestAccNetwork(t *testing.T) {
	network := new(egoscale.Network)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkCreate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkExists("exoscale_network.net", network),
					testAccCheckNetwork(network, nil),
					testAccCheckNetworkAttributes(map[string]schema.SchemaValidateFunc{
						"display_text":   ValidateString("Terraform Acceptance Test (create)"),
						"tags.managedby": ValidateString("terraform"),
					}),
				),
			}, {
				Config: testAccNetworkUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkExists("exoscale_network.net", network),
					testAccCheckNetwork(network, net.ParseIP("10.0.0.1")),
					testAccCheckNetworkAttributes(map[string]schema.SchemaValidateFunc{
						"display_text": ValidateString("Terraform Acceptance Test (update)"),
						"start_ip":     ValidateString("10.0.0.1"),
						"end_ip":       ValidateString("10.0.0.5"),
						"netmask":      ValidateString("255.0.0.0"),
					}),
				),
			},
		},
	})
}

func testAccCheckNetworkExists(name string, network *egoscale.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network ID is set")
		}

		id, err := egoscale.ParseUUID(rs.Primary.ID)
		if err != nil {
			return err
		}

		client := GetComputeClient(testAccProvider.Meta())
		network.ID = id
		network.Name = "" // Reset network name to avoid side-effects from previous test steps
		resp, err := client.Get(network)
		if err != nil {
			return err
		}

		return Copy(network, resp.(*egoscale.Network))
	}
}

func testAccCheckNetwork(network *egoscale.Network, expectedStartIP net.IP) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if network.ID == nil {
			return fmt.Errorf("Network is nil")
		}

		if !network.StartIP.Equal(expectedStartIP) {
			return fmt.Errorf("expected StartIP to be %v, got %v", expectedStartIP, network.StartIP)
		}

		return nil
	}
}

func testAccCheckNetworkAttributes(expected map[string]schema.SchemaValidateFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "exoscale_network" {
				continue
			}

			return checkResourceAttributes(expected, rs.Primary.Attributes)
		}

		return errors.New("network resource not found in the state")
	}
}

func testAccCheckNetworkDestroy(s *terraform.State) error {
	client := GetComputeClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "exoscale_network" {
			continue
		}

		id, err := egoscale.ParseUUID(rs.Primary.ID)
		if err != nil {
			return err
		}

		key := &egoscale.Network{ID: id}
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
	return errors.New("Network: still exists")
}

var testAccNetworkCreate = fmt.Sprintf(`
resource "exoscale_network" "net" {
  zone = %q
  network_offering = %q
  name = "terraform-test-network1"
  display_text = "Terraform Acceptance Test (create)"

  tags = {
    managedby = "terraform"
  }
}
`,
	defaultExoscaleZone,
	defaultExoscaleNetworkOffering,
)

var testAccNetworkUpdate = fmt.Sprintf(`
resource "exoscale_network" "net" {
  zone = %q
  network_offering = %q
  name = "terraform-test-network2"
  display_text = "Terraform Acceptance Test (update)"

  start_ip = "10.0.0.1"
  end_ip = "10.0.0.5"
  netmask = "255.0.0.0"
}
`,
	defaultExoscaleZone,
	defaultExoscaleNetworkOffering,
)
