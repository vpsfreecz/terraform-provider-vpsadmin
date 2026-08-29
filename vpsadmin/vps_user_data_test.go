package vpsadmin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vpsfreecz/vpsadmin-go-client/client"
)

func TestResourceVpsUserDataSchema(t *testing.T) {
	t.Parallel()

	resource := resourceVpsUserData()
	if resource.Importer == nil || resource.Importer.State == nil {
		t.Fatal("vpsAdmin user data has no importer")
	}

	content := resource.Schema["content"]
	if !content.Required || !content.Sensitive || content.StateFunc == nil {
		t.Fatalf("unexpected content schema: %#v", content)
	}
	format := resource.Schema["format"]
	if !format.Optional || format.Default != defaultUserDataFormat {
		t.Fatalf("unexpected format schema: %#v", format)
	}
	if _, errors := format.ValidateFunc("invalid", "format"); len(errors) == 0 {
		t.Error("format accepted an unsupported value")
	}
	if _, errors := resource.Schema["label"].ValidateFunc("", "label"); len(errors) == 0 {
		t.Error("label accepted an empty value")
	}

	const secretMarker = "stored-user-data-secret-marker"
	_, errors := content.ValidateFunc(
		strings.Repeat("x", maxUserDataBytes+1)+secretMarker,
		"content",
	)
	if len(errors) == 0 {
		t.Error("content accepted content over the API limit")
	}
	if strings.Contains(fmt.Sprint(errors), secretMarker) {
		t.Fatalf("content validation exposed content: %v", errors)
	}
}

func TestResourceVpsUserDataCreateSendsRawContentAndStoresHash(t *testing.T) {
	const (
		content = "#!/bin/sh\nprintf stored > /root/provider-stored-user-data\n"
		label   = "provider stored user data"
	)

	requestCount := 0
	cfg := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v7.0/vps_user_data":
			assertVpsUserDataRequest(t, r, map[string]interface{}{
				"content": content,
				"format":  "script",
				"label":   label,
			})
			writeAPIResponse(t, w, "vps_user_data", testVpsUserData(42, label, "script", content))
		case r.Method == http.MethodGet && r.URL.Path == "/v7.0/vps_user_data/42":
			writeAPIResponse(t, w, "vps_user_data", testVpsUserData(42, label, "script", content))
		default:
			http.NotFound(w, r)
		}
	})

	d := schema.TestResourceDataRaw(t, resourceVpsUserData().Schema, map[string]interface{}{
		"label":   label,
		"content": content,
	})
	if err := resourceVpsUserDataCreate(d, cfg); err != nil {
		t.Fatal(err)
	}
	if requestCount != 2 {
		t.Fatalf("request count = %d, want 2", requestCount)
	}
	if d.Id() != "42" {
		t.Fatalf("ID = %q, want 42", d.Id())
	}

	state := d.State().Attributes
	if state["content"] != userDataStateHash(content) {
		t.Fatalf("content state = %q, want SHA-256 digest", state["content"])
	}
	if strings.Contains(fmt.Sprintf("%#v", state), content) {
		t.Fatalf("state contains raw content: %#v", state)
	}
}

func TestResourceVpsUserDataUpdateSendsChangedFields(t *testing.T) {
	const (
		oldContent = "#!/bin/sh\nprintf old\n"
		newContent = "#cloud-config\nwrite_files: []\n"
	)

	requestCount := 0
	cfg := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/v7.0/vps_user_data/42":
			assertVpsUserDataRequest(t, r, map[string]interface{}{
				"content": newContent,
				"format":  "cloudinit_config",
				"label":   "new label",
			})
			writeAPIResponse(
				t,
				w,
				"vps_user_data",
				testVpsUserData(42, "new label", "cloudinit_config", newContent),
			)
		case r.Method == http.MethodGet && r.URL.Path == "/v7.0/vps_user_data/42":
			writeAPIResponse(
				t,
				w,
				"vps_user_data",
				testVpsUserData(42, "new label", "cloudinit_config", newContent),
			)
		default:
			http.NotFound(w, r)
		}
	})

	d := newResourceDataWithDiff(
		t,
		resourceVpsUserData().Schema,
		"42",
		map[string]string{
			"label":   "old label",
			"format":  "script",
			"content": userDataStateHash(oldContent),
		},
		map[string]*terraform.ResourceAttrDiff{
			"label": {
				Old: "old label",
				New: "new label",
			},
			"format": {
				Old: "script",
				New: "cloudinit_config",
			},
			"content": {
				Old:      userDataStateHash(oldContent),
				New:      userDataStateHash(newContent),
				NewExtra: newContent,
			},
		},
	)
	if err := resourceVpsUserDataUpdate(d, cfg); err != nil {
		t.Fatal(err)
	}
	if requestCount != 2 {
		t.Fatalf("request count = %d, want 2", requestCount)
	}
	if got := d.State().Attributes["content"]; got != userDataStateHash(newContent) {
		t.Fatalf("content state = %q, want updated SHA-256 digest", got)
	}
}

func TestResourceVpsUserDataReadTracksRemoteContent(t *testing.T) {
	const content = "#!/bin/sh\nprintf remote\n"
	cfg := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v7.0/vps_user_data/42" {
			http.NotFound(w, r)
			return
		}
		writeAPIResponse(t, w, "vps_user_data", testVpsUserData(42, "remote", "script", content))
	})

	d := schema.TestResourceDataRaw(t, resourceVpsUserData().Schema, map[string]interface{}{
		"label":   "configured",
		"content": "#!/bin/sh\nprintf configured\n",
	})
	d.SetId("42")
	if err := resourceVpsUserDataRead(d, cfg); err != nil {
		t.Fatal(err)
	}
	state := d.State().Attributes
	if state["label"] != "remote" || state["content"] != userDataStateHash(content) {
		t.Fatalf("unexpected refreshed state: %#v", state)
	}
}

func TestResourceVpsUserDataReadRemovesMissingResource(t *testing.T) {
	cfg := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		writeVpsUserDataNotFound(t, w)
	})
	d := schema.TestResourceDataRaw(t, resourceVpsUserData().Schema, map[string]interface{}{
		"label":   "missing",
		"content": "#!/bin/sh\ntrue\n",
	})
	d.SetId("42")

	if err := resourceVpsUserDataRead(d, cfg); err != nil {
		t.Fatal(err)
	}
	if d.Id() != "" {
		t.Fatalf("ID = %q, want empty ID", d.Id())
	}
}

func TestResourceVpsUserDataDeleteAcceptsMissingResource(t *testing.T) {
	cfg := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v7.0/vps_user_data/42" {
			http.NotFound(w, r)
			return
		}
		writeVpsUserDataNotFound(t, w)
	})
	d := schema.TestResourceDataRaw(t, resourceVpsUserData().Schema, nil)
	d.SetId("42")

	if err := resourceVpsUserDataDelete(d, cfg); err != nil {
		t.Fatal(err)
	}
	if d.Id() != "" {
		t.Fatalf("ID = %q, want empty ID", d.Id())
	}
}

func TestResourceVpsUserDataImport(t *testing.T) {
	const content = "#!/bin/sh\ntrue\n"
	cfg := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		writeAPIResponse(t, w, "vps_user_data", testVpsUserData(42, "imported", "script", content))
	})
	d := schema.TestResourceDataRaw(t, resourceVpsUserData().Schema, nil)
	d.SetId("42")

	resources, err := resourceVpsUserDataImport(d, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(resources) != 1 || resources[0].Id() != "42" {
		t.Fatalf("unexpected imported resources: %#v", resources)
	}
	if got := resources[0].State().Attributes["content"]; got != userDataStateHash(content) {
		t.Fatalf("content state = %q, want SHA-256 digest", got)
	}
}

func TestDataSourceVpsUserDataReadExposesOnlyContentHash(t *testing.T) {
	const content = "#!/bin/sh\nprintf data-source\n"
	cfg := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v7.0/vps_user_data/42" {
			http.NotFound(w, r)
			return
		}
		writeAPIResponse(t, w, "vps_user_data", testVpsUserData(42, "existing", "script", content))
	})
	d := schema.TestResourceDataRaw(t, dataSourceVpsUserData().Schema, map[string]interface{}{
		"vps_user_data_id": 42,
	})

	if err := dataSourceVpsUserDataRead(d, cfg); err != nil {
		t.Fatal(err)
	}
	if d.Id() != "42" || d.Get("label").(string) != "existing" {
		t.Fatalf("unexpected data-source state: %#v", d.State().Attributes)
	}
	if got := d.Get("content_sha256").(string); got != userDataStateHash(content) {
		t.Fatalf("content_sha256 = %q, want SHA-256 digest", got)
	}
	if _, exists := dataSourceVpsUserData().Schema["content"]; exists {
		t.Fatal("data source exposes raw content")
	}
	if strings.Contains(fmt.Sprintf("%#v", d.State().Attributes), content) {
		t.Fatalf("data-source state contains raw content: %#v", d.State().Attributes)
	}
}

func TestDataSourceVpsUserDataReadReportsMissingResource(t *testing.T) {
	cfg := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		writeVpsUserDataNotFound(t, w)
	})
	d := schema.TestResourceDataRaw(t, dataSourceVpsUserData().Schema, map[string]interface{}{
		"vps_user_data_id": 42,
	})

	err := dataSourceVpsUserDataRead(d, cfg)
	if err == nil || !strings.Contains(err.Error(), "Input error") {
		t.Fatalf("dataSourceVpsUserDataRead() error = %v, want missing-resource error", err)
	}
}

func TestResponseReportsResourceNotFound(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		envelope *client.Envelope
		want     bool
	}{
		{name: "nil", envelope: nil, want: false},
		{name: "message", envelope: &client.Envelope{Message: "resource not found"}, want: true},
		{
			name: "parameter error",
			envelope: &client.Envelope{
				Message: "Input error",
				Errors:  map[string][]string{"vps_user_data_id": {"resource not found"}},
			},
			want: true,
		},
		{name: "other error", envelope: &client.Envelope{Message: "permission denied"}, want: false},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := responseReportsResourceNotFound(test.envelope); got != test.want {
				t.Fatalf("responseReportsResourceNotFound() = %t, want %t", got, test.want)
			}
		})
	}
}

func assertVpsUserDataRequest(t *testing.T, r *http.Request, want map[string]interface{}) {
	t.Helper()

	var request struct {
		VpsUserData map[string]interface{} `json:"vps_user_data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(request.VpsUserData, want) {
		t.Fatalf("vps_user_data = %#v, want %#v", request.VpsUserData, want)
	}
}

func testVpsUserData(id int64, label string, format string, content string) map[string]interface{} {
	return map[string]interface{}{
		"id":         id,
		"label":      label,
		"format":     format,
		"content":    content,
		"created_at": "2026-08-29T12:00:00Z",
		"updated_at": "2026-08-29T12:00:00Z",
	}
}

func writeVpsUserDataNotFound(t *testing.T, w http.ResponseWriter) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  false,
		"message": "Input error",
		"errors": map[string][]string{
			"vps_user_data_id": {"resource not found"},
		},
	}); err != nil {
		t.Fatal(err)
	}
}
