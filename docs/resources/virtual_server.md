---
page_title: "virtual_server resource - terraform-provider-anxcloud"
subcategory: ""
description: |-
  The Virual Server resource allows you to create virtual machines at Anexia Cloud.
---

# Resource `anxcloud_virtual_server`

-> Visit the [Perform CRUD operations with Providers](https://learn.hashicorp.com/tutorials/terraform/provider-use?in=terraform/providers&utm_source=WEBSITE&utm_medium=WEB_IO&utm_offer=ARTICLE_PAGE&utm_content=DOCS) Learn tutorial for an interactive getting started experience.

The virtual_server resource allows you to configure and run Virtual Machine at Anexia Cloud.

## Example Usage

```hcl
resource "anxcloud_virtual_server" "example" {
  hostname      = "example-terraform"
  location_id   = "52b5f6b2fd3a4a7eaaedf1a7c019e9ea"
  template_id   = "12c28aa7-604d-47e9-83fb-5f1d1f1837b3"
  template_type = "templates"

  cpus     = 4
  memory   = 4096
  password = "flatcar#1234$%"

  # set two network interfaces
  # NIC 1
  network {
    vlan_id  = "ff70791b398e4ab29786dd34f211694c"
    nic_type = "vmxnet3"
  }

  # NIC 2
  network {
    vlan_id  = "ff70791b398e4ab29786dd34f211694c"
    nic_type = "vmxnet3"
  }

  disks {
    disk: 100
  }

  disks {
    disk: 200
  }

  dns = ["8.8.8.8"]
}
```

## Argument Reference

- `hostname` - (Required) Virtual server hostname.
- `location_id` - (Required) Location identifier.
- `template_id` - (Required) Template identifier. Can be obtained by template data source.
- `template_type` - (Required) Operating system template type. Can be `templates` or `from_scratch`.
- `cpus` - (Required) Amounts of CPUs. 
- `cpu_performance_type` - (Optional) CPU type. Example: `best-effort`, `standard`, `enterprise`, `performance`, defaults to `standard`.
- `sockets` - (Optional) Amount of CPU sockets Number of cores have to be a multiple of sockets, as they will be spread evenly across all sockets. Defaults to number of cores, i.e. one socket per CPU core.
- `memory` - (Required) Memory in MB.
- `network` - (Optional) Network interface. See [network](#network) below for details. 
- `dns` - (Optional) DNS configuration. Maximum items 4. Defaults to template settings.
- `password` (Required) Plaintext password. Example: ('!anx123mySuperStrongPassword123anx!', 'go3ju0la1ro3', …). USE IT AT YOUR OWN RISK! (or SSH key instead).
- `ssh_key` - (Required) Public key (instead of password, only for Linux systems). Recommended over providing a plaintext password.
- `script` - (Optional) Script to be executed after provisioning. Should be base64 encoded. Consider the corresponding shebang at the beginning of your script. If you want to use PowerShell, the first line should be: #ps1_sysnative.
- `boot_delay` - (Optional) Boot delay in seconds. Example: (0, 1, …).
- `enter_bios_setup` - (Optional) Start the VM into BIOS setup on next boot. Defaults to false.
- `force_restart_if_needed` - (Optional) Certain operations may only be performed in powered off stat. Such as: shrinking memory, shrinking/adding cpu, removing disk, scale a disk beyond 2 GB. Passing this value as true will always execute a power offand reboot request after completing all other operations. Without this flag set to true scaling operations requiring a reboot will fail. Defaults to false.
- `critical_operation_confirmed` - (Optional) Confirms a critical operation (if needed). Potentially dangerous operations (e.g. resulting in data loss) require an additional confirmation. The parameter is used for VM UPDATE requests. Defaults to false.
- `tags` - (Optional) List of tags names that should be attached to Virtual Server (those should be created over `anxcloud_tag` resource).

### Network

- `vlan_id` - (Required) VLAN identifier.
- `nic_type` - (Required) Network interface card type.
- `ips` - (Optional) Requested list of IPs and IPs identifiers. IPs are ignored when using template_type 'from_scratch'. Defaults to free IPs from IP pool attached to VLAN.

### Disks

- `disk` - (Required) Disk size in GB.
- `disk_type` - (Optional) Storage type for this disk. Default as per datacenter.

## Attributes Reference

In addition to all the arguments above, the following attributes are exported:

- `id` - Virtual Server identifier.
- `info` - Virtual Server details. See [info](#info) below for details.

### Info

- `identifier` - Virtual server identifier.
- `status` - Virtual server status.
- `name` - Virtual server name.
- `custom_name` - Virtual server custom name.
- `location_code` - Location code.
- `location_country` - Location country.
- `location_name` - Location name.
- `cpu` - Number of CPUs.
- `cores` - Number of CPU cores.
- `ram` - Memory in MB.
- `disks_number` - Number of the attached disks.
- `disks_info` - Disks info. See [disks info](#disks-info) below for details. 
- `network` - Network interfaces. See [network interfaces](#network-interfaces) below for details.
- `guest_os` - Guest operating system.
- `version_tools` - Version tools.
- `guest_tools_status` - Guest tools status.

### Disks info

- `disk_id` - Disk identifier.
- `disk_gb` - Size of the disk in GB.
- `disk_type` - Disk type.
- `iops` - Disk input/output operations per second.
- `latency` - Disk latency.
- `storage_type` - Disk storage type.
- `bus_type` - Disk device bus type.
- `bus_type_label` - Disk device bus type label.

### Network interfaces

- `id` - Network interface card identifier.
- `ip_v4` - List of IPv4 addresses attached to the interface.
- `ip_v6` - List of IPv6 addresses attached to the interface.
- `nic` - NIC type number.
- `vlan` - VLAN identifier..
- `mac_address` - MAC address of the NIC.
