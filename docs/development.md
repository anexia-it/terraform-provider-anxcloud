# Terraform Provider Anxcloud

## Build provider

Run the following command to build the provider

```shell
go build -o terraform-provider-anxcloud
```

or

```shell
make build
```

## Test sample configuration

First, build and install the provider.

```shell
make install
```

Then, run the following command to initialize the workspace and apply the sample configuration.

```shell
terraform init && terraform apply
```

## Run unit tests

Execute the following command to run unit tests:

```shell
make test
```

## Run integration tests

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
