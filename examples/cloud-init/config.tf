provider "template" {
  version = "~> 2.0"
}

provider "exoscale" {
  version = "~> 0.10.0"
  key     = var.key
  secret  = var.secret
}

