---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "vpsadmin_ssh_key Resource - terraform-provider-vpsadmin"
subcategory: ""
description: |-
  
---

# vpsadmin_ssh_key (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **key** (String) Public key
- **label** (String) Public key label

### Optional

- **auto_add** (Boolean) Automatically add this key to new VPS
- **id** (String) The ID of this resource.

### Read-Only

- **comment** (String) Comment from the public key
- **fingerprint** (String) Key fingerprint

