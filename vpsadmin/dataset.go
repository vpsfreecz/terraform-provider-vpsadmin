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

	return resp.Output, nil
}
