package vpsadmin

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strconv"
	"strings"
)

func resourceSshKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceSshKeyCreate,
		Read:   resourceSshKeyRead,
		Update: resourceSshKeyUpdate,
		Delete: resourceSshKeyDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSshKeyImport,
		},

		Schema: map[string]*schema.Schema{
			"label": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Public key label",
				Required:    true,
			},
			"key": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Public key",
				Required:    true,
				StateFunc: func(val interface{}) string {
					return strings.TrimSpace(val.(string))
				},
			},
			"auto_add": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Automatically add this key to new VPS",
				Optional:    true,
				Default:     false,
			},
			"fingerprint": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Key fingerprint",
				Computed:    true,
			},
			"comment": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Comment from the public key",
				Computed:    true,
			},
		},
	}
}

func resourceSshKeyCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()

	user, err := getCurrentUser(api)

	if err != nil {
		return err
	}

	create := api.User.PublicKey.Create.Prepare()
	create.SetPathParamInt("user_id", user.Id)

	input := create.NewInput()
	input.SetLabel(d.Get("label").(string))
	input.SetKey(strings.TrimSpace(d.Get("key").(string)))
	input.SetAutoAdd(d.Get("auto_add").(bool))

	log.Printf("[DEBUG] SSH key create configuration: %#v", create.Input)

	resp, err := create.Call()

	if err != nil {
		return err
	} else if !resp.Status {
		return fmt.Errorf("SSH key creation failed: %s", resp.Message)
	}

	d.SetId(strconv.FormatInt(resp.Output.Id, 10))

	return resourceSshKeyRead(d, m)
}

func resourceSshKeyRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()

	id, err := strconv.Atoi(d.Id())

	if err != nil {
		return fmt.Errorf("Invalid SSH key id: %v", err)
	}

	user, err := getCurrentUser(api)

	if err != nil {
		return err
	}

	show := api.User.PublicKey.Show.Prepare()
	show.SetPathParamInt("user_id", user.Id)
	show.SetPathParamInt("public_key_id", int64(id))

	resp, err := show.Call()

	if err != nil {
		return err
	} else if !resp.Status {
		return fmt.Errorf("Failed to fetch SSH key: %s", resp.Message)
	}

	key := resp.Output

	d.Set("label", key.Label)
	d.Set("key", key.Key)
	d.Set("auto_add", key.AutoAdd)
	d.Set("fingerprint", key.Fingerprint)
	d.Set("comment", key.Comment)

	return nil
}

func resourceSshKeyUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()

	id, err := strconv.Atoi(d.Id())

	if err != nil {
		return fmt.Errorf("Invalid SSH key id: %v", err)
	}

	user, err := getCurrentUser(api)

	if err != nil {
		return err
	}

	update := api.User.PublicKey.Update.Prepare()
	update.SetPathParamInt("user_id", user.Id)
	update.SetPathParamInt("public_key_id", int64(id))

	input := update.NewInput()

	if d.HasChange("label") {
		input.SetLabel(d.Get("label").(string))
	}

	if d.HasChange("key") {
		input.SetKey(strings.TrimSpace(d.Get("key").(string)))
	}

	if d.HasChange("auto_add") {
		input.SetAutoAdd(d.Get("auto_add").(bool))
	}

	if input.AnySelected() {
		resp, err := update.Call()

		if err != nil {
			return err
		} else if !resp.Status {
			return fmt.Errorf("SSH key update failed: %s", resp.Message)
		}
	}

	return resourceSshKeyRead(d, m)
}

func resourceSshKeyDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()

	id, err := strconv.Atoi(d.Id())

	if err != nil {
		return fmt.Errorf("Invalid SSH key id: %v", err)
	}

	user, err := getCurrentUser(api)

	if err != nil {
		return err
	}

	log.Printf("[INFO] Deleting SSH key: %s", d.Id())

	del := api.User.PublicKey.Delete.Prepare()
	del.SetPathParamInt("user_id", user.Id)
	del.SetPathParamInt("public_key_id", int64(id))

	resp, err := del.Call()

	if err != nil {
		return err
	} else if !resp.Status {
		return fmt.Errorf("SSH key deletion failed: %s", resp.Message)
	}

	return nil
}

func resourceSshKeyImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	err := resourceSshKeyRead(d, m)

	if err != nil {
		return nil, fmt.Errorf("invalid SSH key id: %v", err)
	}

	results := make([]*schema.ResourceData, 1)
	results[0] = d

	return results, nil
}
