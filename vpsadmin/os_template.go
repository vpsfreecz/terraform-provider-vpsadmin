package vpsadmin

import (
	"fmt"
	"github.com/vpsfreecz/vpsadmin-go-client/client"
)

func getOsTemplateIdByName(api *client.Client, name string) (int64, error) {
	list := api.OsTemplate.Index.Prepare()
	list.SetInput(&client.ActionOsTemplateIndexInput{
		Limit: 10000,
	})
	list.Input.SelectParameters("Limit")
	resp, err := list.Call()

	if err != nil {
		return 0, err
	} else if !resp.Status {
		return 0, fmt.Errorf("Failed to list OS templates: %s", resp.Message)
	}

	for _, tpl := range resp.Output {
		if tpl.Name == name {
			return tpl.Id, nil
		}
	}

	return 0, fmt.Errorf("OS template with name '%s' not found", name)
}
