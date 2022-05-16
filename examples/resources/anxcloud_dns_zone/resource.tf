resource "anxcloud_dns_zone" "example" {
  name         = "example.com"
  admin_email  = "admin@example.com"
  dns_sec_mode = "unvalidated"
  is_master    = true
  refresh      = 14400
  retry        = 3600
  expire       = 604800
  ttl          = 3600
}
