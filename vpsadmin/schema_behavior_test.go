package vpsadmin

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceSshKeyTrimsKeyState(t *testing.T) {
	t.Parallel()

	stateFunc := resourceSshKey().Schema["key"].StateFunc
	got := stateFunc("  ssh-ed25519 AAAAC3Nza test@example\n")
	want := "ssh-ed25519 AAAAC3Nza test@example"

	if got != want {
		t.Fatalf("StateFunc() = %q, want %q", got, want)
	}
}

func TestResourceVpsIpCountDiffSuppress(t *testing.T) {
	t.Parallel()

	resource := resourceVps()
	for _, key := range []string{
		"public_ipv4_count",
		"private_ipv4_count",
		"public_ipv6_count",
	} {
		key := key
		t.Run(key, func(t *testing.T) {
			t.Parallel()

			field := resource.Schema[key]
			newData := newResourceDataWithDiff(t, resource.Schema, "", nil, nil)
			if field.DiffSuppressFunc(key, "1", "2", newData) {
				t.Fatalf("DiffSuppressFunc() before create = true, want false")
			}

			existingData := newResourceDataWithDiff(t, resource.Schema, "123", nil, nil)
			if !field.DiffSuppressFunc(key, "1", "2", existingData) {
				t.Fatalf("DiffSuppressFunc() after create = false, want true")
			}
		})
	}
}

func TestResourceVpsSshKeysRejectZeroValues(t *testing.T) {
	t.Parallel()

	elem := resourceVps().Schema["ssh_keys"].Elem.(*schema.Schema)
	if _, errs := elem.ValidateFunc("", "ssh_keys"); len(errs) == 0 {
		t.Fatal("ValidateFunc(\"\") returned no errors")
	}
	if _, errs := elem.ValidateFunc("123", "ssh_keys"); len(errs) != 0 {
		t.Fatalf("ValidateFunc(\"123\") errors = %v, want none", errs)
	}
}

func TestResourceDatasetExportDiffSuppress(t *testing.T) {
	t.Parallel()

	resource := resourceDataset()
	for _, key := range []string{
		"export_enable",
		"export_root_squash",
		"export_read_write",
		"export_sync",
	} {
		key := key
		t.Run(key, func(t *testing.T) {
			t.Parallel()

			field := resource.Schema[key]
			disabled := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
				"name":           "app",
				"export_dataset": false,
			})
			if !field.DiffSuppressFunc(key, "false", "true", disabled) {
				t.Fatalf("DiffSuppressFunc() with export disabled = false, want true")
			}

			enabled := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
				"name":           "app",
				"export_dataset": true,
			})
			if field.DiffSuppressFunc(key, "false", "true", enabled) {
				t.Fatalf("DiffSuppressFunc() with export enabled = true, want false")
			}
		})
	}
}
