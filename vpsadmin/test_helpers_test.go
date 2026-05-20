package vpsadmin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vpsfreecz/vpsadmin-go-client/client"
)

func newTestConfig(t *testing.T, handler http.HandlerFunc) *Config {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return &Config{
		client: client.New(server.URL),
	}
}

func writeAPIResponse(t *testing.T, w http.ResponseWriter, key string, value interface{}) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status": true,
		"response": map[string]interface{}{
			key: value,
		},
	}); err != nil {
		t.Fatal(err)
	}
}

func writeAPIError(t *testing.T, w http.ResponseWriter, message string) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  false,
		"message": message,
	}); err != nil {
		t.Fatal(err)
	}
}

func newResourceDataWithDiff(
	t *testing.T,
	resourceSchema map[string]*schema.Schema,
	id string,
	state map[string]string,
	diff map[string]*terraform.ResourceAttrDiff,
) *schema.ResourceData {
	t.Helper()

	d, err := schema.InternalMap(resourceSchema).Data(
		&terraform.InstanceState{
			ID:         id,
			Attributes: state,
		},
		&terraform.InstanceDiff{
			Attributes: diff,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	return d
}
