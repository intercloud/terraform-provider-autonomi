data "autonomi_transport_product" "transport" {
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
    cheapest = true 
}