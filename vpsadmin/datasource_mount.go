package vpsadmin

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
)

func dataSourceMount() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceMountRead,

		Schema: map[string]*schema.Schema{
			"vps": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "VPS ID",
				Required:    true,
			},
			"mount_id": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Mount ID",
				Required:    true,
			},
			"dataset": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "ID of the mounted dataset",
				Computed:    true,
			},
			"mountpoint": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Mountpoint inside the VPS",
				Computed:    true,
			},
			"enable": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Whether the mount is enabled",
				Computed:    true,
			},
			"mode": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Read-write or read-only mode",
				Computed:    true,
			},
			"on_start_fail": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Action for when the mount fails during VPS start",
				Computed:    true,
			},
		},
	}
}

func dataSourceMountRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()

	mount, err := mountShow(api, d.Get("vps").(int), d.Get("mount_id").(int))
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(int(mount.Id)))
	d.Set("dataset", mount.Dataset.Id)
	d.Set("mountpoint", mount.Mountpoint)
	d.Set("enable", mount.Enabled)
	d.Set("mode", mount.Mode)
	d.Set("on_start_fail", mount.OnStartFail)

	return nil
}
