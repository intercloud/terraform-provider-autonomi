---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "autonomi Provider"
subcategory: ""
description: |-
  Use the Autonomi provider to create and manage Autonomi resources using Autonomi REST API.
  Autonomi allows you to easily inter-connect your clouds and enterprise resources.
  You must configure the provider with the proper credentials before you can use it.
---

# autonomi Provider

Use the Autonomi provider to create and manage Autonomi resources using Autonomi REST API.
Autonomi allows you to easily inter-connect your clouds and enterprise resources.
You must configure the provider with the proper credentials before you can use it.

## Example Usage

```terraform
provider "autonomi" {
  terms_and_conditions = true
  personal_access_token = var.pat_token
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `terms_and_conditions` (Boolean) Terms and conditions

### Optional

- `personal_access_token` (String, Sensitive) Personal Access Token (PAT) to authenticate through Autonomi API. This token can be obtained from the Autonomi service and is required to access and manage resources via the API. Can be set as variable or in environment as AUTONOMI_PAT
