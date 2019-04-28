package vpsadmin

import (
	"github.com/hashicorp/terraform/helper/schema"
	"fmt"
	"strconv"
)

func dataSourceSshKey() *schema.Resource {
	return &schema.Resource{
		Read:   dataSourceSshKeyRead,

		Schema: map[string]*schema.Schema{
			"label": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Public key label",
				Required:    true,
			},
			"key": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Public key",
				Computed:    true,
			},
			"auto_add": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Automatically add this key to new VPS",
				Computed:    true,
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
func dataSourceSshKeyRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()
	label := d.Get("label").(string)

	user, err := getCurrentUser(api)

	if err != nil {
		return err
	}

	key, err := getPublicKeyByLabel(api, user.Id, label)

	if err != nil {
		return fmt.Errorf("Invalid key label: %v", err)
	}

	d.SetId(strconv.FormatInt(key.Id, 10))
	d.Set("label", key.Label)
	d.Set("key", key.Key)
	d.Set("auto_add", key.AutoAdd)
	d.Set("fingerprint", key.Fingerprint)
	d.Set("comment", key.Comment)

	return nil
}
