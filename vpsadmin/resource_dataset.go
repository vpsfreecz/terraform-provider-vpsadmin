package vpsadmin

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vpsfreecz/vpsadmin-go-client/client"
	"log"
	"strconv"
)

func resourceDataset() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatasetCreate,
		Read:   resourceDatasetRead,
		Update: resourceDatasetUpdate,
		Delete: resourceDatasetDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDatasetImport,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Dataset name",
				Required:    true,
				ForceNew:    true,
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
				Optional:    true,
				Computed:    true,
			},
			"refquota": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Reference quota, in MiB",
				Optional:    true,
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
				Default:     false,
				Optional:    true,
			},
			"export_id": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Export ID",
				Computed:    true,
			},
			"export_enable": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Enable the NFS server",
				Default:     true,
				Optional:    true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return !d.Get("export_dataset").(bool)
				},
			},
			"export_root_squash": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Enable root squash on the export",
				Default:     false,
				Optional:    true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return !d.Get("export_dataset").(bool)
				},
			},
			"export_read_write": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Read-write access by default",
				Default:     true,
				Optional:    true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return !d.Get("export_dataset").(bool)
				},
			},
			"export_sync": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Server will reply only after changes were committed",
				Default:     true,
				Optional:    true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return !d.Get("export_dataset").(bool)
				},
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

func resourceDatasetCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()

	create := api.Dataset.Create.Prepare()

	input := create.NewInput()
	input.SetName(d.Get("name").(string))

	if v, ok := d.GetOk("quota"); ok {
		input.SetQuota(int64(v.(int)))
	}

	if v, ok := d.GetOk("refquota"); ok {
		input.SetRefquota(int64(v.(int)))
	}

	resp, err := create.Call()

	if err != nil {
		return err
	} else if !resp.Status {
		return fmt.Errorf("Dataset creation failed: %s", resp.Message)
	}

	if err := waitForOperation(resp); err != nil {
		return fmt.Errorf("Dataset creation failed: %v", err)
	}

	d.SetId(strconv.FormatInt(resp.Output.Id, 10))

	if d.Get("export_dataset").(bool) {
		if err := createDatasetExport(api, resp.Output.Id, d); err != nil {
			return err
		}
	}

	return resourceDatasetRead(d, m)
}

func resourceDatasetRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Invalid dataset id: %v", err)
	}

	ds, err := datasetShow(api, id)
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(id))
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

func resourceDatasetUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Invalid dataset id: %v", err)
	}

	dsUpdate := api.Dataset.Update.Prepare()
	dsUpdate.SetPathParamInt("dataset_id", int64(id))

	input := dsUpdate.NewInput()

	if d.HasChange("quota") {
		input.SetQuota(d.Get("quota").(int64))
	}

	if d.HasChange("refquota") {
		input.SetRefquota(d.Get("refquota").(int64))
	}

	if input.AnySelected() {
		updateResp, err := dsUpdate.Call()

		if err != nil {
			return err
		} else if !updateResp.Status {
			return fmt.Errorf("Dataset update failed: %s", updateResp.Message)
		}

		if err := waitForOperation(updateResp); err != nil {
			return fmt.Errorf("Dataset update failed: %v", err)
		}
	}

	if d.HasChange("export_dataset") {
		ds, err := datasetShow(api, id)
		if err != nil {
			return err
		}

		new_export := d.Get("export_dataset").(bool)

		if new_export && ds.Export == nil {
			if err := createDatasetExport(api, ds.Id, d); err != nil {
				return err
			}
		} else if !new_export && ds.Export != nil {
			if err := deleteDatasetExport(api, ds.Export.Id); err != nil {
				return err
			}
		}
	} else {
		if d.HasChanges("export_enable", "export_root_squash", "export_read_write", "export_sync") {
			ds, err := datasetShow(api, id)
			if err != nil {
				return err
			}

			if ds.Export != nil {
				if err := updateDatasetExport(api, ds.Export.Id, d); err != nil {
					return err
				}
			}
		}
	}

	return resourceDatasetRead(d, m)
}

func resourceDatasetDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Invalid dataset id: %v", err)
	}

	ds, err := datasetShow(api, id)
	if err != nil {
		return err
	}

	if ds.Export != nil {
		log.Printf("[INFO] Deleting dataset export: %s", ds.Export.Id)

		if err := deleteDatasetExport(api, ds.Export.Id); err != nil {
			return err
		}
	}

	log.Printf("[INFO] Deleting dataset: %s", d.Id())

	del := api.Dataset.Delete.Prepare()
	del.SetPathParamInt("dataset_id", int64(id))

	resp, err := del.Call()

	if err != nil {
		return err
	} else if !resp.Status {
		return fmt.Errorf("Dataset deletion failed: %s", resp.Message)
	}

	if err := waitForOperation(resp); err != nil {
		return fmt.Errorf("Dataset deletion failed: %v", err)
	}

	return nil
}

func resourceDatasetImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	api := m.(*Config).getClient()

	find := api.Dataset.FindByName.Prepare()

	input := find.NewInput()
	input.SetName(d.Id())

	resp, err := find.Call()
	if err != nil {
		return nil, err
	} else if !resp.Status {
		return nil, fmt.Errorf("Dataset not found: %s", resp.Message)
	}

	d.Set("name", d.Id())
	d.SetId(strconv.FormatInt(resp.Output.Id, 10))

	if err := resourceDatasetRead(d, m); err != nil {
		return nil, fmt.Errorf("invalid dataset id: %v", err)
	}

	results := make([]*schema.ResourceData, 1)
	results[0] = d

	return results, nil
}

func createDatasetExport(api *client.Client, datasetId int64, d *schema.ResourceData) error {
	create := api.Export.Create.Prepare()

	input := create.NewInput()
	input.SetDataset(datasetId)
	input.SetEnabled(d.Get("export_enable").(bool))
	input.SetRootSquash(d.Get("export_root_squash").(bool))
	input.SetRw(d.Get("export_read_write").(bool))
	input.SetSync(d.Get("export_sync").(bool))

	resp, err := create.Call()

	if err != nil {
		return err
	} else if !resp.Status {
		return fmt.Errorf("Export creation failed: %s", resp.Message)
	}

	if err := waitForOperation(resp); err != nil {
		return fmt.Errorf("Export creation failed: %v", err)
	}

	return nil
}

func updateDatasetExport(api *client.Client, id int64, d *schema.ResourceData) error {
	update := api.Export.Update.Prepare()
	update.SetPathParamInt("export_id", id)

	input := update.NewInput()

	if d.HasChange("export_enable") {
		input.SetEnabled(d.Get("export_enable").(bool))
	}

	if d.HasChange("export_root_squashfs") {
		input.SetRootSquash(d.Get("export_root_squash").(bool))
	}

	if d.HasChange("export_read_write") {
		input.SetRw(d.Get("export_read_write").(bool))
	}

	if d.HasChange("export_sync") {
		input.SetSync(d.Get("export_sync").(bool))
	}

	resp, err := update.Call()

	if err != nil {
		return err
	} else if !resp.Status {
		return fmt.Errorf("Export update failed: %s", resp.Message)
	}

	if err := waitForOperation(resp); err != nil {
		return fmt.Errorf("Export update failed: %v", err)
	}

	return nil
}

func deleteDatasetExport(api *client.Client, id int64) error {
	del := api.Export.Delete.Prepare()
	del.SetPathParamInt("export_id", id)

	resp, err := del.Call()
	if err != nil {
		return err
	} else if !resp.Status {
		return fmt.Errorf("Export deletion failed: %s", resp.Message)
	}

	if err := waitForOperation(resp); err != nil {
		return fmt.Errorf("Export deletion failed: %v", err)
	}

	return nil
}
