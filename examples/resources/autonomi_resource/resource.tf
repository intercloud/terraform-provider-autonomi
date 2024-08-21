resource "autonomi_workspace" "workspace" {
  name = "Workspace created with Terraform"
  description = "this is a description"
}

resource "autonomi_cloud_node" "cloud_node" {
  name = "Node created with Terraform"
  workspace_id = autonomi_workspace.workspace.id
  provider_config = {
    aws_account_id = var.aws_account_id
  }
  product = {
    sku = data.autonomi_cloud_products.clouds.hits.0.sku
  }
}

resource "autonomi_transport" "transport_FR5_LD5" {
  name = "Transport created with Terraform"
  workspace_id = autonomi_workspace.workspace.id
  product = {
    sku = data.autonomi_transport_products.transports.hits.0.sku
  }
}

resource "autonomi_attachment" "attachment_node_transport_LD5"{
  workspace_id = autonomi_workspace.workspace_test.id
  node_id = autonomi_cloud_node.cloud_node_LD5.id
  transport_id = autonomi_transport.transport_FR5_LD5.id
}

resource "autonomi_access_node" "access_node_FR5" {
  name = "Access Node created with Terraform"
  workspace_id = autonomi_workspace.workspace_test.id
  product = {
    sku = data.autonomi_access_products.access.hits.0.sku
  }
  physical_port_id = var.physical_port_id
  vlan = var.access_vlan
}