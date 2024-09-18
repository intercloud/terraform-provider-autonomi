---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "autonomi_virtual_access_product Data Source - autonomi"
subcategory: ""
description: |-
  Datasource to retrieve a single access node product by filters.
  If zero, or more than one, product are retrieved with the filters, this datasource raises an error.
---

# autonomi_virtual_access_product (Data Source)

Datasource to retrieve a single access node product by filters.
If zero, or more than one, product are retrieved with the filters, this datasource raises an error.

## Example Usage

```terraform
data "autonomi_virtual_access_product" "virtual_access_product" {
  cheapest = true
  filters = [
    {
      name     = "location"
      operator = "="
      values   = ["EQUINIX FR5"]
    },
    {
      name    = "bandwidth"
      operator = "TO"
      values   = ["100", "500"]
    },
    {
      name    = "provider"
      operator = "="
      values   = ["MEGAPORT"]
    },
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `cheapest` (Boolean) To ensure only one hit is returned we advise to set at true
- `filters` (Attributes List) List of filters: [location, bandwidth]. (see [below for nested schema](#nestedatt--filters))

### Read-Only

- `facet_distribution` (Attributes) The **facet_distribution** attribute provides an overview of the distribution of various facets
within the access products returned by the Meilisearch query. This attribute allows you to analyze the frequency of
different categories or attributes in the search results. (see [below for nested schema](#nestedatt--facet_distribution))
- `hit` (Attributes) The **hit** attribute contains the access products returned by the Meilisearch query.
Each hit represents an access product that matches the specified search criteria.
If no hit is returned, an error will be returned. (see [below for nested schema](#nestedatt--hit))

<a id="nestedatt--filters"></a>
### Nested Schema for `filters`

Optional:

- `name` (String) Name of the filter among [location, bandwidth]
- `operator` (String) Comparison operator
- `values` (List of String) Values of the filter

<a id="nestedatt--facet_distribution"></a>
### Nested Schema for `facet_distribution`

Read-Only:

- `bandwidth` (Map of Number)
- `location` (Map of Number)
- `provider` (Map of Number)
- `type` (Map of Number)

<a id="nestedatt--hit"></a>
### Nested Schema for `hit`

Read-Only:

- `bandwidth` (Number)
- `cost_mrc` (Number)
- `cost_nrc` (Number)
- `date` (String)
- `duration` (Number)
- `id` (Number)
- `location` (String)
- `price_mrc` (Number)
- `price_nrc` (Number)
- `provider` (String)
- `sku` (String)
- `type` (String)