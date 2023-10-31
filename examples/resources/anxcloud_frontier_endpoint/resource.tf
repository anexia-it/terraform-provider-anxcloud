resource "anxcloud_frontier_endpoint" "foo" {
  name = "foo"
  path = "bar/baz"
  api = anxcloud_frontier_api.foo.id
}
