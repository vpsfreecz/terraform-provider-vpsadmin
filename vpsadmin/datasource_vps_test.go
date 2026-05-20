package vpsadmin

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vpsfreecz/vpsadmin-go-client/client"
)

func TestDataSourceVpsReadSetsComputedFields(t *testing.T) {
	cfg := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v7.0/vpses/123":
			if got := r.URL.Query().Get("_meta[includes]"); got != "node__location,os_template,dns_resolver" {
				t.Errorf("includes query = %q", got)
			}
			writeAPIResponse(t, w, "vps", &client.ActionVpsShowOutput{
				Id:                         123,
				Hostname:                   "app01",
				ManageHostname:             true,
				Cpu:                        4,
				Memory:                     8192,
				Swap:                       1024,
				StartMenuTimeout:           15,
				Dataset:                    &client.ActionDatasetShowOutput{Id: 456},
				DnsResolver:                &client.ActionDnsResolverShowOutput{Label: "resolver-a"},
				Node:                       testNode("node-a", "prg"),
				OsTemplate:                 &client.ActionOsTemplateShowOutput{Name: "debian-12"},
				EnableOsTemplateAutoUpdate: true,
			})
		case "/v7.0/datasets/456":
			writeAPIResponse(t, w, "dataset", &client.ActionDatasetShowOutput{
				Id:       456,
				Refquota: 40960,
			})
		case "/v7.0/vpses/123/features":
			writeAPIResponse(t, w, "features", []*client.ActionVpsFeatureIndexOutput{
				{Name: "fuse", Enabled: true},
				{Name: "kvm", Enabled: false},
				{Name: "nfs", Enabled: true},
			})
		case "/v7.0/host_ip_addresses":
			writeAPIResponse(t, w, "host_ip_addresses", hostIPAddressesForQuery(r))
		default:
			http.NotFound(w, r)
		}
	})

	d := schema.TestResourceDataRaw(t, dataSourceVps().Schema, map[string]interface{}{
		"vps_id": 123,
	})

	if err := dataSourceVpsRead(d, cfg); err != nil {
		t.Fatal(err)
	}

	assertResourceValue(t, d, "location", "prg")
	assertResourceValue(t, d, "node", "node-a")
	assertResourceValue(t, d, "os_template", "debian-12")
	assertResourceValue(t, d, "hostname", "app01")
	assertResourceValue(t, d, "manage_hostname", true)
	assertResourceValue(t, d, "dns_resolver", "resolver-a")
	assertResourceValue(t, d, "cpu", 4)
	assertResourceValue(t, d, "memory", 8192)
	assertResourceValue(t, d, "swap", 1024)
	assertResourceValue(t, d, "diskspace", 40960)
	assertResourceValue(t, d, "public_ipv4_address", "198.51.100.10")
	assertResourceValue(t, d, "private_ipv4_address", "10.0.0.10")
	assertResourceValue(t, d, "public_ipv6_address", "2001:db8::10")
	assertResourceValue(t, d, "feature_fuse", true)
	assertResourceValue(t, d, "feature_kvm", false)
	assertResourceValue(t, d, "feature_lxc", false)
	assertResourceValue(t, d, "start_menu_timeout", 15)
}

func TestDataSourceVpsReadClearsDnsResolver(t *testing.T) {
	cfg := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v7.0/vpses/123":
			writeAPIResponse(t, w, "vps", &client.ActionVpsShowOutput{
				Id:             123,
				Hostname:       "app01",
				ManageHostname: false,
				Cpu:            2,
				Memory:         2048,
				Dataset:        &client.ActionDatasetShowOutput{Id: 456},
				Node:           testNode("node-a", "prg"),
				OsTemplate:     &client.ActionOsTemplateShowOutput{Name: "debian-12"},
			})
		case "/v7.0/datasets/456":
			writeAPIResponse(t, w, "dataset", &client.ActionDatasetShowOutput{Id: 456})
		case "/v7.0/vpses/123/features":
			writeAPIResponse(t, w, "features", []*client.ActionVpsFeatureIndexOutput{})
		case "/v7.0/host_ip_addresses":
			writeAPIResponse(t, w, "host_ip_addresses", []*client.ActionHostIpAddressIndexOutput{})
		default:
			http.NotFound(w, r)
		}
	})

	d := schema.TestResourceDataRaw(t, dataSourceVps().Schema, map[string]interface{}{
		"vps_id": 123,
	})
	if err := d.Set("dns_resolver", "stale"); err != nil {
		t.Fatal(err)
	}

	if err := dataSourceVpsRead(d, cfg); err != nil {
		t.Fatal(err)
	}

	if _, ok := d.GetOk("dns_resolver"); ok {
		t.Fatalf("dns_resolver = %#v, want unset", d.Get("dns_resolver"))
	}
}

func testNode(domainName string, location string) *client.ActionNodeShowOutput {
	return &client.ActionNodeShowOutput{
		DomainName: domainName,
		Location:   &client.ActionLocationShowOutput{Label: location},
	}
}

func hostIPAddressesForQuery(r *http.Request) []*client.ActionHostIpAddressIndexOutput {
	q := r.URL.Query()
	if q.Get("host_ip_address[vps]") != "123" ||
		q.Get("host_ip_address[assigned]") != "1" ||
		q.Get("host_ip_address[limit]") != "1" {
		return nil
	}

	switch {
	case q.Get("host_ip_address[version]") == "4" &&
		q.Get("host_ip_address[role]") == "public_access":
		return []*client.ActionHostIpAddressIndexOutput{{Addr: "198.51.100.10"}}
	case q.Get("host_ip_address[version]") == "4" &&
		q.Get("host_ip_address[role]") == "private_access":
		return []*client.ActionHostIpAddressIndexOutput{{Addr: "10.0.0.10"}}
	case q.Get("host_ip_address[version]") == "6" &&
		q.Get("host_ip_address[role]") == "public_access":
		return []*client.ActionHostIpAddressIndexOutput{{Addr: "2001:db8::10"}}
	default:
		return nil
	}
}

func assertResourceValue(t *testing.T, d *schema.ResourceData, key string, want interface{}) {
	t.Helper()

	if got := d.Get(key); got != want {
		t.Fatalf("%s = %#v, want %#v", key, got, want)
	}
}
