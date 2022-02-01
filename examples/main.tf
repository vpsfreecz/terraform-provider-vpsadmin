provider "vpsadmin" {
  # vpsAdmin API URL (default)
  # Can be also changed using environment variable VPSADMIN_API_URL
  # api_url = "https://api.vpsfree.cz"

  # Authentication token
  # Can be also set using environment variable VPSADMIN_API_TOKEN
  auth_token = var.vpsadmin_token
}

# Declare a public key for connection over SSH
resource "vpsadmin_ssh_key" "my-key" {
  label = "My public key"

  # Set your public key here. The file has to contain exactly one public key.
  key = file("~/.ssh/my_key.pub")
}

# Create a VPS
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

  # Install nginx once created
  provisioner "remote-exec" {
    inline = [
      "export PATH=$PATH:/usr/bin",
      "apt-get update",
      "apt-get -y install nginx",

      # Uncomment to mount dataset nas/backups over NFS
      # "apt-get -y install nfs-common",
      # "mkdir -p /mnt/backups",
      # "mount -t nfs ${vpsadmin_dataset.nas-backups.export_ip_address}:${vpsadmin_dataset.nas-backups.export_path} /mnt/backups",
    ]

    connection {
      type = "ssh"
      host = vpsadmin_vps.my-vps.public_ipv4_address

      # Set your private key here
      private_key = file("~/.ssh/my_key")
      user        = "root"
      timeout     = "2m"
    }
  }
}

# Create a dataset on NAS (Network-Attached Storage)
resource "vpsadmin_dataset" "nas-backups" {
  name = "nas/backups"

  # Export the dataset over NFS
  export_dataset = true
}
