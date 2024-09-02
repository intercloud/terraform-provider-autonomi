data "autonomi_cloud_product" "single_cloud_product" {
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
        operator = "="
        values   = ["100"]
      },
    ]
  cheapest = true
}