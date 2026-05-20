package vpsadmin

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vpsfreecz/vpsadmin-go-client/client"
)

func TestDataSourceDatasetReadMapsExportFields(t *testing.T) {
	cfg := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v7.0/datasets/find_by_name" {
			http.NotFound(w, r)
			return
		}

		if got := r.URL.Query().Get("dataset[name]"); got != "tank/app" {
			t.Errorf("dataset[name] = %q", got)
		}

		writeAPIResponse(t, w, "dataset", &client.ActionDatasetFindByNameOutput{
			Id:          77,
			Name:        "tank/app",
			Used:        10,
			Referenced:  20,
			Avail:       30,
			Quota:       40,
			Refquota:    50,
			Compression: true,
			Recordsize:  131072,
			Atime:       false,
			Relatime:    true,
			Sync:        "standard",
			Export:      testExport(88),
		})
	})

	d := schema.TestResourceDataRaw(t, dataSourceDataset().Schema, map[string]interface{}{
		"name": "tank/app",
	})

	if err := dataSourceDatasetRead(d, cfg); err != nil {
		t.Fatal(err)
	}

	if d.Id() != "77" {
		t.Fatalf("id = %q, want 77", d.Id())
	}
	assertResourceValue(t, d, "full_name", "tank/app")
	assertResourceValue(t, d, "used", 10)
	assertResourceValue(t, d, "referenced", 20)
	assertResourceValue(t, d, "avail", 30)
	assertResourceValue(t, d, "quota", 40)
	assertResourceValue(t, d, "refquota", 50)
	assertResourceValue(t, d, "compression", true)
	assertResourceValue(t, d, "recordsize", 131072)
	assertResourceValue(t, d, "atime", false)
	assertResourceValue(t, d, "relatime", true)
	assertResourceValue(t, d, "sync", "standard")
	assertResourceValue(t, d, "export_dataset", true)
	assertResourceValue(t, d, "export_id", 88)
	assertResourceValue(t, d, "export_enable", false)
	assertResourceValue(t, d, "export_root_squash", true)
	assertResourceValue(t, d, "export_read_write", false)
	assertResourceValue(t, d, "export_sync", false)
	assertResourceValue(t, d, "export_ip_address", "192.0.2.55")
	assertResourceValue(t, d, "export_path", "/exports/tank/app")
}

func TestResourceDatasetReadMapsFetchedExportFields(t *testing.T) {
	cfg := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v7.0/datasets/77":
			writeAPIResponse(t, w, "dataset", &client.ActionDatasetShowOutput{
				Id:     77,
				Name:   "tank/app",
				Export: &client.ActionExportShowOutput{Id: 88},
			})
		case "/v7.0/exports/88":
			if got := r.URL.Query().Get("_meta[includes]"); got != "host_ip_address" {
				t.Errorf("includes query = %q", got)
			}
			writeAPIResponse(t, w, "export", testExport(88))
		default:
			http.NotFound(w, r)
		}
	})

	d := schema.TestResourceDataRaw(t, resourceDataset().Schema, map[string]interface{}{
		"name": "app",
	})
	d.SetId("77")

	if err := resourceDatasetRead(d, cfg); err != nil {
		t.Fatal(err)
	}

	assertResourceValue(t, d, "export_dataset", true)
	assertResourceValue(t, d, "export_id", 88)
	assertResourceValue(t, d, "export_ip_address", "192.0.2.55")
	assertResourceValue(t, d, "export_path", "/exports/tank/app")
}

func TestResourceDatasetReadClearsMissingExport(t *testing.T) {
	cfg := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v7.0/datasets/77" {
			http.NotFound(w, r)
			return
		}

		writeAPIResponse(t, w, "dataset", &client.ActionDatasetShowOutput{
			Id:   77,
			Name: "tank/app",
		})
	})

	d := schema.TestResourceDataRaw(t, resourceDataset().Schema, map[string]interface{}{
		"name": "app",
	})
	d.SetId("77")
	if err := d.Set("export_dataset", true); err != nil {
		t.Fatal(err)
	}
	if err := d.Set("export_id", 88); err != nil {
		t.Fatal(err)
	}

	if err := resourceDatasetRead(d, cfg); err != nil {
		t.Fatal(err)
	}

	assertResourceValue(t, d, "export_dataset", false)
	if _, ok := d.GetOk("export_id"); ok {
		t.Fatalf("export_id = %#v, want unset", d.Get("export_id"))
	}
}

func testExport(id int64) *client.ActionExportShowOutput {
	return &client.ActionExportShowOutput{
		Id:            id,
		Enabled:       false,
		RootSquash:    true,
		Rw:            false,
		Sync:          false,
		Path:          "/exports/tank/app",
		HostIpAddress: &client.ActionHostIpAddressShowOutput{Addr: "192.0.2.55"},
	}
}
