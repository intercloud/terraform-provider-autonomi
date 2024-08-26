variable "host_url" { // @TODO remove when it will be published
  description = "The hostname or base URL of the API endpoint for the Autonomi service. This URL is used by the custom Terraform provider to interact with the Autonomi API."
  type        = string
  sensitive   = true
}

variable "catalog_url" { // @TODO remove when it will be published
  description = "The hostname or base URL of the API endpoint for the Autonomi's catalog service. This URL is used by the custom Terraform provider to interact with the Autonomi's catalog API."
  type        = string
  sensitive   = true
}

variable "aws_account_id" {
  description = "The AWS Account ID associated with the resources to be managed. This ID is required to uniquely identify and manage resources within the specified AWS account."
  type        = string
  default = null
}

