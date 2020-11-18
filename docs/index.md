---
page_title: "Provider: AnxCloud"
subcategory: ""
description: |-
  Terraform provider for interacting with Anexia Cloud API.
---

# AnxCloud Provider

-> Visit the [Anexia official website](https://anexia.com/en/) to get more info about Anexia Cloud.

The AnxCloud provider is used to interact with Anexia Cloud API.

## Example Usage

Do not keep your authentication token in HCL for production environments, use Terraform environment variables.

```hcl
provider "anxcloud" {
  token = "example-token"
}
```

## Guides

Check the [authenticating guide](guides/authenticating.md) to get knowledge how to run authenticate terraform that interacts with Anexia Cloud.

## Contact

e-mail: opensource@anexia-it.com
