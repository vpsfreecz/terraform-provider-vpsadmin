package vpsadmin

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
)

func dataSourceDataset() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDatasetRead,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Dataset name",
				Required:    true,
			},
			"full_name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Full dataset name",
				Computed:    true,
			},
			"used": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Used space, in MiB",
				Computed:    true,
			},
			"referenced": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Referenced space, in MiB",
				Computed:    true,
			},
			"avail": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Available space, in MiB",
				Computed:    true,
			},
			"quota": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Quota, in MiB",
				Computed:    true,
			},
			"refquota": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Reference quota, in MiB",
				Computed:    true,
			},
			"compression": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Compression enabled",
				Computed:    true,
			},
			"recordsize": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Record size, in bytes",
				Computed:    true,
			},
			"atime": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Enabled atime",
				Computed:    true,
			},
			"relatime": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Enabled relatime",
				Computed:    true,
			},
			"sync": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Sync mode",
				Computed:    true,
			},
			"export_dataset": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Export dataset over NFS",
				Computed:    true,
			},
			"export_id": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Export ID",
				Computed:    true,
			},
			"export_enable": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Enable the NFS server",
				Computed:    true,
			},
			"export_root_squash": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Enable root squash on the export",
				Computed:    true,
			},
			"export_read_write": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Read-write access by default",
				Computed:    true,
			},
			"export_sync": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Server will reply only after changes were committed",
				Computed:    true,
			},
			"export_ip_address": &schema.Schema{
				Type:        schema.TypeString,
				Description: "IP address of the NFS server",
				Computed:    true,
			},
			"export_path": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Path to mount from the NFS server",
				Computed:    true,
			},
		},
	}
}

func dataSourceDatasetRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()

	find := api.Dataset.FindByName.Prepare()

	input := find.NewInput()
	input.SetName(d.Get("name").(string))

	resp, err := find.Call()
	if err != nil {
		return err
	} else if !resp.Status {
		return fmt.Errorf("Dataset not found: %s", resp.Message)
	}

	ds := resp.Output

	d.SetId(strconv.Itoa(int(ds.Id)))
	d.Set("full_name", ds.Name)
	d.Set("used", ds.Used)
	d.Set("referenced", ds.Referenced)
	d.Set("avail", ds.Avail)
	d.Set("quota", ds.Quota)
	d.Set("refquota", ds.Refquota)
	d.Set("compression", ds.Compression)
	d.Set("recordsize", ds.Recordsize)
	d.Set("atime", ds.Atime)
	d.Set("relatime", ds.Relatime)
	d.Set("sync", ds.Sync)

	if ds.Export != nil {
		d.Set("export_dataset", true)
		d.Set("export_id", ds.Export.Id)
		d.Set("export_enable", ds.Export.Enabled)
		d.Set("export_root_squash", ds.Export.RootSquash)
		d.Set("export_read_write", ds.Export.Rw)
		d.Set("export_sync", ds.Export.Sync)
		d.Set("export_ip_address", ds.Export.HostIpAddress.Addr)
		d.Set("export_path", ds.Export.Path)
	} else {
		d.Set("export_dataset", false)
		d.Set("export_id", nil)
	}

	return nil
}
