# How to release `terraform-provider-anxcloud`

1. Make your code changes on the branch of your fork of [terraform-provider-anxcloud](https://github.com/anexia-it/terraform-provider-anxcloud)
1. Create a pull request
1. Trigger the integration tests via a `/ok-to-test sha=$SHA` comment. The $SHA represents the last commit in the PR.
1. Merge the PR into the main branch of [terraform-provider-anxcloud](https://github.com/anexia-it/terraform-provider-anxcloud)
1. Create a tag on your fork, eg via `git tag v0.2.4`
1. Push the tag via `git push upstream --tags` 

=> The [release workflow](https://github.com/anexia-it/terraform-provider-anxcloud/blob/main/.github/workflows/release.yml) will create the release
