package vpsadmin

import (
	"fmt"
	"github.com/vpsfreecz/vpsadmin-go-client/client"
)

func datasetShow(api *client.Client, id int) (*client.ActionDatasetShowOutput, error) {
	show := api.Dataset.Show.Prepare()
	show.SetPathParamInt("dataset_id", int64(id))

	resp, err := show.Call()

	if err != nil {
		return nil, err
	} else if !resp.Status {
		return nil, fmt.Errorf("Dataset show failed: %s", resp.Message)
	}

	if resp.Output.Export != nil {
		// Due to an API limitation, we cannot prefetch the export
		exportShow := api.Export.Show.Prepare()
		exportShow.SetPathParamInt("export_id", resp.Output.Export.Id)
		exportShow.SetMetaInput(&client.ActionExportShowMetaGlobalInput{
			Includes: "host_ip_address",
		})
		exportShow.MetaInput.SelectParameters("Includes")

		exportResp, err := exportShow.Call()

		if err != nil {
			return nil, err
		} else if !exportResp.Status {
			return nil, fmt.Errorf("Export show failed: %s", exportResp.Message)
		}

		resp.Output.Export = exportResp.Output
	}

	return resp.Output, nil
}
