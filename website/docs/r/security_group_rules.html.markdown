---
layout: "exoscale"
page_title: "Exoscale: exoscale_security_group_rules"
sidebar_current: "docs-exoscale-security-group_rules"
description: |-
  Manages a set of rules to a security group.
---

# exoscale_security_group_rules

A security group rules represents a set of `ingress` and/or `egress` rules
which has to be linked to a `exoscale_security_group` resource. Note: any
other rule created outside of Terraform for this security group won't be
managed.

## Example usage

```hcl
resource "exoscale_security_group_rules" "http" {
  security_group_id = "${exoscale_security_group.http.id}"

  ingress {
    protocol = "TCP"
    cidr_list = ["0.0.0.0/0", "::/0"]
    ports = ["80", "8000-8888"]
    user_security_group_list = ["default", "etcd"]
  }

  ingress {
    protocol = "ICMP"
    cidr_list = ["0.0.0.0/0"]
    icmp_type = 8
  }

  ingress {
    protocol = "ICMPv6"
    cidr_list = ["::/0"]
    icmp_type = 128
  }

  egress {
    // ...
  }
}
```

## Argument Reference
- `security_group` - (Required) Security Group name to add rules to

- `security_group_id` - (Required) Security Group ID to add rules to

- `egress` or `ingress` - set of rules for the incoming or outgoing traffic

    - `protocol` - (Required) the protocol, e.g. `TCP`, `UDP`, `ICMP`, `ICMPv6`, .., or `ALL`

    - `description` - human description

    - `ports` - a set of port ranges

    - `icmp_type` and `icmp_code` - for `ICMP`, `ICMPv6` traffic

    - `cidr_list` - source/destination of the traffic as an IP subnet

    - `user_security_group_list` - source/destination of the traffic identified by a security group

## Attributes Reference

- `security_group` - Name of the security group

- `security_group_id` - Identifier of the security group
