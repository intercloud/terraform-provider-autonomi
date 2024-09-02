data "autonomi_transport_products" "transports" {
  filters = [
      {
        name    = "provider"
        operator = "="
        values   = ["EQUINIX"]
      },
      {
        name    = "locationTo"
        operator = "IN"
        values   = ["EQUINIX FR5", "EQUINIX LD5"]
      },
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