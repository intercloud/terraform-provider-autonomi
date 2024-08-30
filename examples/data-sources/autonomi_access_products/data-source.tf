data "autonomi_access_products" "access" {
  filters = [
    {
      name     = "location"
      operator = "IN"
      values   = ["EQUINIX FR5", "EQUINIX LD5"]
    },
    {
      name    = "bandwidth"
      operator = "TO"
      values   = ["100", "500"]
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