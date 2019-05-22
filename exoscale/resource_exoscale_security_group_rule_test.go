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

func TestAccSecurityGroupRule(t *testing.T) {
	sg := new(egoscale.SecurityGroup)
	cidr := new(egoscale.EgressRule)
	usg := new(egoscale.IngressRule)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccSecurityGroupRule1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists("exoscale_security_group.sg", sg),
					testAccCheckEgressRuleExists("exoscale_security_group_rule.cidr", sg, cidr),
					testAccCheckSecurityGroupRule(cidr),
					testAccCheckSecurityGroupRule((*egoscale.EgressRule)(usg)),
					testAccCheckSecurityGroupRuleAttributes(map[string]schema.SchemaValidateFunc{
						"security_group": ValidateString("terraform-test-security-group"),
						"protocol":       ValidateString("TCP"),
						"type":           ValidateString("EGRESS"),
						"cidr":           ValidateString("::/0"),
						"start_port":     ValidateString("2"),
						"end_port":       ValidateString("1024"),
					}),
				),
			},
			resource.TestStep{
				Config: testAccSecurityGroupRule2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists("exoscale_security_group.sg", sg),
					testAccCheckIngressRuleExists("exoscale_security_group_rule.usg", sg, usg),
					testAccCheckSecurityGroupRule(usg),
					testAccCheckSecurityGroupRule((*egoscale.EgressRule)(usg)),
					testAccCheckSecurityGroupRuleAttributes(map[string]schema.SchemaValidateFunc{
						"security_group":      ValidateString("terraform-test-security-group"),
						"protocol":            ValidateString("ICMPv6"),
						"type":                ValidateString("INGRESS"),
						"icmp_type":           ValidateString("128"),
						"icmp_code":           ValidateString("0"),
						"user_security_group": ValidateString("terraform-test-security-group"),
					}),
				),
			},
		},
	})
}

func testAccCheckEgressRuleExists(n string, sg *egoscale.SecurityGroup, rule *egoscale.EgressRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Security Group Rule ID is set")
		}

		if len(sg.EgressRule) == 0 {
			return fmt.Errorf("no egress rules found")
		}

		return Copy(rule, sg.EgressRule[0])
	}
}

func testAccCheckIngressRuleExists(n string, sg *egoscale.SecurityGroup, rule *egoscale.IngressRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Security Group Rule ID is set")
		}

		if len(sg.IngressRule) == 0 {
			return fmt.Errorf("no Ingress rules found")
		}

		return Copy(rule, sg.IngressRule[0])
	}
}

func testAccCheckSecurityGroupRule(v interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		switch v.(type) {
		case egoscale.IngressRule, egoscale.EgressRule:
			r, _ := v.(egoscale.IngressRule)
			if r.RuleID == nil {
				return fmt.Errorf("security group rule id is nil")
			}
		}

		return nil
	}
}

func testAccCheckSecurityGroupRuleAttributes(expected map[string]schema.SchemaValidateFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "exoscale_security_group_rule" {
				continue
			}

			return testResourceAttributes(expected, rs.Primary.Attributes)
		}

		return errors.New("security_group_rule resource not found in the state")
	}
}

func testAccCheckSecurityGroupRuleDestroy(s *terraform.State) error {
	client := GetComputeClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "exoscale_security_group_rule" {
			continue
		}

		sgID, err := egoscale.ParseUUID(rs.Primary.Attributes["security_group_id"])
		if err != nil {
			return err
		}

		sg := &egoscale.SecurityGroup{ID: sgID}
		_, err = client.Get(sg)
		if err != nil {
			if r, ok := err.(*egoscale.ErrorResponse); ok {
				if r.ErrorCode == egoscale.ParamError {
					return nil
				}
			}
			return err
		}
	}

	return fmt.Errorf("security group rule still exists")
}

var testAccSecurityGroupRule1 = `
resource "exoscale_security_group" "sg" {
  name = "terraform-test-security-group"
  description = "Terraform Security Group Test"
}

resource "exoscale_security_group_rule" "cidr" {
  security_group_id = "${exoscale_security_group.sg.id}"
  protocol = "TCP"
  type = "EGRESS"
  cidr = "::/0"
  start_port = 2
  end_port = 1024
}
`

var testAccSecurityGroupRule2 = `
resource "exoscale_security_group" "sg" {
  name = "terraform-test-security-group"
  description = "Terraform Security Group Test"
}

resource "exoscale_security_group_rule" "usg" {
  security_group = "${exoscale_security_group.sg.name}"
  protocol = "ICMPv6"
  type = "INGRESS"
  icmp_type = 128
  icmp_code = 0
  user_security_group = "${exoscale_security_group.sg.name}"
}
`
