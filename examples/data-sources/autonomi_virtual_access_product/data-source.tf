data "autonomi_virtual_access_product" "virtual_access_product" {
  cheapest = true
  filters = [
    {
      name     = "location"
      operator = "="
      values   = ["EQUINIX FR5"]
    },
    {
      name    = "bandwidth"
      operator = "TO"
      values   = ["100", "500"]
    },
    {
      name    = "provider"
      operator = "="
      values   = ["MEGAPORT"]
    },
  ]
}