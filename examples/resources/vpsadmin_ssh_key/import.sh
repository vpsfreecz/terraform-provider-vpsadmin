# Find SSH key ID with vpsfree-client:
#   vpsfreectl user current -o id
#   vpsfreectl user.public_key list $user_id
terraform import vpsadmin_ssh_key.my-key $key_id
