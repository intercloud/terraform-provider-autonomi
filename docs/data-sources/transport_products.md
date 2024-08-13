---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "autonomi_transport_products Data Source - autonomi"
subcategory: ""
description: |-
  
---

# autonomi_transport_products (Data Source)

```terraform
data "autonomi_cloud_products" "clouds" {
  underlay_provider = "EQUINIX"
  location = "EQUINIX AM2"
  location_to = "EQUINIX DC2"
  bandwidth = 100
} 
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `bandwidth` (Number) Name of the Provider: expected values are [50, 100, 110, 500, 1000, 5000, 10000]
- `location` (String) Name of the first Location: expected values are [...]
- `location_to` (String) Name of the second Location: expected values are [...]
- `underlay_provider` (String) Name of the Provider: expected values are [Equinix, Megaport]

### Read-Only

- `facet_distribution` (Attributes) The **facet_distribution** attribute provides an overview of the distribution of 
various facets within the transport products returned by the Meilisearch query. This attribute allows you to analyze
the frequency of different categories or attributes in the search results. (see [below for nested schema](#nestedatt--facet_distribution))
- `hits` (Attributes List) The **hits** attribute contains the list of transport products returned by the Meilisearch
query. Each hit represents a transport product that matches the specified search criteria.
(see [below for nested schema](#nestedatt--hits))

<a id="nestedatt--facet_distribution"></a>
### Nested Schema for `facet_distribution`

Read-Only:

- `bandwidth` (Map of Number)
- `location` (Map of Number)
- `location_to` (Map of Number)
- `provider` (Map of Number)


<a id="nestedatt--hits"></a>
### Nested Schema for `hits`

Read-Only:

- `bandwidth` (Number)
- `cost_mrc` (Number)
- `cost_nrc` (Number)
- `date` (String)
- `duration` (Number)
- `id` (Number)
- `location` (String)
- `location_to` (String)
- `location_to_underlay` (String)
- `location_underlay` (String)
- `price_mrc` (Number)
- `price_nrc` (Number)
- `provider` (String)
- `sku` (String)