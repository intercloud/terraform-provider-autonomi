provider "autonomi" {
  host_url = var.host_url
  terms_and_conditions = true
  personal_access_token = var.pat_token
  catalog_url = var.catalog_url
}