package vpsadmin

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vpsfreecz/vpsadmin-go-client/client"
	"fmt"
	"log"
	"strconv"
)

func resourceVps() *schema.Resource {
	return &schema.Resource{
		Create: resourceVpsCreate,
		Read:   resourceVpsRead,
		Update: resourceVpsUpdate,
		Delete: resourceVpsDelete,

		Schema: map[string]*schema.Schema{
			"location": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"node": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"os_template": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hostname": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"cpu": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"memory": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"swap": &schema.Schema{
				Type:     schema.TypeInt,
				Default:  0,
				Optional: true,
			},
			"diskspace": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"public_ipv4_count": &schema.Schema{
				Type:     schema.TypeInt,
				Default:  1,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// This field is used only when the VPS is being created
					return d.Id() != ""
				},
			},
			"private_ipv4_count": &schema.Schema{
				Type:     schema.TypeInt,
				Default:  0,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// This field is used only when the VPS is being created
					return d.Id() != ""
				},
			},
			"public_ipv6_count": &schema.Schema{
				Type:     schema.TypeInt,
				Default:  1,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// This field is used only when the VPS is being created
					return d.Id() != ""
				},
			},
		},
	}
}

func resourceVpsCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()

	locationId, err := getLocationIdByLabel(api, d.Get("location").(string))

	if err != nil {
		return err
	}

	templateId, err := getOsTemplateIdByName(api, d.Get("os_template").(string))

	if err != nil {
		return err
	}

	create := api.Vps.Create.Prepare()
	create.SetInput(&client.ActionVpsCreateInput{
		Location: locationId,
		OsTemplate: templateId,
		Hostname: d.Get("hostname").(string),
		Cpu: int64(d.Get("cpu").(int)),
		Memory: int64(d.Get("memory").(int)),
		Swap: int64(d.Get("swap").(int)),
		Diskspace: int64(d.Get("diskspace").(int)),
		Ipv4: int64(d.Get("public_ipv4_count").(int)),
		Ipv4Private: int64(d.Get("private_ipv4_count").(int)),
		Ipv6: int64(d.Get("public_ipv6_count").(int)),
	})
	create.Input.SelectParameters(
		"Location", "OsTemplate", "Hostname",
		"Cpu", "Memory", "Swap", "Diskspace",
		"Ipv4", "Ipv4Private", "Ipv6",
	)

	log.Printf("[DEBUG] VPS create configuration: %#v", create.Input)

	resp, err := create.Call()

	if err != nil {
		return err
	} else if !resp.Status {
		return fmt.Errorf("VPS creation failed: %s", resp.Message)
	}

	if err := waitForOperation(resp); err != nil {
		return fmt.Errorf("VPS creation failed: %v", err)
	}

	d.SetId(strconv.FormatInt(resp.Output.Id, 10))

	return resourceVpsRead(d, m)
}

func resourceVpsRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()

	id, err := strconv.Atoi(d.Id())

	if err != nil {
		return fmt.Errorf("Invalid VPS id: %v", err)
	}

	vps, err := vpsShow(api, id)

	if err != nil {
		return err
	}

	// Dataset cannot be prefetched, API limitation
	ds, err := datasetShow(api, int(vps.Dataset.Id))

	if err != nil {
		return err
	}

	d.Set("location", vps.Node.Location.Label)
	d.Set("node", vps.Node.DomainName)
	d.Set("os_template", vps.OsTemplate.Name)
	d.Set("hostname", vps.Hostname)
	d.Set("cpu", vps.Cpu)
	d.Set("memory", vps.Memory)
	d.Set("swap", vps.Swap)
	d.Set("diskspace", ds.Refquota)

	return nil
}

func resourceVpsUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()

	id, err := strconv.Atoi(d.Id())

	if err != nil {
		return fmt.Errorf("Invalid VPS id: %v", err)
	}

	d.Partial(true)

	vpsUpdate := api.Vps.Update.Prepare()
	vpsUpdate.SetPathParamInt("vps_id", int64(id))

	input := &client.ActionVpsUpdateInput{}

	if d.HasChange("hostname") {
		input.SetHostname(d.Get("hostname").(string))
	}

	if d.HasChange("cpu") {
		input.SetCpu(int64(d.Get("cpu").(int)))
	}

	if d.HasChange("memory") {
		input.SetMemory(int64(d.Get("memory").(int)))
	}

	if d.HasChange("swap") {
		input.SetSwap(int64(d.Get("swap").(int)))
	}

	if input.AnySelected() {
		vpsUpdate.SetInput(input)

		vpsResp, err := vpsUpdate.Call()

		if err != nil {
			return err
		} else if !vpsResp.Status {
			return fmt.Errorf("VPS update failed: %s", vpsResp.Message)
		}

		if err := waitForOperation(vpsResp); err != nil {
			return fmt.Errorf("VPS update failed: %v", err)
		}

		d.SetPartial("hostname")
		d.SetPartial("cpu")
		d.SetPartial("memory")
		d.SetPartial("swap")
	}

	if d.HasChange("diskspace") {
		vps, err := vpsShow(api, id)
		datasetUpdate := api.Dataset.Update.Prepare()
		datasetUpdate.SetPathParamInt("dataset_id", vps.Dataset.Id)
		datasetUpdate.SetInput(&client.ActionDatasetUpdateInput{
			Refquota: int64(d.Get("diskspace").(int)),
		})
		datasetUpdate.Input.SelectParameters("Refquota")

		datasetResp, err := datasetUpdate.Call()

		if err != nil {
			return err
		} else if !datasetResp.Status {
			return fmt.Errorf("Dataset update failed: %s", datasetResp.Message)
		}

		if err := waitForOperation(datasetResp); err != nil {
			return fmt.Errorf("Dataset update failed: %v", err)
		}

		d.SetPartial("diskspace")
	}

	d.Partial(false)

	return resourceVpsRead(d, m)
}

func resourceVpsDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()

	id, err := strconv.Atoi(d.Id())

	if err != nil {
		return fmt.Errorf("Invalid VPS id: %v", err)
	}

	log.Printf("[INFO] Deleting VPS: %s", d.Id())

	del := api.Vps.Delete.Prepare()
	del.SetPathParamInt("vps_id", int64(id))

	resp, err := del.Call()

	if err != nil {
		return err
	} else if !resp.Status {
		return fmt.Errorf("VPS deletion failed: %s", resp.Message)
	}

	if err := waitForOperation(resp); err != nil {
		return fmt.Errorf("VPS deletion failed: %v", err)
	}

	return nil
}
