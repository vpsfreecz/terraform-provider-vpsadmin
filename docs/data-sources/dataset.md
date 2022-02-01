---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "vpsadmin_dataset Data Source - terraform-provider-vpsadmin"
subcategory: ""
description: |-
  
---

# vpsadmin_dataset (Data Source)



## Example Usage

```terraform
data "vpsadmin_dataset" "nas" {
  name = "nas"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **name** (String) Dataset name

### Optional

- **id** (String) The ID of this resource.

### Read-Only

- **atime** (Boolean) Enabled atime
- **avail** (Number) Available space, in MiB
- **compression** (Boolean) Compression enabled
- **export_dataset** (Boolean) Export dataset over NFS
- **export_enable** (Boolean) Enable the NFS server
- **export_id** (Number) Export ID
- **export_ip_address** (String) IP address of the NFS server
- **export_path** (String) Path to mount from the NFS server
- **export_read_write** (Boolean) Read-write access by default
- **export_root_squash** (Boolean) Enable root squash on the export
- **export_sync** (Boolean) Server will reply only after changes were committed
- **full_name** (String) Full dataset name
- **quota** (Number) Quota, in MiB
- **recordsize** (Number) Record size, in bytes
- **referenced** (Number) Referenced space, in MiB
- **refquota** (Number) Reference quota, in MiB
- **relatime** (Boolean) Enabled relatime
- **sync** (String) Sync mode
- **used** (Number) Used space, in MiB

