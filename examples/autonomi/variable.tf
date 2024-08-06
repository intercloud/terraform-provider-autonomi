variable "pat_token" {
  description = "The Personal Access Token (PAT) used to authenticate with the Autonomi API. This token can be obtained from the Autonomi service and is required to access and manage resources via the API."
  type        = string
  sensitive   = true
}

variable "aws_account_id" {
  description = "The AWS Account ID associated with the resources to be managed. This ID is required to uniquely identify and manage resources within the specified AWS account."
  type        = string
}