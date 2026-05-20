package vpsadmin

import (
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceVpsUpdateReturnsDiskspaceLookupError(t *testing.T) {
	cfg := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v7.0/vpses/123" {
			http.NotFound(w, r)
			return
		}

		writeAPIError(t, w, "missing VPS")
	})

	d := newResourceDataWithDiff(
		t,
		resourceVps().Schema,
		"123",
		map[string]string{
			"diskspace": "1024",
		},
		map[string]*terraform.ResourceAttrDiff{
			"diskspace": {
				Old: "1024",
				New: "2048",
			},
		},
	)

	err := resourceVpsUpdate(d, cfg)
	if err == nil || !strings.Contains(err.Error(), "missing VPS") {
		t.Fatalf("resourceVpsUpdate() error = %v, want missing VPS", err)
	}
}

func TestHasAnyVpsFeatureChange(t *testing.T) {
	t.Parallel()

	for _, feature := range supportedVpsFeatures {
		feature := feature
		t.Run(feature, func(t *testing.T) {
			t.Parallel()

			d := newResourceDataWithDiff(
				t,
				resourceVps().Schema,
				"123",
				map[string]string{
					"feature_" + feature: "false",
				},
				map[string]*terraform.ResourceAttrDiff{
					"feature_" + feature: {
						Old: "false",
						New: "true",
					},
				},
			)

			if !hasAnyVpsFeatureChange(d) {
				t.Fatalf("hasAnyVpsFeatureChange() = false, want true")
			}
		})
	}

	t.Run("no change", func(t *testing.T) {
		t.Parallel()

		d := newResourceDataWithDiff(
			t,
			resourceVps().Schema,
			"123",
			nil,
			nil,
		)

		if hasAnyVpsFeatureChange(d) {
			t.Fatalf("hasAnyVpsFeatureChange() = true, want false")
		}
	})
}
