provider "exoscale" {
  version          = "~> 0.10"
}

resource "exoscale_ssh_keypair" "key" {
  name = "mykey"
}

resource "exoscale_security_group" "sg" {
  name = "mysg"
}

resource "exoscale_security_group_rules" "sg_rules" {
  security_group_id = exoscale_security_group.sg.id
  ingress {
    protocol  = "TCP"
    ports     = ["22"]
    cidr_list = ["0.0.0.0/0", "::/0"]
  }
}

resource "exoscale_compute" "vm" {
  display_name = "myvm"
  template     = "Linux Ubuntu 18.04 LTS 64-bit"
  size         = "Medium"
  key_pair     = exoscale_ssh_keypair.key.name
  disk_size    = 10
  zone         = "at-vie-1"

  security_group_ids = [exoscale_security_group.sg.id]

  provisioner "remote-exec" {
    connection {
      host        = exoscale_compute.vm.ip_address
      private_key = exoscale_ssh_keypair.key.private_key
    }

    inline = [
      "uname -a",
    ]
  }
}

