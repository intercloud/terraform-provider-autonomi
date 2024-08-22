data "autonomi_cloud_products" "clouds" {
  filters = [
      {
        name    = "cspName"
        operator = "="
        values   = ["AWS"]
      },
      {
        name     = "cspRegion"
        operator = "="
        values   = ["eu-central-1"]
      },
          {
        name    = "cspCity"
        operator = "="
        values   = ["Frankfurt"]
      },    {
        name    = "provider"
        operator = "="
        values   = ["EQUINIX"]
      },    {
        name    = "bandwidth"
        operator = "TO"
        values   = ["100", "500"]
      },
    ]
} 

data "autonomi_transport_products" "transports" {
  filters = [
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
}

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
} 