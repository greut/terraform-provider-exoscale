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

func TestAccSSHKeyPair(t *testing.T) {
	sshkey := new(egoscale.SSHKeyPair)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSSHKeyPairDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccSSHKeyPairCreate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSSHKeyPairExists("exoscale_ssh_keypair.key", sshkey),
					testAccCheckSSHKeyPair(sshkey),
					testAccCheckSSHKeyPairAttributes(map[string]schema.SchemaValidateFunc{
						"public_key":  ValidateString("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDN7L45b4vO2ytH68ZUC5PMS1b7JG78zGslwcJ0zolE5BuxsCYor248/FKGC5TXrME+yBu/uLqaAkioq4Wp1PzP6Zy5jEowWQDOdeER7uu1GgZShcvly2Oaf/UKLqTdwL+U3tCknqHY63fOAi1lBwmNTUu1uZ24iNiogfhXwQn7HJLQK9vfoGwg+/qJIzeswR6XDa6qh0fuzdxWQ4JWHw2s8fv8WvGOlklmAg/uEi1kF5D6R7kJpOVaE20FLnT4sjA81iErhMIH77OaZqQKiyVH3i5m/lkQI/2e25ml8aculaWzHOs4ctd7l+K1ZWFYje3qMBY1sar1gd787eaqk6RZ"),
						"fingerprint": ValidateString("4d:31:21:c4:77:9f:19:91:6e:84:9d:7c:12:a8:11:1f"),
					}),
				),
			},
		},
	})
}

func testAccCheckSSHKeyPairExists(n string, sshkey *egoscale.SSHKeyPair) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No key pair ID is set")
		}

		client := GetComputeClient(testAccProvider.Meta())
		sshkey.Name = rs.Primary.ID
		resp, err := client.Get(sshkey)
		if err != nil {
			return err
		}

		return Copy(sshkey, resp.(*egoscale.SSHKeyPair))
	}
}

func testAccCheckSSHKeyPair(sshkey *egoscale.SSHKeyPair) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(sshkey.Fingerprint) != 47 {
			return fmt.Errorf("SSH Key: fingerprint length doesn't match")
		}

		return nil
	}
}

func testAccCheckSSHKeyPairAttributes(expected map[string]schema.SchemaValidateFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "exoscale_ssh_keypair" {
				continue
			}

			return checkResourceAttributes(expected, rs.Primary.Attributes)
		}

		return errors.New("ssh_keypair resource not found in the state")
	}
}

func testAccCheckSSHKeyPairDestroy(s *terraform.State) error {
	client := GetComputeClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "exoscale_ssh_keypair" {
			continue
		}

		key := &egoscale.SSHKeyPair{Name: rs.Primary.ID}
		_, err := client.Get(key)
		if err != nil {
			if r, ok := err.(*egoscale.ErrorResponse); ok {
				if r.ErrorCode == egoscale.ParamError {
					return nil
				}
			}
			return err
		}
	}
	return fmt.Errorf("SSH key: still exists")
}

var testAccSSHKeyPairCreate = `
resource "exoscale_ssh_keypair" "key" {
  name       = "terraform-test-keypair"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDN7L45b4vO2ytH68ZUC5PMS1b7JG78zGslwcJ0zolE5BuxsCYor248/FKGC5TXrME+yBu/uLqaAkioq4Wp1PzP6Zy5jEowWQDOdeER7uu1GgZShcvly2Oaf/UKLqTdwL+U3tCknqHY63fOAi1lBwmNTUu1uZ24iNiogfhXwQn7HJLQK9vfoGwg+/qJIzeswR6XDa6qh0fuzdxWQ4JWHw2s8fv8WvGOlklmAg/uEi1kF5D6R7kJpOVaE20FLnT4sjA81iErhMIH77OaZqQKiyVH3i5m/lkQI/2e25ml8aculaWzHOs4ctd7l+K1ZWFYje3qMBY1sar1gd787eaqk6RZ"
}
`
