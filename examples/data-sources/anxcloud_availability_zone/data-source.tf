data "anxcloud_availability_zone" "zoneA" {
  location_id = data.anxcloud_core_location.anx04.id
  name        = "Zone A"
}