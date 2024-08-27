# Terraform Provider Autonomi

Use the Autonomi provider to interact with the many resources supported by Autonomi. You must configure the provider with the proper credential before you can use it.

## Example Usage

```terraform
terraform {
  required_providers {
    autonomi = {
      source = "hashicorp.com/intercloud/autonomi"
    }
  }
}

provider "autonomi" {
  terms_and_conditions = true
}

resource "autonomi_workspace" "workspace_test" {
  name = "Workspace created with Terraform"
  description = "from autonomi with <3"
}
```

## Authentication and Configuration

Configuration for the Autonomi Provider can be derived from several sources, which are applied in the following order:

1. Parameters in the provider configuration
2. Environment variables

### Provider configuration

Access can be allowed by adding a personal access token to the autonomi provider block.
The [terms and conditions](https://docs.autonomi-platform.com/docs/legal) must be accepted to be able to deploy resources.

Usage:

```terraform
provider "autonomi" {
  terms_and_conditions = true
  personal_access_token = "my-personnal-access-token"
}
```

### Environment Variables

Access can be allowed by using the `AUTONOMI_PAT` environment variables. For a local usage the variables `AUTONOMI_HOST_URL` and `AUTONOMI_CATALOG_URL` must also be set.

For example:

```terraform
provider "autonomi" {
  terms_and_conditions = true
}
```

```bash
export AUTONOMI_PAT=<my-personal-access-token>
export AUTONOMI_HOST_URL=<autonomi-api-url>
export AUTONOMI_CATALOG_URL=<autonomi-catalog-url>
terraform plan
```
