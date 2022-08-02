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
* resource/anxcloud_virtual_server: add delay after `AwaitCompletion` to handle pending changes before read (#111, @marioreggiori)
* (internal) acceptance tests: make ProviderFactories real factories (#102, @marioreggiori)

### Changed
* resource/anxcloud_virtual_server: increase delete timeout (#112, @marioreggiori)
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
