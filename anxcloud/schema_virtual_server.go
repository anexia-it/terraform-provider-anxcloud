package anxcloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func schemaVirtualServer() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Virtual server identifier.",
		},
		"location_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Location identifier.",
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
		"template_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Template identifier.",
		},
		"template_type": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "OS template type.",
		},
		"hostname": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Virtual server hostname.",
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
			Description: "CPU type. Example: (\"best-effort\", \"standard\", \"enterprise\", \"performance\"), defaults to \"standard\".",
		},
		"sockets": {
			Type:     schema.TypeInt,
			Optional: true,
			Description: "Amount of CPU sockets Number of cores have to be a multiple of sockets, as they will be spread evenly across all sockets." +
				"Defaults to number of cores, i.e. one socket per CPU core.",
		},
		"memory": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Memory in MB.",
		},
		"disk": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Requested disk capacity in GB.",
		},
		"disk_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Requested disk category (limits disk performance, e.g. IOPS). Default as defined by data center.",
		},
		"disks_number": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Number of the attached disks.",
		},
		"disk_info": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Disks info.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"disk_id": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Disk identifier.",
					},
					"disk_gb": {
						Type:        schema.TypeString,
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
			Optional:    true,
			Description: "Network interfaces",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Network interface card identifier.",
					},
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
					"ips": {
						Type:     schema.TypeList,
						Optional: true,
						Elem:     schema.TypeString,
						Description: "Requested list of IPs and IPs identifiers. IPs are ignored when using template_type 'from_scratch'." +
							"Defaults to free IPs from IP pool attached to VLAN.",
					},
					"ip_v4": {
						Type:        schema.TypeList,
						Computed:    true,
						Elem:        schema.TypeString,
						Description: "List of IPv4 addresses to the interface.",
					},
					"ip_v6": {
						Type:        schema.TypeList,
						Computed:    true,
						Elem:        schema.TypeString,
						Description: "List of IPv6 addresses to the interface.",
					},
					"nic": {
						Type:     schema.TypeString,
						Computed: true,
						//Description: TODO: fill description of nic
					},
					"mac_address": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "MAC address of the NIC",
					},
				},
			},
		},
		"dns": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    4,
			Elem:        schema.TypeString,
			Description: "DNS configuration. Maximum items 4. Defaults to template settings.",
		},
		"password": {
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
			Description: "Plaintext password. Example: ('!anx123mySuperStrongPassword123anx!', 'go3ju0la1ro3', …). USE IT AT YOUR OWN RISK! (or SSH key instead).",
		},
		"ssh_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Public key (instead of password, only for Linux systems). Recommended over providing a plaintext password.",
		},
		"script": {
			Type:     schema.TypeString,
			Optional: true,
			Description: "Script to be executed after provisioning. Should be base64 encoded." +
				"Consider the corresponding shebang at the beginning of your script." +
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
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Virtual server status.",
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
	}
}
