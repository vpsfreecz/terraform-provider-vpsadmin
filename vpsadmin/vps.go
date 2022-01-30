package vpsadmin

import (
	"fmt"
	"github.com/vpsfreecz/vpsadmin-go-client/client"
)

var supportedVpsFeatures []string = []string{"fuse", "kvm", "lxc", "ppp", "tun"}

func vpsShow(api *client.Client, id int) (*client.ActionVpsShowOutput, error) {
	show := api.Vps.Show.Prepare()
	show.SetPathParamInt("vps_id", int64(id))
	show.SetMetaInput(&client.ActionVpsShowMetaGlobalInput{
		Includes: "node__location,os_template",
	})
	show.MetaInput.SelectParameters("Includes")

	resp, err := show.Call()

	if err != nil {
		return nil, err
	} else if !resp.Status {
		return nil, fmt.Errorf("VPS show failed: %s", resp.Message)
	}

	return resp.Output, nil
}

func vpsFeatureList(api *client.Client, id int) ([]*client.ActionVpsFeatureIndexOutput, error) {
	list := api.Vps.Feature.Index.Prepare()
	list.SetPathParamInt("vps_id", int64(id))

	resp, err := list.Call()

	if err != nil {
		return nil, err
	} else if !resp.Status {
		return nil, fmt.Errorf("VPS feature list failed: %s", resp.Message)
	}

	return resp.Output, nil
}

func isSupportedVpsFeature(feature string) bool {
	for _, name := range supportedVpsFeatures {
		if name == feature {
			return true
		}
	}

	return false
}
