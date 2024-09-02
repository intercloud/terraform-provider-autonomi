data "autonomi_access_product" "one_product" {
  cheapest = true
  filters = [
    {
      name     = "location"
      operator = "="
      values   = ["EQUINIX FR5"]
    },
    {
      name    = "bandwidth"
      operator = "="
      values   = ["100"]
    },
  ]
}