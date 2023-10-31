resource "anxcloud_frontier_deployment" "foo" {
  slug = "foo"
  api = anxcloud_frontier_api.foo.id

  depends_on = [
    # make sure that all actions of the API have been created
    anxcloud_frontier_action.foo
  ]

  # optional: handle automated redeployment (e.g. in CI)
  revision = var.commit_sha

  lifecycle {
    create_before_destroy = true
  }
}
