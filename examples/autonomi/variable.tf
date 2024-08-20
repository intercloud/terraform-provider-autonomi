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

variable "pat_token" {
  description = "The Personal Access Token (PAT) used to authenticate with the Autonomi API. This token can be obtained from the Autonomi service and is required to access and manage resources via the API."
  type        = string
  sensitive   = true
}

variable "aws_account_id" {
  description = "The AWS Account ID associated with the resources to be managed. This ID is required to uniquely identify and manage resources within the specified AWS account."
  type        = string
}

variable "csp_name" {
  description = "The Cloud Service Provider Name"
  type        = string
  default = null
}

variable "csp_city" {
  description = "The Cloud Service Provider City"
  type        = string
  default = null
}

variable "csp_region" {
  description = "The Cloud Service Provider Region"
  type        = string
  default = null
}

variable "underlay_provider" {
  description = "The Underlay Service Provider Name"
  type        = string
  default = null
}
variable "bandwidth" {
  description = "The bandwidth in Mbps"
  type        = number
  default = null
}

variable "location" {
  description = "The Location"
  type        = string
  default = null
}
variable "location_to" {
  description = "The Location"
  type        = string
  default = null
}