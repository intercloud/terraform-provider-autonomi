resource "autonomi_cloud_node" "cloud_node" {
  name = "Node name"
  workspace_id = autonomi_workspace.workspace.id
  provider_config = {
    aws_account_id = "aws_account_id"
  }
  product = {
    sku = "valid_sku"
  }
}