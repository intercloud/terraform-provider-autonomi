data "autonomi_physical_port" "port" {
  most_recent = true,
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