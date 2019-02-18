variable "key" {}
variable "secret" {}
variable "key_pair" {}


// hostnames are used as a source

variable "master" {
  default = "salt-master-001"
}

variable "minions" {
  type = "list"
  default = [
    "salt-minion-001",
    "salt-minion-002",
  ]
}

variable "zone" {
  default = "de-fra-1"
}

variable "template" {
  default = "Linux Ubuntu 18.04 LTS 64-bit"
}
