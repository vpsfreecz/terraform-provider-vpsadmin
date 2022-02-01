resource "vpsadmin_vps" "my-vps" {
  # Location label
  # Possible values
  #   - using vpsfree-client: vpsfreectl location list -o label
  #   - using curl: curl https://api.vpsfree.cz/locations
  location = "Praha"

  # OS template name
  # Possible values
  #   - using vpsfree-client: vpsfreectl os_template list -o name
  #   - using curl: curl https://api.vpsfree.cz/os_templates
  install_os_template = "ubuntu-20.04-x86_64-vpsadminos-minimal"

  # vpsAdmin-managed hostname
  hostname = "my-vps"

  # Number of CPU cores
  cpu = 8

  # Available memory in MB
  memory = 4096

  # Root dataset size in MB
  diskspace = 122880

  # Public keys deployed to /root/.ssh/authorized_keys
  ssh_keys = [
    vpsadmin_ssh_key.my-key.id,
  ]
}
