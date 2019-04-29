provider "vpsadmin" {
  # vpsAdmin API URL (default)
  # Can be also changed using environment variable VPSADMIN_API_URL
  # api_url = "https://api.vpsfree.cz"

  # Authentication token
  # Can be also set using environment variable VPSADMIN_API_TOKEN
  auth_token = "${var.vpsadmin_token}"
}

resource "vpsadmin_ssh_key" "my-key" {
  label = "My public key"

  # Set your public key here
  key = "${file("~/.ssh/my_key.pub")}"
}

resource "vpsadmin_vps" "my-vps" {
  # Location label
  location = "Praha"

  # OS template name, see vpsfreectl os_template list -o name
  os_template = "ubuntu-18.04-x86_64-vpsfree"

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
    "${vpsadmin_ssh_key.my-key.id}"
  ]

  # Install nginx once created
  provisioner "remote-exec" {
    inline = [
      "export PATH=$PATH:/usr/bin",
      "apt-get update",
      "apt-get -y install nginx"
    ]

    connection {
      type     = "ssh"
      # Set your private key here
      private_key = "${file("~/.ssh/my_key")}"
      user     = "root"
      timeout  = "2m"
    }
  }
}
