## 0.3.3
ENHANCEMENTS

* provider: Fix a bug where updating tags hangs until timeout

## 0.3.2
ENHANCEMENTS

* provider: Configure client logging and Add logging helper functions

## 0.3.1
ENHANCEMENTS

* resource/virtual_server - network IP changes require resource recreation ([#45](https://github.com/anexia-it/terraform-provider-anxcloud/pull/45))
* resource/virtual_server - fixed bug with disk sizing ([#46](https://github.com/anexia-it/terraform-provider-anxcloud/pull/46))
* resource/vlan - Allow `vm_provisioning` to be updated inplace ([#48](https://github.com/anexia-it/terraform-provider-anxcloud/pull/48))
* resource/virtual_server - andle incomplete network informations to avoid drift ([#51](https://github.com/anexia-it/terraform-provider-anxcloud/pull/51))

* Upgraded to golang-1.16 build pipeline ([#49](https://github.com/anexia-it/terraform-provider-anxcloud/pull/49))

DOCUMENTATION
* resource/virtual_server - updated example and `disk` attribute documentation ([#47](https://github.com/anexia-it/terraform-provider-anxcloud/pull/47))

## 0.3.0
FEATURES

* resource/anxcloud_virtual_server - support for configuring multiple disks when creating virtual server ([#40](https://github.com/anexia-it/terraform-provider-anxcloud/pull/40))
* documentation/anxcloud_virtual_server - Updated `disk` documentation ([43](https://github.com/anexia-it/terraform-provider-anxcloud/pull/43))

ENHANCEMENTS

* resource/anxcloud_virtual_server - fixed some bugs with the import logic ([40](https://github.com/anexia-it/terraform-provider-anxcloud/pull/40))

## 0.2.4

FEATURES

* resource/vlan, resource/ip_address resource/virtual_server - support for importing existing resources ([#36](https://github.com/anexia-it/terraform-provider-anxcloud/pull/36))
* resource/tag, resource/network_prefix - support for importing existing resources ([#35](https://github.com/anexia-it/terraform-provider-anxcloud/pull/35))
* **New Data Source** `anxcloud_cpu_performance_types, anxcloud_tags, anxcloud_vsphere_locations` ([#29](https://github.com/anexia-it/terraform-provider-anxcloud/pull/29))
* **New Data Source** `anxcloud_nic_type, anxcloud_vlan, anxcloud_ip_address` ([#28](https://github.com/anexia-it/terraform-provider-anxcloud/pull/28))

## 0.2.3

ENHANCEMENTS

* all resources - if resources not found then let terraform to reflect this in the status ([#27](https://github.com/anexia-it/terraform-provider-anxcloud/pull/27))

## 0.2.2

ENHANCEMENTS

* resource/anxlcoud_virtual_server - reserve an IP address before creating a VM ([#20](https://github.com/anexia-it/terraform-provider-anxcloud/pull/20))

## 0.2.1

FEATURES

* **New Data Source:** `anxcloud_template` ([#14](https://github.com/anexia-it/terraform-provider-anxcloud/pull/14))
* **New Resource:** `anxcloud_tag` ([#15](https://github.com/anexia-it/terraform-provider-anxcloud/pull/15))

## 0.2.0

FEATURES

* **New Resource:** `anxcloud_vlan` ([#6](https://github.com/anexia-it/terraform-provider-anxcloud/pull/6))
* **New Resource:** `anxcloud_network_prefix` ([#10](https://github.com/anexia-it/terraform-provider-anxcloud/pull/10))
* **New Resource:** `anxcloud_ip_address` ([#11](https://github.com/anexia-it/terraform-provider-anxcloud/pull/11))
* **New Data Source:** `anxcloud_disk_type` ([#12](https://github.com/anexia-it/terraform-provider-anxcloud/pull/12))

ENHANCEMENTS

* resource/anxlcoud_virtual_server - add update/scale method ([#7](https://github.com/anexia-it/terraform-provider-anxcloud/pull/7))
* resource/anxcloud_vlan - simplify resource deletion ([#9](https://github.com/anexia-it/terraform-provider-anxcloud/pull/9))

## 0.1.0

FEATURES

* **New Resource:** `anxcloud_virtual_server` ([#3](https://github.com/anexia-it/terraform-provider-anxcloud/pull/3))
