package vpsadmin

import (
	"fmt"
	"github.com/vpsfreecz/vpsadmin-go-client/client"
)

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
