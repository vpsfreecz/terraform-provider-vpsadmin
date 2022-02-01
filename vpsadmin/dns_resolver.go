package vpsadmin

import (
	"fmt"
	"github.com/vpsfreecz/vpsadmin-go-client/client"
)

func getDnsResolverIdByLabel(api *client.Client, label string) (int64, error) {
	list := api.DnsResolver.Index.Prepare()
	list.SetInput(&client.ActionDnsResolverIndexInput{
		Limit: 10000,
	})
	list.Input.SelectParameters("Limit")
	resp, err := list.Call()

	if err != nil {
		return 0, err
	} else if !resp.Status {
		return 0, fmt.Errorf("Failed to list DNS resolvers: %s", resp.Message)
	}

	for _, resolver := range resp.Output {
		if resolver.Label == label {
			return resolver.Id, nil
		}
	}

	return 0, fmt.Errorf("DNS resolver with label '%s' not found", label)
}
