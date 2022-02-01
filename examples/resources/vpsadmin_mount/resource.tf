# Mount a VPS subdataset
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

resource "vpsadmin_mount" "vps-subdataset" {
  vps = vpsadmin_vps.my-vps.id
  dataset = vpsadmin_dataset.vps-subdataset.id
  mountpoint = "/mnt/subdataset"
}
