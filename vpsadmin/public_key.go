package vpsadmin

import (
	"fmt"
	"github.com/vpsfreecz/vpsadmin-go-client/client"
)

func getPublicKeyByLabel(api *client.Client, userId int64, label string) (*client.ActionUserPublicKeyIndexOutput, error) {
	list := api.User.PublicKey.Index.Prepare()
	list.SetPathParamInt("user_id", userId)

	input := list.NewInput()
	input.SetLimit(10000)

	resp, err := list.Call()

	if err != nil {
		return nil, err
	} else if !resp.Status {
		return nil, fmt.Errorf("Failed to list public keys: %s", resp.Message)
	}

	for _, key := range resp.Output {
		if key.Label == label {
			return key, nil
		}
	}

	return nil, fmt.Errorf("SSH key with label '%s' not found", label)
}
