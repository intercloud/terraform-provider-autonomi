data "autonomi_physical_ports" "ports" {
  filters = [
      {
        name    = "location"
        operator = "="
        values   = ["EQUINIX FR5"]
      },    {
        name    = "bandwidth"
        operator = "IN"
        values   = ["100", "500"]
      },
    ]
}