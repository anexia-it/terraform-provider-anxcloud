package anxcloud

import (
	"github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/templates"
)

// expanders

// flatteners

func flattenTemplates(in []templates.Template) []interface{} {
	att := []interface{}{}
	if len(in) < 1 {
		return att
	}

	for _, t := range in {
		m := map[string]interface{}{}

		m["id"] = t.ID
		m["name"] = t.Name
		m["build"] = t.Build
		m["bit"] = t.WordSize
		m["params"] = flattenTemplateParameters(t.Parameters)

		att = append(att, m)
	}

	return att
}

func flattenTemplateParameters(in templates.Parameters) []interface{} {
	m := map[string]interface{}{}

	m["hostname"] = flattenANXStringParam(in.Hostname)
	m["dns0"] = flattenANXStringParam(in.DNS0)
	m["dns1"] = flattenANXStringParam(in.DNS1)
	m["dns2"] = flattenANXStringParam(in.DNS2)
	m["dns3"] = flattenANXStringParam(in.DNS3)
	m["vlan"] = flattenANXStringParam(in.VLAN)
	m["password"] = flattenANXStringParam(in.Password)
	m["user"] = flattenANXStringParam(in.User)
	m["disk_type"] = flattenANXStringParam(in.DiskType)
	m["ips"] = flattenANXStringParam(in.IPs)

	m["cpus"] = flattenANXIntParam(in.CPUs)
	m["memory_mb"] = flattenANXIntParam(in.MemoryMB)
	m["disk_gb"] = flattenANXIntParam(in.DiskGB)
	m["boot_delay_seconds"] = flattenANXIntParam(in.BootDelaySeconds)

	m["enter_bios_setup"] = flattenANXBoolParam(in.EnterBIOSSetup)

	m["nics"] = flattenTemplateNICs(in.NICs)

	return []interface{}{m}
}

func flattenANXStringParam(in templates.StringParameter) []interface{} {
	m := map[string]interface{}{}
	m["default_value"] = in.Default
	m["required"] = in.Required
	m["label"] = in.Label

	return []interface{}{m}
}

func flattenANXIntParam(in templates.IntParameter) []interface{} {
	m := map[string]interface{}{}
	m["default_value"] = in.Default
	m["required"] = in.Required
	m["label"] = in.Label
	m["min_value"] = in.Minimum
	m["max_value"] = in.Maximum

	return []interface{}{m}
}

func flattenANXBoolParam(in templates.BoolParameter) []interface{} {
	m := map[string]interface{}{}
	m["default_value"] = in.Default
	m["required"] = in.Required
	m["label"] = in.Label

	return []interface{}{m}
}

func flattenTemplateNICs(in templates.NICParameter) []interface{} {
	m := map[string]interface{}{}
	m["default_value"] = in.Default
	m["required"] = in.Required
	m["label"] = in.Label

	att := []interface{}{}
	for _, n := range in.NICs {
		nm := map[string]interface{}{}
		nm["id"] = n.ID
		nm["name"] = n.Name
		nm["default"] = n.Default

		att = append(att, nm)
	}
	m["data"] = att

	return []interface{}{m}
}
