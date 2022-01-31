package vpsadmin

import (
	"fmt"
	"github.com/vpsfreecz/vpsadmin-go-client/client"
)

func mountShow(api *client.Client, vpsId int, mountId int) (*client.ActionVpsMountShowOutput, error) {
	show := api.Vps.Mount.Show.Prepare()
	show.SetPathParamInt("vps_id", int64(vpsId))
	show.SetPathParamInt("mount_id", int64(mountId))

	resp, err := show.Call()

	if err != nil {
		return nil, err
	} else if !resp.Status {
		return nil, fmt.Errorf("Mount show failed: %s", resp.Message)
	}

	return resp.Output, nil
}

func mountFindById(api *client.Client, id int) (*client.ActionVpsMountShowOutput, error) {
	vpsList := api.Vps.Index.Prepare()
	vpsList.SetInput(&client.ActionVpsIndexInput{
		Limit: 10000,
	})
	vpsList.Input.SelectParameters("Limit")

	vpsResp, err := vpsList.Call()
	if err != nil {
		return nil, err
	} else if !vpsResp.Status {
		return nil, fmt.Errorf("Mount lookup failed, unable to list VPS: %s", vpsResp.Message)
	}

	for _, vps := range vpsResp.Output {
		mountShow := api.Vps.Mount.Show.Prepare()
		mountShow.SetPathParamInt("vps_id", vps.Id)
		mountShow.SetPathParamInt("mount_id", int64(id))

		mountResp, err := mountShow.Call()
		if err != nil {
			return nil, err
		} else if !mountResp.Status {
			continue
		}

		return mountResp.Output, nil
	}

	return nil, fmt.Errorf("Mount not found on any VPS")
}
