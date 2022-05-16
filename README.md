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


## Releasing

1. Make your code changes on the branch of your fork of [terraform-provider-anxcloud](https://github.com/anexia-it/terraform-provider-anxcloud)
2. Create a pull request
3. Trigger the integration tests via a `/ok-to-test sha=$SHA` comment. The $SHA represents the last commit in the PR.
4. Merge the PR into the main branch of [terraform-provider-anxcloud](https://github.com/anexia-it/terraform-provider-anxcloud)
5. Create a tag on your fork, eg via `git tag v0.2.4`
6. Push the tag via `git push upstream --tags` 

=> The [release workflow](https://github.com/anexia-it/terraform-provider-anxcloud/blob/main/.github/workflows/release.yml) will create the release


## Guide for repository maintainers

### Run integration tests from fork repository

To run integration tests from fork repositories maintainer must carefully check changes that a PR is trying to make. **It is the maintainer's responsibility to avoid secrets leak.**

Integration tests are executed after adding a comment:

```bash
/ok-to-test sha=<short-commit>
```

After the `ok-to-test` job has started, move to the GitHub actions page to see the output from integration-tests jobs (click `ok-to-test` job).

*Note: the `integration-fork` job status is updated on the main PR page once it is finished.*


## License

[Apache License](LICENSE)
