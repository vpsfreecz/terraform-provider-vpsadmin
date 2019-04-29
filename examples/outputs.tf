output "Public IPv4" {
  value = "${vpsadmin_vps.my-vps.public_ipv4_address}"
}

output "Private IPv4" {
  value = "${vpsadmin_vps.my-vps.private_ipv4_address}"
}

output "Public IPv6" {
  value = "${vpsadmin_vps.my-vps.public_ipv6_address}"
}
