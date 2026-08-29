package vpsadmin

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceVpsUserData() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVpsUserDataRead,

		Description: "Reads user data stored in vpsAdmin by its numeric ID. Raw content is not stored in Terraform state.",

		Schema: map[string]*schema.Schema{
			"vps_user_data_id": {
				Type:         schema.TypeInt,
				Description:  "ID of the stored user data",
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"label": {
				Type:        schema.TypeString,
				Description: "Label of the stored user data",
				Computed:    true,
			},
			"format": {
				Type:        schema.TypeString,
				Description: "Content format",
				Computed:    true,
			},
			"content_sha256": {
				Type:        schema.TypeString,
				Description: "SHA-256 digest of the stored content",
				Computed:    true,
			},
		},
	}
}

func dataSourceVpsUserDataRead(d *schema.ResourceData, m interface{}) error {
	id := int64(d.Get("vps_user_data_id").(int))
	resp, err := vpsUserDataShow(m.(*Config).getClient(), id)
	if err != nil {
		return err
	} else if !resp.Status {
		return fmt.Errorf("failed to fetch vpsAdmin user data: %s", resp.Message)
	}

	d.SetId(strconv.FormatInt(resp.Output.Id, 10))
	for name, value := range map[string]interface{}{
		"label":          resp.Output.Label,
		"format":         resp.Output.Format,
		"content_sha256": userDataStateHash(resp.Output.Content),
	} {
		if err := d.Set(name, value); err != nil {
			return fmt.Errorf("failed to store vpsAdmin user data field %q: %v", name, err)
		}
	}

	return nil
}
