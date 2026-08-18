# Kubernetes provider manual E2E test

This configuration looks up two existing VLANs and two existing IPv4 prefixes through data sources, then creates one Kubernetes cluster and one node pool. Both networks are nested in the node-pool resource and included in its creation request. The configuration uses the provider installed by this repository's `make install` target rather than the public registry provider.

The portable service workload and LBaaS smoke test have moved to [`../anxcloud-kubernetes-service-e2e`](../anxcloud-kubernetes-service-e2e). That directory is self-contained and can be copied into its own repository.

## Initial apply

From the repository root:

```sh
make install
cd tests/e2e/kubernetes
```

Make sure the variable `test_id` is a unique cluster name, select an available `kubernetes_version`, and confirm the existing VLAN and prefix identifiers are available in `location_code`. Then export the Anexia provider token:

```sh
export ANEXIA_TOKEN="$anexia_token"
terraform init -reconfigure -upgrade
terraform validate
terraform plan -out=tfplan
TF_LOG=INFO terraform apply tfplan
terraform output
```

With `wait_until_ready = true`, provider logs report cluster state transitions and continue waiting through intermediate states, including `Error`, until state `0` is reached.

The Kubernetes API defaults to production. To exercise another service environment, pass the same E2E variable through the cluster and node-pool resources:

```sh
terraform apply -var-file=terraform.tfvars -var='kubernetes_api_environment=stage'
```

Supported values are `prod`, `stage`, and `dev`.

If this E2E directory was applied before the kubeconfig workload was split out,
remove its obsolete kubeconfig entry from local Terraform state once:

```sh
terraform state rm 'anxcloud_kubernetes_kubeconfig.cluster-admin'
```

This only forgets the old Terraform state entry; it does not delete or change
the cluster's kubeconfig in Anexia.

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

The VLANs and prefixes are data sources and will not be deleted. Terraform destroys the node pool and cluster in dependency order:

```sh
terraform destroy
```
