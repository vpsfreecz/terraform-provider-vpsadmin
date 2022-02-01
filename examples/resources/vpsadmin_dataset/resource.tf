# Create a dataset on NAS (Network-Attached Storage)
resource "vpsadmin_dataset" "nas-backups" {
  name = "nas/backups"

  # Quota in MiB
  quota = 100 * 1024

  # Export dataset over NFS
  export_dataset = true
}

# Create a subdataset in a VPS
resource "vpsadmin_vps" "my-vps" {
  location = "Praha"
  install_os_template = "ubuntu-20.04-x86_64-vpsadminos-minimal"
  cpu = 8
  memory = 4096
  diskspace = 60*1024
}

resource "vpsadmin_dataset" "my-subdataset" {
  name = "vps${vpsadmin_vps.my-vps.id}/my-subdataset"
  refquota = 20 * 1024
}
