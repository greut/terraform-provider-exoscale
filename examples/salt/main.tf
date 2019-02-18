provider "template" {}

provider "exoscale" {
  version = "~> 0.9.42"
  key = "${var.key}"
  secret = "${var.secret}"
}

resource "exoscale_security_group" "masters" {
  name = "salt"
}

// https://docs.saltstack.com/en/latest/topics/tutorials/firewall.html
resource "exoscale_security_group_rules" "ssh" {
  security_group = "${exoscale_security_group.masters.name}"

  ingress {
    description = "SSH rules"
    protocol = "TCP"
    ports = ["22"]
    cidr_list = ["0.0.0.0/0", "::/0"]
  }

  ingress {
    description = "salt-api"
    protocol = "TCP"
    ports = ["8443"]
    cidr_list = ["0.0.0.0/0", "::/0"]
  }

  ingress {
    description = "0MQ rules"
    protocol = "TCP"
    ports = ["4505-4506"]
    user_security_group_list = [
      "${exoscale_security_group.masters.name}",
      "${exoscale_security_group.minions.name}",
    ]
  }
}

resource "exoscale_security_group" "minions" {
  name = "minions"
}

resource "exoscale_security_group_rules" "zeromq" {
  security_group = "${exoscale_security_group.minions.name}"

  ingress {
    description = "0MQ rules"
    protocol = "TCP"
    ports = ["4505-4506"]
    user_security_group_list = ["${exoscale_security_group.masters.name}"]
  }
}

data "template_file" "master" {
  template = "${file("master.yaml")}"

  vars {
    hostname = "${var.master}"
  }
}

data "template_cloudinit_config" "master" {
  part  {
    filename = "init.cfg"
    content_type = "text/cloud-config"
    content = "${data.template_file.master.rendered}"
  }
}

resource "exoscale_compute" "master" {
  display_name = "${var.master}"
  template = "${var.template}"
  zone = "${var.zone}"
  size = "medium"
  disk_size = "50"
  key_pair = "${var.key_pair}"
  ip6 = true

  security_groups =  ["default", "${exoscale_security_group.masters.name}"]

  user_data = "${data.template_cloudinit_config.master.rendered}"

  tags {
    managedby = "terraform"
    role = "master"
  }
}

output "master_ip" {
  value = "${join(",", formatlist("%s@%s", exoscale_compute.master.*.username, exoscale_compute.master.*.ip_address))}"
}

data "template_file" "minion" {
  template = "${file("minion.yaml")}"
  count = "${length(var.minions)}"

  vars {
    hostname = "${element(var.minions, count.index)}"
    master_ip = "${exoscale_compute.master.ip_address}"
  }
}

data "template_cloudinit_config" "minion" {
  count = "${length(var.minions)}"

  part  {
    filename = "init.cfg"
    content_type = "text/cloud-config"
    content = "${element(data.template_file.minion.*.rendered, count.index)}"
  }
}

resource "exoscale_compute" "minion" {
  count = "${length(var.minions)}"
  display_name = "${element(var.minions, count.index)}"
  template = "${var.template}"
  zone = "${var.zone}"
  size = "medium"
  disk_size = "50"
  key_pair = "${var.key_pair}"
  ip6 = true

  security_groups =  ["default", "${exoscale_security_group.minions.name}"]

  user_data = "${element(data.template_cloudinit_config.minion.*.rendered, count.index)}"

  tags {
    managedby = "terraform"
    role = "minion"
  }
}
