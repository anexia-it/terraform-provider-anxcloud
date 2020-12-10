package anxcloud

import (
	"github.com/google/go-cmp/cmp"
	"testing"

	"github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/templates"
)

// expanders tests

// flatteners tests

func TestFlattenTemplates(t *testing.T) {
	cases := []struct {
		Input          []templates.Template
		ExpectedOutput []interface{}
	}{
		{
			[]templates.Template{
				{
					ID:       "1234",
					Name:     "Centos7",
					WordSize: "64",
					Build:    "b14",
					Parameters: templates.Parameters{
						Hostname: templates.StringParameter{
							Required: true,
							Label:    "hostname",
							Default:  "",
						},
						CPUs: templates.IntParameter{
							Minimum:  1,
							Maximum:  16,
							Required: true,
							Label:    "cpus",
							Default:  2,
						},
						MemoryMB: templates.IntParameter{
							Minimum:  1000,
							Maximum:  16000,
							Required: true,
							Label:    "memory",
							Default:  0,
						},
						DiskGB: templates.IntParameter{
							Minimum:  50,
							Maximum:  300,
							Required: true,
							Label:    "disk gb",
							Default:  50,
						},
						DNS0: templates.StringParameter{
							Required: false,
							Label:    "dns0",
							Default:  "",
						},
						DNS1: templates.StringParameter{
							Required: false,
							Label:    "dns1",
							Default:  "",
						},
						DNS2: templates.StringParameter{
							Required: false,
							Label:    "dns2",
							Default:  "",
						},
						DNS3: templates.StringParameter{
							Required: false,
							Label:    "dns3",
							Default:  "",
						},
						NICs: templates.NICParameter{
							Required: false,
							Label:    "nics",
							Default:  0,
							NICs: []templates.NIC{
								{
									ID:      1,
									Name:    "ens100",
									Default: false,
								},
								{
									ID:      2,
									Name:    "ens200",
									Default: false,
								},
							},
						},
						VLAN: templates.StringParameter{
							Required: true,
							Label:    "vlan",
							Default:  "",
						},
						IPs: templates.StringParameter{
							Required: false,
							Label:    "ips",
							Default:  "",
						},
						BootDelaySeconds: templates.IntParameter{
							Minimum:  0,
							Maximum:  10,
							Required: false,
							Label:    "boot delay seconds",
							Default:  0,
						},
						EnterBIOSSetup: templates.BoolParameter{
							Required: false,
							Label:    "enter bios",
							Default:  false,
						},
						Password: templates.StringParameter{
							Required: false,
							Label:    "password",
							Default:  "",
						},
						User: templates.StringParameter{
							Required: false,
							Label:    "user",
							Default:  "",
						},
						DiskType: templates.StringParameter{
							Required: false,
							Label:    "disk type",
							Default:  "",
						},
					},
				},
			},
			[]interface{}{
				map[string]interface{}{
					"id":    "1234",
					"name":  "Centos7",
					"bit":   "64",
					"build": "b14",
					"params": []interface{}{
						map[string]interface{}{
							"boot_delay_seconds": []interface{}{
								map[string]interface{}{
									"default_value": 0,
									"label":         "boot delay seconds",
									"max_value":     10,
									"min_value":     0,
									"required":      false,
								},
							},
							"cpus": []interface{}{
								map[string]interface{}{
									"default_value": 2,
									"label":         "cpus",
									"max_value":     16,
									"min_value":     1,
									"required":      true,
								},
							},
							"disk_gb": []interface{}{
								map[string]interface{}{
									"default_value": 50,
									"label":         "disk gb",
									"max_value":     300,
									"min_value":     50,
									"required":      true,
								},
							},
							"disk_type": []interface{}{
								map[string]interface{}{
									"default_value": "",
									"label":         "disk type",
									"required":      false,
								},
							},
							"dns0": []interface{}{
								map[string]interface{}{"default_value": "", "label": "dns0", "required": false},
							},
							"dns1": []interface{}{
								map[string]interface{}{"default_value": "", "label": "dns1", "required": false},
							},
							"dns2": []interface{}{
								map[string]interface{}{"default_value": "", "label": "dns2", "required": false},
							},
							"dns3": []interface{}{
								map[string]interface{}{"default_value": "", "label": "dns3", "required": false},
							},
							"enter_bios_setup": []interface{}{
								map[string]interface{}{
									"default_value": false,
									"label":         "enter bios",
									"required":      false,
								},
							},
							"hostname": []interface{}{
								map[string]interface{}{"default_value": "", "label": "hostname", "required": true},
							},
							"ips": []interface{}{
								map[string]interface{}{
									"default_value": "",
									"label":         "ips",
									"required":      false,
								},
							},
							"memory_mb": []interface{}{
								map[string]interface{}{
									"default_value": 0,
									"label":         "memory",
									"max_value":     16000,
									"min_value":     1000,
									"required":      true,
								},
							},
							"nics": []interface{}{
								map[string]interface{}{
									"data": []interface{}{
										map[string]interface{}{"default": false, "id": 1, "name": "ens100"},
										map[string]interface{}{"default": false, "id": 2, "name": "ens200"},
									},
									"default_value": 0,
									"label":         "nics",
									"required":      false,
								},
							},
							"password": []interface{}{
								map[string]interface{}{
									"default_value": "",
									"label":         "password",
									"required":      false,
								},
							},
							"user": []interface{}{
								map[string]interface{}{"default_value": "", "label": "user", "required": false},
							},
							"vlan": []interface{}{
								map[string]interface{}{"default_value": "", "label": "vlan", "required": true},
							},
						},
					},
				},
			},
		},
		{
			[]templates.Template{},
			[]interface{}{},
		},
	}

	for _, tc := range cases {
		output := flattenTemplates(tc.Input)
		if diff := cmp.Diff(tc.ExpectedOutput, output); diff != "" {
			t.Fatalf("Unexpected output from expander: mismatch (-want +got):\n%s", diff)
		}
	}
}
