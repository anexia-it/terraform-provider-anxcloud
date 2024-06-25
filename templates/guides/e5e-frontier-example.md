---
page_title: "Create a serverless application using e5e and frontier"
---

## Create a serverless application using e5e and frontier

In this guide we will set up a serverless function with e5e and expose it via HTTP using frontier.

### Example source file

```python
# filename: hello_world.py
import os

def hello_world(event, context):
    return {
        'data': os.getenv("MESSAGE"),
    }
 ```

### Example terraform configuration

```terraform
# Set up an e5e application & function

resource "anxcloud_e5e_application" "example" {
  name = "example app"
}

data "archive_file" "hello_world" {
  type        = "zip"
  source_file = "${path.module}/hello_world.py"
  output_path = "${path.module}/function.zip"
}

data "local_file" "hello_world" {
  filename   = data.archive_file.hello_world.output_path
  depends_on = [data.archive_file.hello_world]
}

resource "anxcloud_e5e_function" "hello_world" {
  name        = "hello world"
  application = anxcloud_e5e_application.example.id
  runtime     = "python_310"
  entrypoint  = "hello_world::hello_world"

  # Changes to the `revision` attribute will trigger a new deployment
  revision = data.local_file.hello_world.content_sha1

  # Note: other storage backends are available
  # Check out the anxcloud_e5e_function docs for more information
  storage_backend_archive {
    name    = "function.zip"
    content = "data:application/zip;base64,${data.local_file.hello_world.content_base64}"
  }

  env {
    name  = "MESSAGE"
    value = "Hello from e5e!"
  }
}


# Expose the function via HTTP using frontier

resource "anxcloud_frontier_api" "example" {
  name              = "example api"
  transfer_protocol = "http"
}

resource "anxcloud_frontier_endpoint" "hello_world" {
  name = "hello world"
  path = "hello/world"
  api  = anxcloud_frontier_api.example.id
}

resource "anxcloud_frontier_action" "hello_world" {
  http_request_method = "get"
  endpoint            = anxcloud_frontier_endpoint.hello_world.id

  e5e_function {
    function = anxcloud_e5e_function.hello_world.id
  }
}

# Generate random revision id on every run
resource "random_id" "revision" {
  keepers     = { ts = timestamp() }
  byte_length = 10
}

resource "anxcloud_frontier_deployment" "v1" {
  slug = "v1"
  api  = anxcloud_frontier_api.example.id

  # Changes to the `revision` attribute will trigger a new deployment
  # TODO(user): value should be replaced with something more suitible in production (e.g. commit hash passed in via variable)
  revision = random_id.revision.hex

  depends_on = [
    # Ensure that all actions exist before creating a deployment
    anxcloud_frontier_action.hello_world
  ]

  lifecycle {
    # Ensure that we always have an active deployment
    create_before_destroy = true
  }
}

# Output the endpoint urls

output "urls" {
  value = {
    hello_world = "https://frontier.anexia-it.com/${anxcloud_frontier_api.example.id}/${anxcloud_frontier_deployment.v1.slug}/${anxcloud_frontier_endpoint.hello_world.path}"
  }
}
```
