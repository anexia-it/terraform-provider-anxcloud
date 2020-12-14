package anxcloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func schemaTemplate() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"location_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Location identifier.",
		},
		"template_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "templates",
			Description: "Template type. Defaults to 'templates'.",
		},
		"templates": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Available list of templates.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Template identifier.",
					},
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "OS name.",
					},
					"bit": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "OS bit.",
					},
					"build": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "OS build.",
					},
					"params": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "Params list for the template.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"hostname": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "Hostname parameter.",
									Elem: &schema.Resource{
										Schema: schemaANXParamString(),
									},
								},
								"cpus": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "Hostname parameter.",
									Elem: &schema.Resource{
										Schema: schemaANXParamInt(),
									},
								},
								"memory_mb": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "Memory parameter.",
									Elem: &schema.Resource{
										Schema: schemaANXParamInt(),
									},
								},
								"disk_gb": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "Disk size parameter.",
									Elem: &schema.Resource{
										Schema: schemaANXParamInt(),
									},
								},
								"dns0": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "DNS 0 parameter.",
									Elem: &schema.Resource{
										Schema: schemaANXParamString(),
									},
								},
								"dns1": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "DNS 1 parameter.",
									Elem: &schema.Resource{
										Schema: schemaANXParamString(),
									},
								},
								"dns2": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "DNS 2 parameter.",
									Elem: &schema.Resource{
										Schema: schemaANXParamString(),
									},
								},
								"dns3": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "DNS 3 parameter.",
									Elem: &schema.Resource{
										Schema: schemaANXParamString(),
									},
								},
								"nics": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "NICs parameter.",
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"required": {
												Type:        schema.TypeBool,
												Computed:    true,
												Description: "If it is required value.",
											},
											"label": {
												Type:        schema.TypeString,
												Computed:    true,
												Description: "If it is required value.",
											},
											"default_value": {
												Type:        schema.TypeInt,
												Computed:    true,
												Description: "Default value.",
											},
											"data": {
												Type:        schema.TypeList,
												Computed:    true,
												Description: "Data parameter.",
												Elem: &schema.Resource{
													Schema: map[string]*schema.Schema{
														"id": {
															Type:        schema.TypeInt,
															Computed:    true,
															Description: "Identifier.",
														},
														"name": {
															Type:        schema.TypeString,
															Computed:    true,
															Description: "Name.",
														},
														"default": {
															Type:        schema.TypeBool,
															Computed:    true,
															Description: "If it is default.",
														},
													},
												},
											},
										},
									},
								},
								"vlan": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "VLAN parameter.",
									Elem: &schema.Resource{
										Schema: schemaANXParamString(),
									},
								},
								"ips": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "IPs parameter.",
									Elem: &schema.Resource{
										Schema: schemaANXParamString(),
									},
								},
								"boot_delay_seconds": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "Boot delay parameter.",
									Elem: &schema.Resource{
										Schema: schemaANXParamInt(),
									},
								},
								"enter_bios_setup": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "Enter BIOS parameter.",
									Elem: &schema.Resource{
										Schema: schemaANXParamBool(),
									},
								},
								"password": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "Password parameter.",
									Elem: &schema.Resource{
										Schema: schemaANXParamString(),
									},
								},
								"user": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "User parameter.",
									Elem: &schema.Resource{
										Schema: schemaANXParamString(),
									},
								},
								"disk_type": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "Disk type parameter.",
									Elem: &schema.Resource{
										Schema: schemaANXParamString(),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func schemaANXParamString() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"required": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "If it is required value.",
		},
		"label": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "If it is required value.",
		},
		"default_value": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Default value.",
		},
	}
}

func schemaANXParamBool() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"required": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "If it is required value.",
		},
		"label": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "If it is required value.",
		},
		"default_value": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Default value.",
		},
	}
}

func schemaANXParamInt() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"min_value": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Minimum value.",
		},
		"max_value": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Maximum value.",
		},
		"required": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "If it is required value.",
		},
		"label": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "If it is required value.",
		},
		"default_value": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Default value.",
		},
	}
}
