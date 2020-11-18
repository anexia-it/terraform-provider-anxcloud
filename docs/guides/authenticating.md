---
page_title: "Authenticating with Anexia Cloud"
---

Export `ANEXIA_TOKEN` environment variable to authenticate with Anexia Cloud.

```shell script
export ANEXIA_TOKEN=<token>
```

An alternative for the environment variable is to pass token directly in the main.tf file.

```hcl
provider "anxcloud" {
  token = "<token>"
}
```

The last authentication method is highly **NOT** recommended for production environments. 