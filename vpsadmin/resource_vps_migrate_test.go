package vpsadmin

import (
	"context"
	"testing"

	"github.com/hashicorp/go-cty/cty/msgpack"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceVpsUpgradeV0MovesHistoricalOsTemplate(t *testing.T) {
	t.Parallel()

	const osTemplate = "debian-12"

	vpsResource := resourceVps()
	server := schema.NewGRPCProviderServer(&schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"vpsadmin_vps": vpsResource,
		},
	})

	resp, err := server.UpgradeResourceState(context.Background(), &tfprotov5.UpgradeResourceStateRequest{
		TypeName: "vpsadmin_vps",
		Version:  0,
		RawState: &tfprotov5.RawState{
			Flatmap: map[string]string{
				"id":          "123",
				"location":    "test",
				"os_template": osTemplate,
				"cpu":         "2",
				"memory":      "2048",
				"diskspace":   "40960",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Diagnostics) > 0 {
		for _, diagnostic := range resp.Diagnostics {
			t.Errorf("unexpected diagnostic: %s: %s", diagnostic.Summary, diagnostic.Detail)
		}
		t.FailNow()
	}
	if resp.UpgradedState == nil {
		t.Fatal("missing upgraded state")
	}

	state, err := msgpack.Unmarshal(
		resp.UpgradedState.MsgPack,
		vpsResource.CoreConfigSchema().ImpliedType(),
	)
	if err != nil {
		t.Fatal(err)
	}

	attrs := state.AsValueMap()
	got, ok := attrs["install_os_template"]
	if !ok {
		t.Fatal("install_os_template not found in upgraded state")
	}
	if !got.IsKnown() || got.IsNull() {
		t.Fatalf("install_os_template is not populated: %s", got.GoString())
	}
	if got.AsString() != osTemplate {
		t.Fatalf("install_os_template = %q, want %q", got.AsString(), osTemplate)
	}
	if _, ok := attrs["os_template"]; ok {
		t.Fatal("os_template found in upgraded state")
	}

	rawState, err := resourceVpsUpgradeV0(context.Background(), map[string]interface{}{
		"os_template": osTemplate,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if rawState["install_os_template"] != osTemplate {
		t.Fatalf(
			"install_os_template = %#v, want %#v",
			rawState["install_os_template"],
			osTemplate,
		)
	}
	if _, ok := rawState["os_template"]; ok {
		t.Fatal("os_template found in raw upgraded state")
	}
}
