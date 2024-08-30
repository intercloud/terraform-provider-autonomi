resource "autonomi_access_node" "access_node" {
  name = "Node name"
  workspace_id = autonomi_workspace.workspace.id
  product = {
    sku = "valid_sku"
  }
  physical_port_id = var.physical_port_id
  vlan = var.access_vlan
}