# Kubernetes provider manual E2E test

This configuration looks up two existing VLANs and two existing IPv4 prefixes through data sources, then creates one Kubernetes cluster and one node pool. Both networks are nested in the node-pool resource and included in its creation request. It then installs Traefik from the Anexia generic Helm chart and exposes a small nginx application through an Ingress. The configuration uses the provider installed by this repository's `make install` target rather than the public registry provider.

The Anexia provider reads its token from `ANEXIA_TOKEN`. The same value is passed to the sensitive Terraform variable `anexia_token` and written to the `anexia/anexia-credentials` Kubernetes Secret under the `token` key. Do not put the token in a Terraform file. The token is necessarily present in Terraform state, so keep the local state protected and never commit it.

## Initial apply

From the repository root:

```sh
make install
cd tests/e2e/kubernetes
```

Make sure the variable `test_id` is a unique cluster name, select an available `kubernetes_version`, and confirm the existing VLAN and prefix identifiers are available in `location_code`. Then export the token for both the Anexia provider and the Kubernetes Secret. The CLI configuration loads the locally built Anxcloud provider while downloading the official Kubernetes and Helm providers normally:

```sh
export ANEXIA_TOKEN="$anexia_token"
export TF_VAR_anexia_token="$anexia_token"
terraform init -reconfigure -upgrade
terraform validate
terraform plan -out=tfplan
TF_LOG=INFO terraform apply tfplan
terraform output
```

With `wait_until_ready = true`, provider logs report cluster state transitions and continue waiting through intermediate states, including `Error`, until state `0` is reached.

The Kubernetes API defaults to production. To exercise another service environment, pass the same E2E variable through all three provider resources:

```sh
terraform apply -var-file=terraform.tfvars -var='kubernetes_api_environment=stage'
```

Supported values are `prod`, `stage`, and `dev`.

## Verify the workload and load balancer

The Helm release installs only Traefik from `oci://anx-cr.io/se/charts/ks-generic-helmchart`; every other component in that umbrella chart remains disabled. Terraform also creates the `anexia` namespace, the token Secret, and an nginx smoke application exposed by a hostless Ingress.

Inspect the assigned load-balancer address:

```sh
terraform output load_balancer_smoke_test
```

Open the reported IP address or hostname over HTTP. A successful request returns:

```text
Anexia Kubernetes LBaaS works
```

For Kubernetes-level diagnostics, write the sensitive kubeconfig to a temporary file and inspect the namespace:

```sh
terraform output -raw cluster-admin-config > /tmp/anxcloud-e2e-kubeconfig
KUBECONFIG=/tmp/anxcloud-e2e-kubeconfig kubectl -n anexia get pods,services,ingresses
```

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

The VLANs and prefixes are data sources and will not be deleted. Terraform destroys the workload, Helm release, node pool, and cluster in dependency order:

```sh
terraform destroy
```
