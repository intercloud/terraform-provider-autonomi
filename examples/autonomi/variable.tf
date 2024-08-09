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

  validation {
    condition     = contains(["AWS", "GCP", "AZURE"], var.csp_name)
    error_message = "The csp_name value must be 'AWS', 'GCP' or 'AZURE'."
  }
}

variable "csp_city" {
  description = "The Cloud Service Provider City"
  type        = string
}

variable "csp_region" {
  description = "The Cloud Service Provider Region"
  type        = string
}
variable "underlay_provider" {
  description = "The Underlay Service Provider Name"
  type        = string

  validation {
    condition     = contains(["EQUINIX", "MEGAPORT"], var.underlay_provider)
    error_message = "The underlay_provider value must be either 'EQUINIX' or 'MEGAPORT'."
  }
}

variable "bandwidth" {
  description = "The bandwidth in Mbps"
  type        = number

  validation {
    condition     = contains([50, 100, 200, 400, 500, 1000, 2000, 5000, 10000], var.bandwidth)
    error_message = "The bandwidth value must be one of 50, 100, 200, 400, 500, 1000, 2000, 5000, or 10000 Mbps."
  }
}

variable "location" {
  description = "The Location"
  type        = string

  validation {
    condition     = contains(["EQUINIX AM2", "EQUINIX DC2", "EQUINIX FR5", "EQUINIX HK2", "EQUINIX LD5", "EQUINIX PA3", "EQUINIX SG1", "EQUINIX SV5"], var.location)
    error_message = "The location value must be 'EQUINIX AM2', 'EQUINIX DC2', 'EQUINIX FR5', 'EQUINIX HK2', 'EQUINIX LD5', 'EQUINIX PA3', 'EQUINIX SG1', 'EQUINIX SV5'."
  }
}

variable "host_url" { // @TODO remove when it will be published
  description = "The hostname or base URL of the API endpoint for the Autonomi service. This URL is used by the custom Terraform provider to interact with the Autonomi API."
  type        = string
  sensitive = true
}

variable "catalog_url" { // @TODO remove when it will be published
  description = "The hostname or base URL of the API endpoint for the Autonomi's catalog service. This URL is used by the custom Terraform provider to interact with the Autonomi's catalog API."
  type        = string
  sensitive = true
}