<a href="https://terraform.io">
    <img src="https://raw.githubusercontent.com/hashicorp/terraform-website/master/public/img/logo-text.svg" alt="Terraform logo" title="Terraform" align="right" height="50" />
</a>

# Terraform Provider for Anexia Cloud

- Website: [terraform.io](https://terraform.io)
- Tutorials: [learn.hashicorp.com](https://learn.hashicorp.com/terraform?track=getting-started#getting-started)
- Forum: [discuss.hashicorp.com](https://discuss.hashicorp.com/c/terraform-providers/tf-anxcloud/)
- Chat: [gitter](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing List: [Google Groups](http://groups.google.com/group/terraform-tool)
- Contact: [opensource@anexia-it.com](opensource@anexia-it.com)

This provider is maintained internally by the Anexia Cloud team.

## Documentation

Full documentation is available under [docs/](docs/index.md)

## Development

### Build provider

Run the following command to build the provider

```shell
go build -o terraform-provider-anxcloud
```

or

```shell
make build
```

### Test sample configuration

First, build and install the provider.

```shell
make install
```

Then, run the following command to initialize the workspace and apply the sample configuration.

```shell
terraform init && terraform apply
```

### Run unit tests

Execute the following command to run unit tests:

```shell
make test
```

### Run integration tests

Export `ANEXIA_TOKEN` by executing:

```shell
export ANEXIA_TOKEN='<token>'
```

and run all integration tests by executing the following command:

```shell
make testacc
```

or run specific test case:

```
make testacc TESTARGS='-run=TestAccXXX'
```

## Contributing

To contribute, please read the contribution guidelines: [Contributing to Terraform - Anexia Cloud Provider](CONTRIBUTING.md)

1. Make your code changes on the branch of your fork of [terraform-provider-anxcloud](https://github.com/anexia-it/terraform-provider-anxcloud)
2. Create a pull request
3. Add your changes to the **Unreleased** section in [CHANGELOG.md](CHANGELOG.md)

Additional steps carried out by maintainers:

4. Review pull request and request changes if necessary
5. Approve workflow run via GitHub
6. Merge the PR into the main branch of [terraform-provider-anxcloud](https://github.com/anexia-it/terraform-provider-anxcloud)

## Releasing

1. Create pull request with entries from **Unreleased** section moved into a newly created release section in [CHANGELOG.md](CHANGELOG.md)
2. Draft GitHub release with new changes in the description and configured to create a tag with the new version number on publish
3. Merge previously created pull request into main
4. Publish prepared release
5. That's it! `go-releaser` will do the rest. Terraform registry will be automatically notified after binaries have been built via webhook.

=> The [release workflow](https://github.com/anexia-it/terraform-provider-anxcloud/blob/main/.github/workflows/release.yml) will create the release


## Guide for repository maintainers

### Run integration tests from fork repository

To run integration tests from fork repositories maintainer must carefully check changes that a PR is trying to make. **It is the maintainer's responsibility to avoid secrets leak.**

Integration tests are executed as part of the workflow, which has to be approved for outside collaborators.


## License

[Apache License](LICENSE)
