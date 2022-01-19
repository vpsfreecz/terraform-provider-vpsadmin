terraform {
  required_version = ">= 1.0"

  required_providers {
    vpsadmin = {
      source  = "terraform.vpsfree.cz/vpsfreecz/vpsadmin"
      version = ">= 0.2"
    }
  }
}
