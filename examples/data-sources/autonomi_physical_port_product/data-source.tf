data "autonomi_physical_port_product" "single_physical_port_product" {
  filters = [
      {
        name    = "location"
        operator = "="
        values   = ["EQUINIX FR5"]
      },    {
        name    = "bandwidth"
        operator = "="
        values   = ["100"]
      },    {
        name    = "duration"
        operator = "="
        values   = ["12"]
      },
    ]
  cheapest = true
}