---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "autonomi_cloud_products Data Source - autonomi"
subcategory: ""
description: |-
  
---

# autonomi_cloud_products (Data Source)

```terraform
data "autonomi_cloud_products" "clouds" {
  csp_name = "AWS"
  csp_region = "eu-west-1"
  csp_city = "London"
  underlay_provider = "EQUINIX"
  location = "EQUINIX LD5"
  bandwidth = 100
} 
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `bandwidth` (Number) Name of the Provider: expected values are [50, 100, 110, 500, 1000, 5000, 10000]
- `csp_city` (String) Name of the CSP city
- `csp_name` (String) Name of the CSP expected values are [AWS, Azure, GCP]
- `csp_region` (String) Name of the CSP region
- `location` (String) Name of the Location: expected values are [...]
- `underlay_provider` (String) Name of the Provider: expected values are [Equinix, Megaport]

### Read-Only

- `facet_distribution` (Attributes) The **facet_distribution** attribute provides an overview of the distribution of various facets within the cloud products returned by the Meilisearch query. This attribute allows you to analyze the frequency of different categories or attributes in the search results. (see [below for nested schema](#nestedatt--facet_distribution))
- `hits` (Attributes List) The **hits** attribute contains the list of cloud products returned by the Meilisearch query. Each hit represents a cloud product that matches the specified search criteria. (see [below for nested schema](#nestedatt--hits))

<a id="nestedatt--facet_distribution"></a>
### Nested Schema for `facet_distribution`

Read-Only:

- `bandwidth` (Map of Number)
- `csp_city` (Map of Number)
- `csp_name` (Map of Number)
- `csp_region` (Map of Number)
- `location` (Map of Number)
- `provider` (Map of Number)


<a id="nestedatt--hits"></a>
### Nested Schema for `hits`

Read-Only:

- `bandwidth` (Number)
- `cost_mrc` (Number)
- `cost_nrc` (Number)
- `csp_name` (String)
- `date` (String)
- `duration` (Number)
- `id` (Number)
- `location` (String)
- `price_mrc` (Number)
- `price_nrc` (Number)
- `provider` (String)
- `sku` (String)