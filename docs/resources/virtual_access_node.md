---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "autonomi_virtual_access_node Resource - autonomi"
subcategory: ""
description: |-
  Manages a virtual access node resource.
  Virtual access node resource allows you to create, modify and delete Autonomi virtual access nodes.
  Autonomi virtual access node allows you to easily connect to your datacenters assets via a virtual connection through Megaport / Equinix connections (virtual access nodes).
---

# autonomi_virtual_access_node (Resource)

Manages a virtual access node resource.
Virtual access node resource allows you to create, modify and delete Autonomi virtual access nodes.
Autonomi virtual access node allows you to easily connect to your datacenters assets via a virtual connection through Megaport / Equinix connections (virtual access nodes).

## Example Usage

```terraform
resource "autonomi_virtual_access_node" "virtual_access_node" {
  name = "Virtual Access Node created with Terraform"
  workspace_id = autonomi_workspace.workspace.id
  product = {
    sku = "valid_sku"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the access node
- `product` (Attributes) (see [below for nested schema](#nestedatt--product))
- `workspace_id` (String) ID of the workspace to which the access node belongs.

### Read-Only

- `administrative_state` (String) Administrative state of the access node [creation_pending, creation_proceed, creation_error,
deployed, delete_pending, delete_proceed, delete_error]
- `created_at` (String) Creation date of the access node
- `deployed_at` (String) Deployment date of the access node
- `id` (String) ID of the access node, set after creation
- `service_key` (Attributes) Access node's service key (see [below for nested schema](#nestedatt--service_key))
- `type` (String) Type of the node [access]
- `updated_at` (String) Update date of the access node
- `vlan` (Number) Vlan of the access node

<a id="nestedatt--product"></a>
### Nested Schema for `product`

Required:

- `sku` (String) ID of the product


<a id="nestedatt--service_key"></a>
### Nested Schema for `service_key`

Read-Only:

- `expiration_date` (String) expiration date of the service key
- `id` (String) ID of the service key
- `name` (String) name of the service key