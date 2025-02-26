---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "anxcloud_virtual_server Resource - terraform-provider-anxcloud"
subcategory: ""
description: |-
  The virtual_server resource allows you to configure and run virtual machines.
---

# anxcloud_virtual_server (Resource)

The virtual_server resource allows you to configure and run virtual machines.

### Known limitations
- removal of disks not supported
- removal of networks not supported
- changing the speed on a network interface forces a replacement of the VM

## Example Usage

```terraform
data "anxcloud_core_location" "anx04" {
  code = "ANX04"
}

resource "anxcloud_vlan" "example" {
  location_id          = data.anxcloud_core_location.anx04.id
  vm_provisioning      = true
  description_customer = "example-terraform"
}

resource "anxcloud_network_prefix" "v4" {
  vlan_id              = anxcloud_vlan.example.id
  location_id          = data.anxcloud_core_location.anx04.id
  ip_version           = 4
  netmask              = 30
  description_customer = "example-terraform"
}

resource "anxcloud_network_prefix" "v6" {
  vlan_id              = anxcloud_vlan.example.id
  location_id          = data.anxcloud_core_location.anx04.id
  ip_version           = 6
  netmask              = 126
  description_customer = "example-terraform"
}

resource "anxcloud_ip_address" "v4" {
  address           = cidrhost(anxcloud_network_prefix.v4.cidr, 2)
  network_prefix_id = anxcloud_network_prefix.v4.id
}

resource "anxcloud_ip_address" "v6" {
  address           = cidrhost(anxcloud_network_prefix.v6.cidr, 2)
  network_prefix_id = anxcloud_network_prefix.v6.id
}

resource "anxcloud_virtual_server" "example" {
  hostname    = "example-terraform"
  location_id = data.anxcloud_core_location.anx04.id
  template    = "Debian 11"

  cpus   = 4
  memory = 4096

  ssh_key = file("~/.ssh/id_rsa.pub")

  # define bootstrap script
  # e.g. install software
  script = <<-EOT
    #!/bin/bash

    # install nginx server
    apt update && apt install -y nginx
    EOT

  # Set network interface
  network {
    vlan_id         = anxcloud_vlan.example.id
    ips             = [anxcloud_ip_address.v4.id, anxcloud_ip_address.v6.id]
    nic_type        = "virtio"
    bandwidth_limit = 1000
  }

  # Disk 1
  disk {
    disk_gb   = 100
    disk_type = "STD1"
  }

  # Disk 2
  disk {
    disk_gb   = 200
    disk_type = "STD1"
  }

  dns = ["8.8.8.8"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cpus` (Number) Amount of CPUs.
- `disk` (Block List, Min: 1) Virtual Server Disks (see [below for nested schema](#nestedblock--disk))
- `hostname` (String) Virtual server hostname.
- `location_id` (String) Location identifier.
- `memory` (Number) Memory in MB.

### Optional

- `boot_delay` (Number) Boot delay in seconds. Example: (0, 1, …).
- `cpu_performance_type` (String) CPU type. Example: (`best-effort`, `standard`, `enterprise`, `performance`), defaults to `standard`.
- `critical_operation_confirmed` (Boolean) Confirms a critical operation (if needed). Potentially dangerous operations (e.g. resulting in data loss) require an additional confirmation. The parameter is used for VM UPDATE requests.
- `dns` (List of String) DNS configuration. Maximum items 4. Defaults to template settings.
- `enter_bios_setup` (Boolean) Start the VM into BIOS setup on next boot.
- `force_restart_if_needed` (Boolean) Certain operations may only be performed in powered off state. Such as: shrinking memory, shrinking/adding CPU, removing disk and scaling a disk beyond 2 GB. Passing this value as true will always execute a power off and reboot request after completing all other operations. Without this flag set to true scaling operations requiring a reboot will fail.
- `network` (Block List) Network interface (see [below for nested schema](#nestedblock--network))
- `password` (String, Sensitive) Plaintext password. Example: ('!anx123mySuperStrongPassword123anx!', 'go3ju0la1ro3', …). For systems that support it, we strongly recommend using a SSH key instead.
- `script` (String) Script to be executed after provisioning. Consider the corresponding shebang at the beginning of your script. If you want to use PowerShell, the first line should be: #ps1_sysnative.
- `sockets` (Number) Amount of CPU sockets Number of cores have to be a multiple of sockets, as they will be spread evenly across all sockets. Defaults to number of cores, i.e. one socket per CPU core.
- `ssh_key` (String) Public key (instead of password, only for Linux systems). Recommended over providing a plaintext password.
- `tags` (Set of String) Set of tags attached to the resource.
- `template` (String) Named template. Can be used instead of the template_id to select a template. Example: (`Debian 11`, `Windows 2022`).
- `template_build` (String) Template build identifier optionally used with `template`. Will default to latest build. Example: `b42`
- `template_id` (String) Template identifier.
- `template_type` (String) OS template type.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.
- `info` (List of Object) Virtual server info (see [below for nested schema](#nestedatt--info))

<a id="nestedblock--disk"></a>
### Nested Schema for `disk`

Required:

- `disk_gb` (Number) Disk capacity in GB.

Optional:

- `disk_type` (String) Disk category (limits disk performance, e.g. IOPS). Default value depends on location.

Read-Only:

- `disk_exact` (Number) Exact floating point disk size. Not configurable; just for comparison.
- `disk_id` (Number) Device identifier of the disk.


<a id="nestedblock--network"></a>
### Nested Schema for `network`

Required:

- `nic_type` (String) Network interface card type.
- `vlan_id` (String) VLAN identifier.

Optional:

- `bandwidth_limit` (Number) Network interface bandwidth limit in Megabit/s, default: 1000
- `ips` (Set of String) Requested set of IPs and IPs identifiers. IPs are ignored when using template_type 'from_scratch'. Defaults to free IPs from IP pool attached to VLAN.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `read` (String)
- `update` (String)


<a id="nestedatt--info"></a>
### Nested Schema for `info`

Read-Only:

- `cores` (Number) Number of CPU cores.
- `cpu` (Number) Number of CPUs.
- `custom_name` (String) Virtual server custom name.
- `disks_info` (List of Object) Disks info. (see [below for nested schema](#nestedobjatt--info--disks_info))
- `disks_number` (Number) Number of the attached disks.
- `guest_os` (String) Guest operating system.
- `guest_tools_status` (String) Guest tools status.
- `identifier` (String) Identifier of the API resource.
- `location_code` (String) Location code.
- `location_country` (String) Location country.
- `location_name` (String) Location name.
- `name` (String) Virtual server name.
- `network` (List of Object) Network interfaces. (see [below for nested schema](#nestedobjatt--info--network))
- `ram` (Number) Memory in MB.
- `status` (String) Virtual server status.
- `version_tools` (String) Version tools.

<a id="nestedobjatt--info--disks_info"></a>
### Nested Schema for `info.disks_info`

Read-Only:

- `bus_type` (String) Bus type.
- `bus_type_label` (String) Bus type label.
- `disk_gb` (Number) Size of the disk in GB.
- `disk_id` (Number) Disk identifier.
- `disk_type` (String) Disk type.
- `iops` (Number) Disk input/output operations per second.
- `latency` (Number) Disk latency.
- `storage_type` (String) Disk storage type.


<a id="nestedobjatt--info--network"></a>
### Nested Schema for `info.network`

Read-Only:

- `id` (Number) Network interface card identifier.
- `ip_v4` (List of String) List of IPv4 addresses attached to the interface.
- `ip_v6` (List of String) List of IPv6 addresses attached to the interface.
- `mac_address` (String) MAC address of the NIC.
- `nic` (Number) NIC type number.
- `vlan` (String) VLAN identifier.
- `bandwidth_limit` (Number) Network interface bandwidth limit in Megabit/s, default: 1000


