resource "vpsadmin_vps_user_data" "bootstrap" {
  label = "Bootstrap web server"

  # User data format. Other supported values:
  #   - cloudinit_config
  #   - cloudinit_script
  #   - nixos_configuration
  #   - nixos_flake_configuration
  #   - nixos_flake_uri
  format = "script"

  content = <<-EOF
    #!/bin/sh
    set -eu

    export DEBIAN_FRONTEND=noninteractive
    apt-get update
    apt-get install -y nginx
  EOF
}

resource "vpsadmin_vps" "from-stored-user-data" {
  location            = "Praha"
  install_os_template = "debian-13-x86_64-vpsadminos-minimal"
  hostname            = "web-server"
  cpu                 = 4
  memory              = 4096
  diskspace           = 40960
  vps_user_data_id    = vpsadmin_vps_user_data.bootstrap.id
}
