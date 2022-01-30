package vpsadmin

import (
	"github.com/vpsfreecz/vpsadmin-go-client/client"
	"log"
)

func getPrimaryPublicHostIpv4Address(api *client.Client, vpsId int64) string {
	return getPrimaryHostIpAddress(api, vpsId, 4, "public_access")
}

func getPrimaryPrivateHostIpv4Address(api *client.Client, vpsId int64) string {
	return getPrimaryHostIpAddress(api, vpsId, 4, "private_access")
}

func getPrimaryPublicHostIpv6Address(api *client.Client, vpsId int64) string {
	return getPrimaryHostIpAddress(api, vpsId, 6, "public_access")
}

func getPrimaryHostIpAddress(api *client.Client, vpsId int64, ipVersion int, role string) string {
	action := api.HostIpAddress.Index.Prepare()

	input := action.NewInput()
	input.SetVps(vpsId)
	input.SetVersion(int64(ipVersion))
	input.SetRole(role)
	input.SetAssigned(true)
	input.SetLimit(1)

	log.Printf("[DEBUG] Listing host IP addresses: %+v", input)

	resp, err := action.Call()

	if err != nil {
		log.Printf("[INFO] Failed to list host IP addresses: %v", err)
		return ""
	} else if !resp.Status {
		log.Printf("[INFO] Failed to list host IP addresses: %s", resp.Message)
		return ""
	} else if len(resp.Output) == 0 {
		return ""
	}

	return resp.Output[0].Addr
}

func getPrimaryConnectionAddress(publicIpv4, privateIpv4, publicIpv6 string) string {
	if publicIpv4 != "" {
		return publicIpv4
	} else if publicIpv6 != "" {
		return publicIpv6
	} else {
		return privateIpv4
	}
}
