package vpsadmin

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/vpsfreecz/vpsadmin-go-client/client"
)

func TestGetPublicKeyByLabelUsesAllowedPageLimit(t *testing.T) {
	cfg := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v7.0/users/7/public_keys" {
			http.NotFound(w, r)
			return
		}

		assertQueryValue(t, r, "public_key[limit]", strconv.Itoa(apiPageLimit))
		writeAPIResponse(t, w, "public_keys", []*client.ActionUserPublicKeyIndexOutput{
			{Id: 1, Label: "provider-workflows"},
		})
	})

	key, err := getPublicKeyByLabel(cfg.getClient(), 7, "provider-workflows")
	if err != nil {
		t.Fatal(err)
	}
	if key.Id != 1 {
		t.Fatalf("key id = %d, want 1", key.Id)
	}
}

func TestGetOsTemplateIdByNameUsesAllowedPageLimit(t *testing.T) {
	cfg := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v7.0/os_templates" {
			http.NotFound(w, r)
			return
		}

		assertQueryValue(t, r, "os_template[limit]", strconv.Itoa(apiPageLimit))
		writeAPIResponse(t, w, "os_templates", []*client.ActionOsTemplateIndexOutput{
			{Id: 2, Name: "debian-latest-x86_64-vpsadminos-minimal"},
		})
	})

	id, err := getOsTemplateIdByName(cfg.getClient(), "debian-latest-x86_64-vpsadminos-minimal")
	if err != nil {
		t.Fatal(err)
	}
	if id != 2 {
		t.Fatalf("template id = %d, want 2", id)
	}
}

func TestGetDnsResolverIdByLabelUsesAllowedPageLimit(t *testing.T) {
	cfg := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v7.0/dns_resolvers" {
			http.NotFound(w, r)
			return
		}

		assertQueryValue(t, r, "dns_resolver[limit]", strconv.Itoa(apiPageLimit))
		writeAPIResponse(t, w, "dns_resolvers", []*client.ActionDnsResolverIndexOutput{
			{Id: 3, Label: "resolver-a"},
		})
	})

	id, err := getDnsResolverIdByLabel(cfg.getClient(), "resolver-a")
	if err != nil {
		t.Fatal(err)
	}
	if id != 3 {
		t.Fatalf("resolver id = %d, want 3", id)
	}
}

func TestMountFindByIdUsesAllowedVpsPageLimit(t *testing.T) {
	cfg := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v7.0/vpses":
			assertQueryValue(t, r, "vps[limit]", strconv.Itoa(apiPageLimit))
			writeAPIResponse(t, w, "vpses", []*client.ActionVpsIndexOutput{
				{Id: 123},
			})
		case "/v7.0/vpses/123/mounts/99":
			writeAPIResponse(t, w, "mount", &client.ActionVpsMountShowOutput{
				Id: 99,
			})
		default:
			http.NotFound(w, r)
		}
	})

	mount, err := mountFindById(cfg.getClient(), 99)
	if err != nil {
		t.Fatal(err)
	}
	if mount.Id != 99 {
		t.Fatalf("mount id = %d, want 99", mount.Id)
	}
}
