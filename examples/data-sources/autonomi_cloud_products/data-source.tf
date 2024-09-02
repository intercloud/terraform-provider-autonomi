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