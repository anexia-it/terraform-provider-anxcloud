package anxcloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func schemaVirtualServer() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"hostname": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Virtual server hostname.",
		},
		"location_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Location identifier.",
		},
		"template": {
			Type:         schema.TypeString,
			ForceNew:     true,
			Optional:     true,
			ExactlyOneOf: []string{"template_id", "template"},
			Description: "Named template. Can be used instead of the template_id to select a template. " +
				"Example: (`Debian 11`, `Windows 2022`).",
		},
		"template_build": {
			Type:        schema.TypeString,
			ForceNew:    true,
			Optional:    true,
			Description: "Template build identifier optionally used with `template`. Will default to latest build. Example: `b42`",
		},
		"template_id": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ExactlyOneOf: []string{"template_id", "template"},
			Description:  "Template identifier.",
		},
		"template_type": {
			Type:         schema.TypeString,
			ForceNew:     true,
			Description:  "OS template type.",
			Optional:     true,
			RequiredWith: []string{"template_id"},
		},
		"cpus": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Amount of CPUs.",
		},
		"cpu_performance_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "standard",
			Description: "CPU type. Example: (`best-effort`, `standard`, `enterprise`, `performance`), defaults to `standard`.",
		},
		"sockets": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
			Description: "Amount of CPU sockets Number of cores have to be a multiple of sockets, as they will be spread evenly across all sockets. " +
				"Defaults to number of cores, i.e. one socket per CPU core.",
		},
		"memory": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Memory in MB.",
		},
		"disk": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "Virtual Server Disks",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"disk_id": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Device identifier of the disk.",
					},
					"disk_gb": {
						Type:        schema.TypeInt,
						Required:    true,
						Description: "Disk capacity in GB.",
					},
					"disk_type": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Disk category (limits disk performance, e.g. IOPS). Default value depends on location.",
					},
					"disk_exact": {
						Type:        schema.TypeFloat,
						Computed:    true,
						Description: "Exact floating point disk size. Not configurable; just for comparison.",
					},
				},
			},
		},
		"network": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Network interface",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"vlan_id": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "VLAN identifier.",
					},
					"nic_type": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Network interface card type.",
					},
					"bandwidth_limit": {
						Type:        schema.TypeInt,
						Optional:    true,
						Default:     1000,
						Description: "Network interface bandwidth limit in Megabit/s, default: 1000",
					},
					"ips": {
						Type:     schema.TypeSet,
						Optional: true,
						ForceNew: true,
						Description: "Requested set of IPs and IPs identifiers. IPs are ignored when using template_type 'from_scratch'. " +
							"Defaults to free IPs from IP pool attached to VLAN.",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
		"dns": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    4,
			ForceNew:    true,
			Description: "DNS configuration. Maximum items 4. Defaults to template settings.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"password": {
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
			ForceNew:    true,
			Description: "Plaintext password. Example: ('!anx123mySuperStrongPassword123anx!', 'go3ju0la1ro3', …). For systems that support it, we strongly recommend using a SSH key instead.",
		},
		"ssh_key": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "Public key (instead of password, only for Linux systems). Recommended over providing a plaintext password.",
		},
		"script": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
			Description: "Script to be executed after provisioning. " +
				"Consider the corresponding shebang at the beginning of your script. " +
				"If you want to use PowerShell, the first line should be: #ps1_sysnative.",
		},
		"boot_delay": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Boot delay in seconds. Example: (0, 1, …).",
		},
		"enter_bios_setup": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Start the VM into BIOS setup on next boot.",
		},
		"force_restart_if_needed": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
			Description: "Certain operations may only be performed in powered off state. " +
				"Such as: shrinking memory, shrinking/adding CPU, removing disk and scaling a disk beyond 2 GB. " +
				"Passing this value as true will always execute a power off and reboot request after completing all other operations. " +
				"Without this flag set to true scaling operations requiring a reboot will fail.",
		},
		"critical_operation_confirmed": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
			Description: "Confirms a critical operation (if needed). " +
				"Potentially dangerous operations (e.g. resulting in data loss) require an additional confirmation. " +
				"The parameter is used for VM UPDATE requests.",
		},
		"info": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Virtual server info",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"identifier": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: identifierDescription,
					},
					"status": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Virtual server status.",
					},
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Virtual server name.",
					},
					"custom_name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Virtual server custom name.",
					},
					"location_code": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Location code.",
					},
					"location_country": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Location country.",
					},
					"location_name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Location name.",
					},
					"cpu": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Number of CPUs.",
					},
					"cores": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Number of CPU cores.",
					},
					"ram": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Memory in MB.",
					},
					"disks_number": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Number of the attached disks.",
					},
					"disks_info": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "Disks info.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"disk_id": {
									Type:        schema.TypeInt,
									Computed:    true,
									Description: "Disk identifier.",
								},
								"disk_gb": {
									Type:        schema.TypeInt,
									Computed:    true,
									Description: "Size of the disk in GB.",
								},
								"disk_type": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Disk type.",
								},
								"iops": {
									Type:        schema.TypeInt,
									Computed:    true,
									Description: "Disk input/output operations per second.",
								},
								"latency": {
									Type:        schema.TypeInt,
									Computed:    true,
									Description: "Disk latency.",
								},
								"storage_type": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Disk storage type.",
								},
								"bus_type": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Bus type.",
								},
								"bus_type_label": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Bus type label.",
								},
							},
						},
					},
					"network": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "Network interfaces.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"id": {
									Type:        schema.TypeInt,
									Computed:    true,
									Description: "Network interface card identifier.",
								},
								"ip_v4": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "List of IPv4 addresses attached to the interface.",
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
								"ip_v6": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "List of IPv6 addresses attached to the interface.",
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
								"nic": {
									Type:        schema.TypeInt,
									Computed:    true,
									Description: "NIC type number.",
								},
								"bandwidth_limit": {
									Type:        schema.TypeInt,
									Computed:    true,
									Description: "Bandwidth limit of the interface in Megabit/s",
								},
								"vlan": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "VLAN identifier.",
								},
								"mac_address": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "MAC address of the NIC.",
								},
							},
						},
					},
					"guest_os": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Guest operating system.",
					},
					"version_tools": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Version tools.",
					},
					"guest_tools_status": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Guest tools status.",
					},
				},
			},
		},
	}
}
