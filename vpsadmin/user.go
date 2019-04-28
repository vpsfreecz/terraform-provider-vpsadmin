package vpsadmin

import (
	"errors"
	"github.com/vpsfreecz/vpsadmin-go-client/client"
)

func getCurrentUser(api *client.Client) (*client.ActionUserCurrentOutput, error) {
	resp, err := api.User.Current.Call()

	if err != nil {
		return nil, err
	} else if !resp.Status {
		return nil, errors.New(resp.Message)
	}

	return resp.Output, nil
}
