package vpsadmin

import (
	"github.com/hashicorp/terraform/helper/schema"
	"fmt"
	"strconv"
)

func dataSourceVps() *schema.Resource {
	return &schema.Resource{
		Read:   dataSourceVpsRead,

		Schema: map[string]*schema.Schema{
			"vps_id": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "VPS ID",
				Required:    true,
			},
			"location": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Location label",
				Computed:    true,
			},
			"node": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Read-only node name",
				Computed:    true,
			},
			"os_template": &schema.Schema{
				Type:        schema.TypeString,
				Description: "OS template to base this VPS on",
				Computed:    true,
			},
			"hostname": &schema.Schema{
				Type:        schema.TypeString,
				Description: "VPS hostname managed by vpsAdmin",
				Computed:    true,
			},
			"cpu": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Number of CPU cores",
				Computed:    true,
			},
			"memory": &schema.Schema{
				Type   :     schema.TypeInt,
				Description: "Available memory in MB",
				Computed:    true,
			},
			"swap": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Available swap in MB",
				Computed:    true,
			},
			"diskspace": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Root dataset's size in MB",
				Computed:    true,
			},
			"public_ipv4_address": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Primary public IPv4 address",
				Computed:    true,
			},
			"private_ipv4_address": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Primary private IPv4 address",
				Computed:    true,
			},
			"public_ipv6_address": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Primary public IPv6 address",
				Computed:    true,
			},
		},
	}
}

func dataSourceVpsRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*Config).getClient()
	id := d.Get("vps_id").(int)
	vps, err := vpsShow(api, id)

	if err != nil {
		return fmt.Errorf("Invalid VPS ID: %v", err)
	}

	// Dataset cannot be prefetched, API limitation
	ds, err := datasetShow(api, int(vps.Dataset.Id))

	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(id))
	d.Set("location", vps.Node.Location.Label)
	d.Set("node", vps.Node.DomainName)
	d.Set("os_template", vps.OsTemplate.Name)
	d.Set("hostname", vps.Hostname)
	d.Set("cpu", vps.Cpu)
	d.Set("memory", vps.Memory)
	d.Set("swap", vps.Swap)
	d.Set("diskspace", ds.Refquota)
	d.Set("public_ipv4_address", getPrimaryPublicHostIpv4Address(api, vps.Id))
	d.Set("private_ipv4_address", getPrimaryPrivateHostIpv4Address(api, vps.Id))
	d.Set("public_ipv6_address", getPrimaryPublicHostIpv6Address(api, vps.Id))

	return nil
}
