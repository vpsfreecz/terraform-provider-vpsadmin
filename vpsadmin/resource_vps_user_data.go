package vpsadmin

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vpsfreecz/vpsadmin-go-client/client"
)

func resourceVpsUserData() *schema.Resource {
	return &schema.Resource{
		Create: resourceVpsUserDataCreate,
		Read:   resourceVpsUserDataRead,
		Update: resourceVpsUserDataUpdate,
		Delete: resourceVpsUserDataDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVpsUserDataImport,
		},

		Description: "Stores reusable user data in vpsAdmin. A VPS can use this resource's ID during creation. Updating the stored object does not apply the new content to VPSes that already used it. To replace a VPS when this resource changes, add this resource to the VPS's `lifecycle.replace_triggered_by` list.",

		Schema: map[string]*schema.Schema{
			"label": {
				Type:         schema.TypeString,
				Description:  "Label used to identify the stored user data",
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"format": {
				Type:         schema.TypeString,
				Description:  fmt.Sprintf("Content format. If omitted, `script` is used. Supported values are %s.", supportedUserDataFormatsText),
				Optional:     true,
				Default:      defaultUserDataFormat,
				ValidateFunc: validation.StringInSlice(supportedUserDataFormats, false),
			},
			"content": {
				Type:         schema.TypeString,
				Description:  "Content stored in vpsAdmin. Content is limited to 65,535 UTF-8 bytes. Terraform state stores a SHA-256 digest instead of the content.",
				Required:     true,
				Sensitive:    true,
				StateFunc:    userDataStateHash,
				ValidateFunc: validateUserDataContent,
			},
		},
	}
}

func resourceVpsUserDataCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()
	create := api.VpsUserData.Create.Prepare()
	input := create.NewInput()
	input.SetLabel(d.Get("label").(string))
	input.SetFormat(d.Get("format").(string))
	input.SetContent(d.Get("content").(string))

	log.Printf(
		"[DEBUG] Creating vpsAdmin user data with label %q and format %q",
		input.Label,
		input.Format,
	)

	resp, err := create.Call()
	if err != nil {
		return err
	} else if !resp.Status {
		return fmt.Errorf("vpsAdmin user data creation failed: %s", resp.Message)
	}

	d.SetId(strconv.FormatInt(resp.Output.Id, 10))
	return resourceVpsUserDataRead(d, m)
}

func resourceVpsUserDataRead(d *schema.ResourceData, m interface{}) error {
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid vpsAdmin user data ID: %v", err)
	}

	resp, err := vpsUserDataShow(m.(*Config).getClient(), id)
	if err != nil {
		return err
	} else if !resp.Status {
		if responseReportsResourceNotFound(resp.Envelope) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("failed to fetch vpsAdmin user data: %s", resp.Message)
	}

	return setVpsUserDataResourceState(d, resp.Output)
}

func resourceVpsUserDataUpdate(d *schema.ResourceData, m interface{}) error {
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid vpsAdmin user data ID: %v", err)
	}

	update := m.(*Config).getClient().VpsUserData.Update.Prepare()
	update.SetPathParamInt("vps_user_data_id", id)
	input := update.NewInput()

	if d.HasChange("label") {
		input.SetLabel(d.Get("label").(string))
	}
	if d.HasChange("format") {
		input.SetFormat(d.Get("format").(string))
	}
	if d.HasChange("content") {
		input.SetContent(d.Get("content").(string))
	}

	if input.AnySelected() {
		log.Printf("[DEBUG] Updating vpsAdmin user data %d", id)
		resp, err := update.Call()
		if err != nil {
			return err
		} else if !resp.Status {
			return fmt.Errorf("vpsAdmin user data update failed: %s", resp.Message)
		}
	}

	return resourceVpsUserDataRead(d, m)
}

func resourceVpsUserDataDelete(d *schema.ResourceData, m interface{}) error {
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid vpsAdmin user data ID: %v", err)
	}

	log.Printf("[INFO] Deleting vpsAdmin user data: %d", id)
	del := m.(*Config).getClient().VpsUserData.Delete.Prepare()
	del.SetPathParamInt("vps_user_data_id", id)
	resp, err := del.Call()
	if err != nil {
		return err
	} else if !resp.Status && !responseReportsResourceNotFound(resp.Envelope) {
		return fmt.Errorf("vpsAdmin user data deletion failed: %s", resp.Message)
	}

	d.SetId("")
	return nil
}

func resourceVpsUserDataImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if err := resourceVpsUserDataRead(d, m); err != nil {
		return nil, fmt.Errorf("invalid vpsAdmin user data ID: %v", err)
	}
	if d.Id() == "" {
		return nil, fmt.Errorf("invalid vpsAdmin user data ID: resource not found")
	}

	return []*schema.ResourceData{d}, nil
}

func vpsUserDataShow(api *client.Client, id int64) (*client.ActionVpsUserDataShowResponse, error) {
	show := api.VpsUserData.Show.Prepare()
	show.SetPathParamInt("vps_user_data_id", id)
	return show.Call()
}

func setVpsUserDataResourceState(d *schema.ResourceData, data *client.ActionVpsUserDataShowOutput) error {
	for name, value := range map[string]interface{}{
		"label":   data.Label,
		"format":  data.Format,
		"content": userDataStateHash(data.Content),
	} {
		if err := d.Set(name, value); err != nil {
			return fmt.Errorf("failed to store vpsAdmin user data field %q: %v", name, err)
		}
	}

	return nil
}

func responseReportsResourceNotFound(envelope *client.Envelope) bool {
	if envelope == nil {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(envelope.Message), "resource not found") {
		return true
	}

	for _, errors := range envelope.Errors {
		for _, message := range errors {
			if strings.EqualFold(strings.TrimSpace(message), "resource not found") {
				return true
			}
		}
	}

	return false
}
