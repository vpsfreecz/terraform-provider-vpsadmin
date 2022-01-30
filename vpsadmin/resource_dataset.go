package vpsadmin

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

	return resourceDatasetRead(d, m)
}

func resourceDatasetDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Invalid dataset id: %v", err)
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
	err := resourceDatasetRead(d, m)
	if err != nil {
		return nil, fmt.Errorf("invalid dataset id: %v", err)
	}

	results := make([]*schema.ResourceData, 1)
	results[0] = d

	return results, nil
}
