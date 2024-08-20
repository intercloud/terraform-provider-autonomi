data "autonomi_cloud_products" "clouds" {
  csp_name = var.csp_name
  csp_region = var.csp_region
  csp_city = var.csp_city
  underlay_provider = var.underlay_provider
  location = var.location_to
  bandwidth = var.bandwidth
} 

data "autonomi_transport_products" "transports" {
  underlay_provider = var.underlay_provider
  location = var.location
  location_to = "EQUINIX LD5"
  bandwidth = var.bandwidth
}