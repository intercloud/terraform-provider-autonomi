resource "autonomi_virtual_access_node" "virtual_access_node" {
  name = "Virtual Access Node created with Terraform"
  workspace_id = autonomi_workspace.workspace.id
  product = {
    sku = "valid_sku"
  }
}