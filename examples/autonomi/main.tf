terraform {
  required_providers {
    autonomi = {
      source = "hashicorp.com/edu/autonomi"
    }
  }
}

provider "autonomi" {
  host_url = var.host_url
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