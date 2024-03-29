package vpsadmin

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vpsfreecz/vpsadmin-go-client/client"
	"log"
	"strconv"
)

func resourceVps() *schema.Resource {
	return &schema.Resource{
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceVpsSchemaV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceVpsUpgradeV0,
				Version: 0,
			},
		},
		Create: resourceVpsCreate,
		Read:   resourceVpsRead,
		Update: resourceVpsUpdate,
		Delete: resourceVpsDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVpsImport,
		},

		Description: `
Represents a virtual server instance. To create a VPS, you need to have
a sufficient amount of resources assigned to your account in vpsAdmin. Contact
support in case you need more.
`,

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
			"installed_os_template": &schema.Schema{
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
			"dns_resolver": &schema.Schema{
				Type:          schema.TypeString,
				Description:   "DNS resolver used by the VPS if managed by vpsAdmin",
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"manage_dns_resolver"},
			},
			"manage_dns_resolver": &schema.Schema{
				Type:          schema.TypeBool,
				Description:   "Manage DNS resolver by vpsAdmin if true, manually if false",
				Default:       true,
				Optional:      true,
				ConflictsWith: []string{"dns_resolver"},
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
			"feature_fuse": {
				Type:        schema.TypeBool,
				Description: "Allow access to FUSE filesystems",
				Default:     true,
				Optional:    true,
			},
			"feature_kvm": {
				Type:        schema.TypeBool,
				Description: "Allow access to /dev/kvm for hardware virtualization",
				Default:     true,
				Optional:    true,
			},
			"feature_lxc": {
				Type:        schema.TypeBool,
				Description: "Enable support for LXC/LXD containers",
				Default:     false,
				Optional:    true,
			},
			"feature_ppp": {
				Type:        schema.TypeBool,
				Description: "Allow access to /dev/ppp",
				Default:     false,
				Optional:    true,
			},
			"feature_tun": {
				Type:        schema.TypeBool,
				Description: "Allow access to /dev/net/tun, e.g. for VPNs",
				Default:     true,
				Optional:    true,
			},
			"start_menu_timeout": {
				Type:        schema.TypeInt,
				Description: "Start menu timeout before the VPS is started, in seconds",
				Computed:    true,
				Optional:    true,
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

	templateId, err := getOsTemplateIdByName(api, d.Get("install_os_template").(string))
	if err != nil {
		return err
	}

	manageHostname := d.Get("manage_hostname").(bool)
	manageResolver := d.Get("manage_dns_resolver").(bool)

	create := api.Vps.Create.Prepare()

	input := create.NewInput()
	input.SetLocation(locationId)
	input.SetOsTemplate(templateId)

	if manageHostname {
		input.SetHostname(d.Get("hostname").(string))
	} else {
		input.SetHostname("vps")
	}

	if manageResolver {
		if v, ok := d.GetOk("dns_resolver"); ok {
			resolverId, err := getDnsResolverIdByLabel(api, v.(string))
			if err != nil {
				return err
			}

			input.SetDnsResolver(resolverId)
		}
	}

	input.SetCpu(int64(d.Get("cpu").(int)))
	input.SetMemory(int64(d.Get("memory").(int)))
	input.SetSwap(int64(d.Get("swap").(int)))
	input.SetDiskspace(int64(d.Get("diskspace").(int)))
	input.SetIpv4(int64(d.Get("public_ipv4_count").(int)))
	input.SetIpv4Private(int64(d.Get("private_ipv4_count").(int)))
	input.SetIpv6(int64(d.Get("public_ipv6_count").(int)))

	if v, ok := d.GetOk("start_menu_timeout"); ok {
		input.SetStartMenuTimeout(int64(v.(int)))
	}

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

	// Hostname
	if !manageHostname {
		update := api.Vps.Update.Prepare()
		update.SetPathParamInt("vps_id", resp.Output.Id)
		updateInput := update.NewInput()
		updateInput.SetManageHostname(false)

		log.Printf("[DEBUG] Configuring manual hostname management")

		updateResp, err := update.Call()

		if err != nil {
			return err
		} else if !updateResp.Status {
			return fmt.Errorf("VPS update failed: %s", updateResp.Message)
		}

		if err := waitForOperation(updateResp); err != nil {
			return fmt.Errorf("VPS update failed: %v", err)
		}
	}

	// VPS features
	featureSet := api.Vps.Feature.UpdateAll.Prepare()
	featureSet.SetPathParamInt("vps_id", resp.Output.Id)
	featureInput := featureSet.NewInput()

	for _, name := range supportedVpsFeatures {
		enabled := d.Get(fmt.Sprintf("feature_%s", name)).(bool)

		if name == "fuse" {
			featureInput.SetFuse(enabled)
		} else if name == "kvm" {
			featureInput.SetKvm(enabled)
		} else if name == "lxc" {
			featureInput.SetLxc(enabled)
		} else if name == "ppp" {
			featureInput.SetPpp(enabled)
		} else if name == "tun" {
			featureInput.SetTun(enabled)
		}
	}

	featureResp, err := featureSet.Call()
	if err != nil {
		return err
	} else if !featureResp.Status {
		return fmt.Errorf("VPS feature set failed: %s", featureResp.Message)
	}

	if err := waitForOperation(featureResp); err != nil {
		return fmt.Errorf("VPS feature set failed: %v", err)
	}

	// SSH keys
	if keys, ok := d.GetOk("ssh_keys"); ok {
		if err := deploySshKeys(api, resp.Output.Id, keys.(*schema.Set).List()); err != nil {
			return err
		}
	}

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

	publicIpv4 := getPrimaryPublicHostIpv4Address(api, vps.Id)
	privateIpv4 := getPrimaryPrivateHostIpv4Address(api, vps.Id)
	publicIpv6 := getPrimaryPublicHostIpv6Address(api, vps.Id)

	features, err := vpsFeatureList(api, id)
	if err != nil {
		return err
	}

	d.Set("location", vps.Node.Location.Label)
	d.Set("node", vps.Node.DomainName)
	d.Set("installed_os_template", vps.OsTemplate.Name)
	d.Set("hostname", vps.Hostname)
	d.Set("real_hostname", vps.Hostname)
	d.Set("manage_hostname", vps.ManageHostname)

	if vps.DnsResolver != nil {
		d.Set("dns_resolver", vps.DnsResolver.Label)
	} else {
		d.Set("dns_resolver", nil)
	}

	d.Set("manage_dns_resolver", vps.DnsResolver != nil)
	d.Set("cpu", vps.Cpu)
	d.Set("memory", vps.Memory)
	d.Set("swap", vps.Swap)
	d.Set("diskspace", ds.Refquota)
	d.Set("public_ipv4_address", publicIpv4)
	d.Set("private_ipv4_address", privateIpv4)
	d.Set("public_ipv6_address", publicIpv6)

	if addr := getPrimaryConnectionAddress(publicIpv4, privateIpv4, publicIpv6); addr != "" {
		log.Printf("[INFO] Setting connection host to: '%s'", addr)
		d.SetConnInfo(map[string]string{
			"type": "ssh",
			"host": addr,
		})
	} else {
		log.Printf("[INFO] No connection host found")
	}

	for _, feature := range features {
		if isSupportedVpsFeature(feature.Name) {
			d.Set(fmt.Sprintf("feature_%s", feature.Name), feature.Enabled)
		}
	}

	d.Set("start_menu_timeout", vps.StartMenuTimeout)

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

	input := vpsUpdate.NewInput()

	if d.HasChange("installed_os_template") {
		templateId, err := getOsTemplateIdByName(api, d.Get("installed_os_template").(string))
		if err != nil {
			return err
		}
		input.SetOsTemplate(templateId)
	}

	if d.HasChange("hostname") {
		input.SetHostname(d.Get("hostname").(string))
	}

	if d.HasChange("manage_hostname") {
		input.SetManageHostname(d.Get("manage_hostname").(bool))
	}

	if d.HasChanges("dns_resolver", "manage_dns_resolver") {
		manageResolver := d.Get("manage_dns_resolver").(bool)

		if manageResolver {
			resolverId, err := getDnsResolverIdByLabel(api, d.Get("dns_resolver").(string))
			if err != nil {
				return err
			}
			input.SetDnsResolver(resolverId)
		} else {
			input.SetDnsResolverNil(true)
		}
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

	if d.HasChange("start_menu_timeout") {
		input.SetStartMenuTimeout(int64(d.Get("start_menu_timeout").(int)))
	}

	if input.AnySelected() {
		vpsResp, err := vpsUpdate.Call()

		if err != nil {
			return err
		} else if !vpsResp.Status {
			return fmt.Errorf("VPS update failed: %s", vpsResp.Message)
		}

		if err := waitForOperation(vpsResp); err != nil {
			return fmt.Errorf("VPS update failed: %v", err)
		}
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
	}

	if hasAnyVpsFeatureChange(d) {
		featureSet := api.Vps.Feature.UpdateAll.Prepare()
		featureSet.SetPathParamInt("vps_id", int64(id))
		featureInput := featureSet.NewInput()

		for _, name := range supportedVpsFeatures {
			enabled := d.Get(fmt.Sprintf("feature_%s", name)).(bool)

			if name == "fuse" {
				featureInput.SetFuse(enabled)
			} else if name == "kvm" {
				featureInput.SetKvm(enabled)
			} else if name == "lxc" {
				featureInput.SetLxc(enabled)
			} else if name == "ppp" {
				featureInput.SetPpp(enabled)
			} else if name == "tun" {
				featureInput.SetTun(enabled)
			}
		}

		featureResp, err := featureSet.Call()
		if err != nil {
			return err
		} else if !featureResp.Status {
			return fmt.Errorf("VPS feature set failed: %s", featureResp.Message)
		}

		if err := waitForOperation(featureResp); err != nil {
			return fmt.Errorf("VPS feature set failed: %v", err)
		}
	}

	if d.HasChange("ssh_keys") {
		if err := deploySshKeys(api, int64(id), d.Get("ssh_keys").(*schema.Set).List()); err != nil {
			return err
		}
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

func resourceVpsImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	err := resourceVpsRead(d, m)

	if err != nil {
		return nil, fmt.Errorf("invalid VPS id: %v", err)
	}

	d.Set("install_os_template", d.Get("installed_os_template"))

	results := make([]*schema.ResourceData, 1)
	results[0] = d

	return results, nil
}

func resourceVpsUpgradeV0(ctx context.Context,
	rawState map[string]interface{},
	m interface{}) (map[string]interface{}, error) {
	rawState["install_os_template"] = rawState["os_template"]
	delete(rawState, "os_template")
	return rawState, nil
}

func hasAnyVpsFeatureChange(d *schema.ResourceData) bool {
	for _, name := range supportedVpsFeatures {
		if d.HasChange(fmt.Sprintf("feature_%s", name)) {
			return true
		}
	}

	return false
}

func deploySshKeys(api *client.Client, vpsId int64, sshKeys []interface{}) error {
	for _, v := range sshKeys {
		keyId := v.(string)

		deploy := api.Vps.DeployPublicKey.Prepare()
		deploy.SetPathParamInt("vps_id", vpsId)

		input := deploy.NewInput()

		n, err := strconv.ParseInt(keyId, 10, 64)

		if err != nil {
			return err
		}

		input.SetPublicKey(n)

		log.Printf("[INFO] Deploying SSH key %d to VPS %d", n, vpsId)

		resp, err := deploy.Call()

		if err != nil {
			return err
		} else if !resp.Status {
			return fmt.Errorf("SSH key deploy failed: %s", resp.Message)
		}

		if err := waitForOperation(resp); err != nil {
			return fmt.Errorf("SSH key deploy failed: %v", err)
		}
	}

	return nil
}
