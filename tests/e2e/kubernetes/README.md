# Kubernetes provider manual E2E test

This configuration looks up two existing VLANs and two existing IPv4 prefixes through data sources, then creates one Kubernetes cluster and one node pool. The internal network is included in the node-pool POST because the API requires at least one network during creation. A standalone node-pool network resource attaches the external network afterward. The configuration uses the provider installed by this repository's `make install` target rather than the public registry provider.

The Anexia token is read from `ANEXIA_TOKEN`; do not put it in a Terraform file.

## Initial apply

From the repository root:

```sh
make install
cd tests/e2e/kubernetes
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars` so `test_id` is unique, select an available `kubernetes_version`, and confirm the existing VLAN and prefix identifiers in `main.tf` are available in `location_code`. Then export the token and initialize Terraform against the local plugin directory:

```sh
export ANEXIA_TOKEN='replace-with-your-token'
terraform init -reconfigure -upgrade -plugin-dir="$HOME/.terraform.d/plugins"
terraform validate
terraform plan -out=tfplan
TF_LOG=INFO terraform apply tfplan
terraform output
```

With `wait_until_ready = true`, provider logs report cluster state transitions and continue waiting through intermediate states, including `Error`, until state `0` is reached.

## Exercise in-place updates

The following must produce in-place `~ update` actions, not `-/+ replace` actions:

```sh
terraform plan \
  -var-file=terraform.tfvars \
  -var='enable_lbaas=false' \
  -var='node_replicas=2' \
  -var='node_cpus=4' \
  -var='node_pool_network_bandwidth_limit=10000' \
  -out=tfplan.update

TF_LOG=INFO terraform apply tfplan.update
terraform apply -refresh-only
terraform output
```

Reapply the normal values from `terraform.tfvars` to exercise the reverse PATCH, including boolean and numeric values returning to their defaults:

```sh
terraform plan -var-file=terraform.tfvars -out=tfplan.restore
TF_LOG=INFO terraform apply tfplan.restore
```

If the cluster is already in an API reconciliation error state, change one of the patchable values and run another apply. Terraform should send the PATCH and keep the same cluster ID instead of replacing or removing the resource.

## Test asynchronous cluster creation

For a separate fresh run, set `wait_until_ready = false`. The cluster create operation should return after the API has stored the object, regardless of its current reconciliation state:

```sh
terraform apply -var-file=terraform.tfvars -var='wait_until_ready=false'
```

## Cleanup

Cluster deletion removes the API object immediately while backend cleanup continues. For a manual E2E run, remove resources in dependency order:

```sh
terraform destroy -target=anxcloud_kubernetes_node_pool_network.external
terraform destroy -target=anxcloud_kubernetes_node_pool.e2e
terraform destroy -target=anxcloud_kubernetes_cluster.e2e
```

The VLANs and prefixes are data sources and will not be deleted. After cluster backend cleanup has completed:

```sh
terraform destroy
```

Targeted destroy is intentionally used here only to control teardown order for this manual test.
