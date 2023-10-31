resource "anxcloud_frontier_action" "foo" {
  http_request_method = "get"
  endpoint = anxcloud_frontier_endpoint.foo.id

  mock_response {
    language = "plaintext"
    body = "hello world!"
  }
}

resource "anxcloud_frontier_action" "bar" {
  http_request_method = "post"
  endpoint = anxcloud_frontier_endpoint.bar.id

  e5e_function {
    function = var.e5e_function_identifier
  }
}
