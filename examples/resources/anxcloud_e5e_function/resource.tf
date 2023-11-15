resource "anxcloud_e5e_function" "example" {
  name        = "example function"
  application = anxcloud_e5e_application.example.id
  runtime     = "python_310"
  entrypoint  = "my.module::example_function"

  quota_timeout = 120

  # configure a s3 storage backend where the function code is located
  storage_backend_s3 {
    endpoint    = "https://s3.example.com"
    bucket_name = "example-bucket"
    object_path = "example/path"
    access_key  = "example-access-key"
    secret_key  = "example-secret-key"
  }

  # # alternative backend configurations:
  #
  # storage_backend_git {
  #   url      = "https://example.com/example.git"
  #   username = "foo"
  #   password = "bar"
  # }

  # storage_backend_archive {
  #   name    = "function.zip"
  #   content = "data:application/zip;base64,${filebase64("${path.module}/function.zip")}"
  # }

  # configure environment variables
  env {
    name  = "EXAMPLE_VARIABLE"
    value = "example"
  }

  # configure secret environment variables
  # note: changes to secret environment variables outside the terraform state
  # are not tracked
  env {
    name   = "EXAMPLE_SECRET"
    value  = "secret"
    secret = true
  }

  # configure hostnames
  hostname {
    hostname = "example.com"
    ip       = "198.51.100.1"
  }
}
