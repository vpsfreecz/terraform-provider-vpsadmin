resource "vpsadmin_ssh_key" "my-key" {
  label = "My public key"

  # Set your public key here. The file has to contain exactly one public key.
  key = file("~/.ssh/my_key.pub")
}
