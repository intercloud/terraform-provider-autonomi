terraform {
  required_providers {
    autonomi = {
      source = "hashicorp.com/edu/autonomi"
    }
  }
}

provider "autonomi" {
  host = var.host
  terms_and_conditions = true
  personal_access_token = var.pat_token
}

resource "autonomi_workspace" "workspace" {
  name = "Workspace created with Terraform"
  description = "this is a description"
}

output "workspace_creation" {
  value = autonomi_workspace.workspace
}

resource "autonomi_node" "node" {
  name = "Node created with Terraform"
  account_id = autonomi_workspace.workspace.account_id
  workspace_id = autonomi_workspace.workspace.id
  type = "cloud"
  provider_config = {
    aws_account_id = var.aws_account_id
  }
  product = {
    sku = "CEQULD5100AWS"
  }
}

output "node_creation" {
  value = autonomi_node.node
}