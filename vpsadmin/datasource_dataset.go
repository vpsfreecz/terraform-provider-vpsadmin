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
			"dataset_id": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Dataset ID",
				Required:    true,
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Dataset name",
				Computed:    true,
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
		},
	}
}

func dataSourceDatasetRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()
	id := d.Get("dataset_id").(int)
	ds, err := datasetShow(api, id)

	if err != nil {
		return fmt.Errorf("Invalid dataset ID: %v", err)
	}

	d.SetId(strconv.Itoa(id))
	d.Set("name", ds.Name)
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

	return nil
}
