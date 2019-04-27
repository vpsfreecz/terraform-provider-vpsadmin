package vpsadmin

import (
	"fmt"
	"github.com/vpsfreecz/vpsadmin-go-client/client"
)

func getLocationIdByLabel(api *client.Client, label string) (int64, error) {
	list := api.Location.List.Prepare()
	resp, err := list.Call()

	if err != nil {
		return 0, err
	} else if !resp.Status {
		return 0, fmt.Errorf("Failed to list locations: %s", resp.Message)
	}

	for _, location := range resp.Output {
		if location.Label == label {
			return location.Id, nil
		}
	}

	return 0, fmt.Errorf("Location with label '%s' not found", label)
}
