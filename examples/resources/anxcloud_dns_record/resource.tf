resource "anxcloud_dns_record" "example" {
  name      = "webmail"
  zone_name = "example.com"
  type      = "A"
  rdata     = "198.51.100.10"
  ttl       = 3600
}
