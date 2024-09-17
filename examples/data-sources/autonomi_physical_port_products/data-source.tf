data "autonomi_physical_port_products" "physical_port_products" {
  filters = [
      {
        name    = "location"
        operator = "="
        values   = ["EQUINIX FR5"]
      },    {
        name    = "bandwidth"
        operator = "TO"
        values   = ["100", "500"]
      },{
        name    = "duration"
        operator = "="
        values   = ["12"]
      },
    ]
  sort = [
    {
      name     = "price_mrc"
      value   = "asc"
    },
    {
      name    = "bandwidth"
      value   = "desc"
    },
    {
      name    = "duration"
      value   = "desc"
    },
  ]
} 