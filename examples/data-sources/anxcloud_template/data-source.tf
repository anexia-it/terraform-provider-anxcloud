data "anxcloud_core_location" "anx04" {
  code = "ANX04"
}

data "anxcloud_template" "anx04" {
  location_id = data.anxcloud_core_location.anx04.id
}

// To get a specific OS, filter the `anxcloud_template` data-source:
locals {
  debian_11_templates = values({
    for i, template in data.anxcloud_template.anx04.templates : tostring(i) => template
    if substr(template.name, 0, 9) == "Debian 11"
  })
}
// now you can use local.debian_11_templates[0].id
