data "autonomi_virtual_access_products" "virtual_access_products" {
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
  sort = [
    {
      name     = "priceMrc"
      value   = "asc"
    },
    {
      name    = "bandwidth"
      value   = "desc"
    },
  ]  
}