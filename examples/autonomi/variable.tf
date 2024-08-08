variable "pat_token" {
  description = "The Personal Access Token (PAT) used to authenticate with the Autonomi API. This token can be obtained from the Autonomi service and is required to access and manage resources via the API."
  type        = string
  sensitive   = true
}

variable "aws_account_id" {
  description = "The AWS Account ID associated with the resources to be managed. This ID is required to uniquely identify and manage resources within the specified AWS account."
  type        = string
}

variable "host_url" { // @TODO remove when it will be published
  description = "The hostname or base URL of the API endpoint for the Autonomi service. This URL is used by the custom Terraform provider to interact with the Autonomi API."
  type        = string
  sensitive = true
}