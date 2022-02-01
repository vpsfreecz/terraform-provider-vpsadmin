package vpsadmin

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strconv"
)

func resourceMount() *schema.Resource {
	return &schema.Resource{
		Create: resourceMountCreate,
		Read:   resourceMountRead,
		Update: resourceMountUpdate,
		Delete: resourceMountDelete,
		Importer: &schema.ResourceImporter{
			State: resourceMountImport,
		},

		Description: "Mount VPS subdatasets into VPS.",

		Schema: map[string]*schema.Schema{
			"vps": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "VPS ID",
				Required:    true,
				ForceNew:    true,
			},
			"dataset": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "ID of the mounted dataset",
				Required:    true,
				ForceNew:    true,
			},
			"mountpoint": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Mountpoint inside the VPS",
				Required:    true,
				ForceNew:    true,
			},
			"enable": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Whether the mount is enabled",
				Default:     true,
				Optional:    true,
			},
			"mode": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Read-write or read-only mode",
				Default:     "rw",
				Optional:    true,
				ForceNew:    true,
			},
			"on_start_fail": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Action for when the mount fails during VPS start",
				Default:     "mount_later",
				Optional:    true,
			},
		},
	}
}

func resourceMountCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()

	create := api.Vps.Mount.Create.Prepare()
	create.SetPathParamInt("vps_id", int64(d.Get("vps").(int)))

	input := create.NewInput()
	input.SetDataset(int64(d.Get("dataset").(int)))
	input.SetMountpoint(d.Get("mountpoint").(string))
	input.SetEnabled(d.Get("enable").(bool))
	input.SetMode(d.Get("mode").(string))
	input.SetOnStartFail(d.Get("on_start_fail").(string))

	resp, err := create.Call()

	if err != nil {
		return err
	} else if !resp.Status {
		return fmt.Errorf("Mount creation failed: %s", resp.Message)
	}

	if err := waitForOperation(resp); err != nil {
		return fmt.Errorf("Mount creation failed: %v", err)
	}

	d.SetId(strconv.FormatInt(resp.Output.Id, 10))

	return resourceMountRead(d, m)
}

func resourceMountRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Invalid mount id: %v", err)
	}

	mount, err := mountShow(api, d.Get("vps").(int), id)
	if err != nil {
		return err
	}

	d.Set("dataset", mount.Dataset.Id)
	d.Set("mountpoint", mount.Mountpoint)
	d.Set("enable", mount.Enabled)
	d.Set("mode", mount.Mode)
	d.Set("on_start_fail", mount.OnStartFail)

	return nil
}

func resourceMountUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Invalid mount id: %v", err)
	}

	update := api.Vps.Mount.Update.Prepare()
	update.SetPathParamInt("vps_id", int64(d.Get("vps").(int)))
	update.SetPathParamInt("mount_id", int64(id))

	input := update.NewInput()

	if d.HasChange("enable") {
		input.SetEnabled(d.Get("enable").(bool))
	}

	if d.HasChange("on_start_fail") {
		input.SetOnStartFail(d.Get("on_start_fail").(string))
	}

	if input.AnySelected() {
		resp, err := update.Call()

		if err != nil {
			return err
		} else if !resp.Status {
			return fmt.Errorf("Mount update failed: %s", resp.Message)
		}

		if err := waitForOperation(resp); err != nil {
			return fmt.Errorf("Mount update failed: %v", err)
		}
	}

	return resourceMountRead(d, m)
}

func resourceMountDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Invalid mount id: %v", err)
	}

	log.Printf("[INFO] Deleting mount: %s", d.Id())

	del := api.Vps.Mount.Delete.Prepare()
	del.SetPathParamInt("vps_id", int64(d.Get("vps").(int)))
	del.SetPathParamInt("mount_id", int64(id))

	resp, err := del.Call()

	if err != nil {
		return err
	} else if !resp.Status {
		return fmt.Errorf("Mount deletion failed: %s", resp.Message)
	}

	if err := waitForOperation(resp); err != nil {
		return fmt.Errorf("Mount deletion failed: %v", err)
	}

	return nil
}

func resourceMountImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	api := m.(*Config).getClient()

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return nil, fmt.Errorf("Invalid mount id: %v", err)
	}

	mount, err := mountFindById(api, id)
	if err != nil {
		return nil, err
	}

	d.Set("vps", mount.Vps.Id)

	if err := resourceMountRead(d, m); err != nil {
		return nil, fmt.Errorf("invalid mount id: %v", err)
	}

	results := make([]*schema.ResourceData, 1)
	results[0] = d

	return results, nil
}
