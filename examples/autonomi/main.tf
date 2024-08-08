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
  csp_name = var.csp_name
  underlay_provider = var.underlay_provider
  location = var.location
  bandwidth = var.bandwidth
}

output "autonomi_cloud_products" {
  value = data.autonomi_cloud_products.clouds
}