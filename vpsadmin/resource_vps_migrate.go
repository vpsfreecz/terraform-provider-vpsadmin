package vpsadmin

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceVpsSchemaV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"location": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Location label",
				Required:    true,
				ForceNew:    true,
			},
			"node": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Read-only node name",
				Computed:    true,
			},
			"install_os_template": &schema.Schema{
				Type:        schema.TypeString,
				Description: "OS template which is installed to the VPS",
				Required:    true,
				ForceNew:    true,
			},
			"use_os_template": &schema.Schema{
				Type:        schema.TypeString,
				Description: "OS template which corresponds to the VPS at the moment",
				Computed:    true,
				Optional:    true,
			},
			"hostname": &schema.Schema{
				Type:          schema.TypeString,
				Description:   "VPS hostname managed by vpsAdmin",
				Optional:      true,
				Default:       "vps",
				ConflictsWith: []string{"manage_hostname"},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return !d.Get("manage_hostname").(bool)
				},
			},
			"real_hostname": &schema.Schema{
				Type:        schema.TypeString,
				Description: "VPS hostname as reported by the VPS",
				Computed:    true,
			},
			"manage_hostname": &schema.Schema{
				Type:          schema.TypeBool,
				Description:   "Manage hostname by vpsAdmin if true, manually if false",
				Default:       true,
				Optional:      true,
				ConflictsWith: []string{"hostname"},
			},
			"cpu": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Number of CPU cores",
				Required:    true,
			},
			"memory": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Available memory in MB",
				Required:    true,
			},
			"swap": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Available swap in MB",
				Default:     0,
				Optional:    true,
			},
			"diskspace": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Root dataset's size in MB",
				Required:    true,
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
			"public_ipv4_count": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Number of public IPv4 addresses to add when the VPS is created",
				Default:     1,
				Optional:    true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// This field is used only when the VPS is being created
					return d.Id() != ""
				},
			},
			"private_ipv4_count": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Number of private IPv4 addresses to add when the VPS is created",
				Default:     0,
				Optional:    true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// This field is used only when the VPS is being created
					return d.Id() != ""
				},
			},
			"public_ipv6_count": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Number of public IPv6 addresses to add when the VPS is created",
				Default:     1,
				Optional:    true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// This field is used only when the VPS is being created
					return d.Id() != ""
				},
			},
			"ssh_keys": &schema.Schema{
				Type:        schema.TypeSet,
				Description: "List of SSH key IDs to append to /root/.ssh_authorized_keys",
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.NoZeroValues,
				},
			},
		},
	}
}
