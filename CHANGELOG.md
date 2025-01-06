# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

<!--
Please add your changelog entry under this comment in the correct category (Security, Fixed, Added, Changed, Deprecated, Removed - in this order).

Changelog entries are best in the following format, where scope is something like "provider", "data-source/anxcloud_dns_zones" or "resource/anxcloud_ip_address".
If the change isn't user-facing but still relevant enough for a changelog entry, add
"(internal)" before the scope.

* (internal)? scope: short description (#pr, @author)
-->

### Fixed

* golang/x/net: update to 0.33.0 due to CVE-2024-45338 (#236, @drpsychick)

## [0.6.6] - 2024-12-17

### Added
* resource/anxcloud_kubernetes_cluster: add `apiserver_allowlist` attribute to configure CIDRs allowed to access the apiserver (#228 @drpsychick)

## [0.6.5] - 2024-11-12

### Fixed
* resource/anxcloud_virtual_server: Handle empty template_id from API and make integration-test pass (#208, @drpsychick)
* resource/anxcloud_virtual_server: Handle missing VM info more gracefully to prevent division by zero panic (#170, @anx-mschaefer)

### Added
* resource/anxcloud_virtual_server: add `bandwidth_limit` attribute to networks (#206 @89q12)

### Notes

This release is based on a commit that was apparently rebased before getting merged to main.

## [0.6.4] - 2024-06-27

The v0.6.3 release wasn't published because there was an issue in our release workflow.
Check [0.6.3](#063---2024-06-25) to see what has changed since the last published release.

### Fixed
* (internal) Update GoReleaser to fix release workflow (#168, @nachtjasmin)

## [0.6.3] - 2024-06-25

## Added
* `anxcloud_vlan` data source (#165, @nachtjasmin)

### Changed
* resource/anxcloud_ip_address: reserve an available address based on filters (#163, @anx-mschaefer)

## [0.6.2] - 2024-05-28

### Added
* anxcloud_kubernetes_cluster: add `enable_autoscaling` to enable/disable autoscaling (#160, @nachtjasmin)

### Changed
* resource/anxcloud_kubernetes_cluster: increase create timeout (#161, @anx-mschaefer)

## [0.6.1] - 2024-03-21

### Fixed
* (internal) go releaser configuration

## [0.6.0] - 2024-03-21

### Breaking
* terraform cli 1.0 or later required from now on (#153, @anx-mschaefer)

### Changed
* (internal) resource/anxcloud_virtual_server: optimize creation of vms with multiple disks (#147, @anx-mschaefer)
* provider server: updated to protocol version 6 (#153, @anx-mschaefer)

### Added
* resources to manage the e5e service: (#142, @anx-mschaefer)
  * `anxcloud_e5e_application`
  * `anxcloud_e5e_function`
* resources to manage the frontier service: (#139, @anx-mschaefer)
  * `anxcloud_frontier_api`
  * `anxcloud_frontier_endpoint`
  * `anxcloud_frontier_action`
  * `anxcloud_frontier_deployment`

## [0.5.5] - 2024-01-12

### Fixed
* resource/anxcloud_virtual_server now ignores changes in the order of IP addresses returned by the engine (within a network) (#149, @anx-mschaefer)

## [0.5.4] - 2023-12-18

### Changed
* resource/anxcloud_dns_record:
  - create and delete operations are now handled in batches internally
  - attribute changes now trigger a replacement of the resource

## [0.5.3] - 2023-10-30

### Fixed
* resource/anxcloud_virtual_server now correctly calculates sockets on read (#136, @anx-mschaefer)

## [0.5.2] - 2023-02-20

### Added
* anxcloud_kubernetes_cluster: optional fields for using existing prefixes to deploy the cluster into (#123, @marioreggiori)

## [0.5.1] - 2023-01-19

### Changed
* resource tagging:
  * changing only the tags of a resource no longer causes a noop update call of the resource itself (#121, @marioreggiori)
  * (internal) change type of `tags` field from list to set and remove obsolete code (#121, @marioreggiori)

## [0.5.0] - 2022-11-30

### Added
* resource/anxcloud_kubernetes_{cluster,node_pool,kubeconfig} implemented to handle Kubernetes clusters, node pools and kubeconfigs (#118, @marioreggiori)
* data-source/anxcloud_kubernetes_cluster implemented to retrieve clusters by name (#118, @marioreggiori)

## [0.4.2] - 2022-10-12

### Fixed
* taggable resources: skip reading tags of manually deleted resources on read to prevent error (#117, @marioreggiori)

### Changed
* (internal) dependency: upgrade `go-anxcloud` to v0.4.5 (#116, @marioreggiori)
* data-source/anxcloud_core_location (#116, @marioreggiori)
  - optionally retrieve location by identifier
  - (internal) optimize retrieve by code

## [0.4.1] - 2022-08-09

### Fixed
* resource/anxcloud_virtual_server: add delay after `AwaitCompletion` to handle pending changes before read (#111, @marioreggiori)
* (internal) acceptance tests: make ProviderFactories real factories (#102, @marioreggiori)
* resource/anxcloud_virtual_server: `from_scratch` template provisioning ability restored (#114, @marioreggiori)

### Added
* resource/anxcloud_lbaas_loadbalancer: add a first LBaaS resource (#107, @marioreggiori)

### Changed
* resource/anxcloud_virtual_server: increase delete timeout (#112 & #113, @marioreggiori)
* (internal) acceptance tests: configured to run parallel (#102, @marioreggiori)

## [0.4.0] - 2022-07-07

### Breaking
* manual pagination removed from data sources (#90, @marioreggiori)
  - anxcloud_core_locations
  - anxcloud_ip_addresses
  - anxcloud_tags
  - anxcloud_vlans
* data-source/anxcloud_vsphere_locations: removed (previously marked deprecated; you can just use `anxcloud_core_location` everywhere this was needed before) (#104, @marioreggiori)

### Fixed
* resource/anxcloud_virtual_server: tags changed outside of terraform will now get reverted back to terraform config on apply (#101, @marioreggiori)
* docs: implicit `id` fields of data sources are now rendered as read-only instead of optional (#103, @marioreggiori)

### Added
* tagging capabilities to supported resources (#101, @marioreggiori)
  - anxcloud_ip_address
  - anxcloud_network_prefix
  - anxcloud_vlan

## [0.3.5] - 2022-06-13

### Fixed
* resource/anxcloud_virtual_server: removed mentioning of base64 encoding for bootstrap script (#89, @marioreggiori)
* resource/anxcloud_virtual_server: `cpu_performance_type` updates on read (#99, @marioreggiori)

### Added
* data-source/anxcloud_ip_address: allows users to retreive IP objects by id or address (#91, @marioreggiori)
* resource/anxcloud_virtual_server: support named vServer templates (#95, @marioreggiori)

### Changed
* resource/anxcloud_virtual_server: bootstrap script example added (#89, @marioreggiori)
* (internal) tools: upgrade golangci-lint to v1.46.2 to support go1.18 (#93, @marioreggiori)


## [0.3.4] - 2022-05-16
### Fixed
* resource/vlan: fix VLAN update leads to `vm_provisioning` flakiness (#71, @kstiehl)
* resource/vlan: await desired `vm_provisioning` state on create (#86, @marioreggiori)
* docs: fixed naming in development docs (#65, @HaveFun83)

### Added
* clouddns support
  - data-source/anxcloud_dns_records (#69, @X4mp)
  - data-source/anxcloud_dns_zones (#70, @X4mp)
  - resource/{anxcloud_dns_zone, anxcloud_dns_record} (#82, @marioreggiori)
  - data-source/anxcloud_core_location (#84, @marioreggiori)

### Changed
* provider: Upgrade Terraform plugin SDK (#87, @marioreggiori)
* docs: mostly now generated automatically and easier to keep up to date (#83, @marioreggiori)

## [0.3.3] - 2021-10-08
### Fixed
* provider: Fix a bug where updating tags hangs until timeout (#59, @kstiehl)
* resource/vlan: Fix a bug where permission issue lead to a crash (#61, @kstiehl)

### Changed
* provider: Add user agent to go client and cross compile for darwin/arm64 (#62, @kstiehl)
* resource/virtual_server: use deprovision progress instead of polling vmware API (#64, @kstiehl)

## [0.3.2] - 2021-06-17
### Changed
* (internal) provider: Configure client logging and Add logging helper functions (#54, @X4mp)

## [0.3.1] - 2021-06-02
### Fixed
* resource/virtual_server: fixed bug with disk sizing (#46, @X4mp)

### Changed
* resource/virtual_server: network IP changes require resource recreation (#45, @X4mp)
* resource/vlan: Allow `vm_provisioning` to be updated inplace (#48, @X4mp)
* resource/virtual_server: handle incomplete network informations to avoid drift (#51, @X4mp)
* (internal) build pipeline: Upgraded to golang-1.16 build pipeline (#49, @X4mp)
* docs/resource/virtual_server: updated example and `disk` attribute documentation (#47, @X4mp)

## [0.3.0] - 2021-03-26
### Fixed
* resource/anxcloud_virtual_server: fixed some bugs with the import logic (#40, @X4mp)

### Added
* resource/anxcloud_virtual_server: support for configuring multiple disks when creating virtual server (#40, @X4mp)

### Changed
* documentation/anxcloud_virtual_server: Updated `disk` documentation (#43, @X4mp)


## [0.2.4] - 2021-02-03
### Added
* data-source/{anxcloud_cpu_performance_types, anxcloud_tags, anxcloud_vsphere_locations} (#29, @stroebitzer)
* data-source/{anxcloud_nic_type, anxcloud_vlan, anxcloud_ip_address} (#28, @mfranczy)

### Changed
* resource/vlan, resource/ip_address resource/virtual_server: support for importing existing resources (#36, @mfranczy)
* resource/tag, resource/network_prefix: support for importing existing resources (#35, stroebitzer)


## [0.2.3] - 2021-01-13
### Changed
* resources (all): if resources not found then let terraform to reflect this in the status (#27, @mfranczy)

## [0.2.2] - 2020-12-17
### Changed
* resource/anxlcoud_virtual_server: reserve an IP address before creating a VM (#20, @mfranczy)

## [0.2.1] - 2020-12-14
### Added
* data-source/anxcloud_template (#14, @mfranczy)
* data-source/anxcloud_tag (#15, @mfranczy)

## [0.2.0] - 2020-12-07
### Added
* resource/anxcloud_vlan (#6, @mfranczy)
* resource/anxcloud_network_prefix (#10, @mfranczy)
* resource/anxcloud_ip_address (#11, @mfranczy)
* data-source/anxcloud_disk_type (#12, @mfranczy)

### Changed
* resource/anxlcoud_virtual_server: add update/scale method (#7, @mfranczy)
* resource/anxcloud_vlan: simplify resource deletion (#9, @mfranczy)

## [0.1.0] - 2020-11-22
### Added
* data-source/anxcloud_virtual_server (#3, @mfranczy)
