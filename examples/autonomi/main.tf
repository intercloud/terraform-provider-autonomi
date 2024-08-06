terraform {
  required_providers {
    autonomi = {
      source = "hashicorp.com/intercloud/autonomi"
    }
  }
}

provider "autonomi" {
  terms_and_conditions = true
  personal_access_token = var.pat_token
}


data "autonomi_cloud_products" "clouds" {
  csp_name = "AWS"
  underlay_provider = "EQUINIX"
  location = "EQUINIX LD5"
  bandwidth = "100"
}

output "autonomi_cloud_products" {
  value = data.autonomi_cloud_products.clouds
}