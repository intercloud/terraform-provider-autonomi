resource "autonomi_transport" "transport" {
  name = "Transport name"
  workspace_id = autonomi_workspace.workspace.id
  product = {
    sku = "valid_sku"
  }
}