package vpsadmin

import (
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestProviderRegistersExpectedSchemas(t *testing.T) {
	provider := Provider()
	if provider == nil {
		t.Fatal("Provider() returned nil")
	}

	assertMapKeys(t, provider.Schema, []string{
		"api_url",
		"auth_token",
	})

	for _, name := range []string{"api_url", "auth_token"} {
		t.Run(name, func(t *testing.T) {
			field := provider.Schema[name]
			if field == nil {
				t.Fatalf("schema %q is nil", name)
			}
			if field.Type != schema.TypeString {
				t.Fatalf("schema %q type = %v, want %v", name, field.Type, schema.TypeString)
			}
			if !field.Required {
				t.Fatalf("schema %q is not required", name)
			}
			if field.DefaultFunc == nil {
				t.Fatalf("schema %q has no default function", name)
			}
		})
	}

	if provider.ConfigureFunc == nil {
		t.Fatal("ConfigureFunc is nil")
	}
}

func TestProviderRegistersExpectedDataSources(t *testing.T) {
	provider := Provider()
	if provider == nil {
		t.Fatal("Provider() returned nil")
	}

	assertMapKeys(t, provider.DataSourcesMap, []string{
		"vpsadmin_dataset",
		"vpsadmin_mount",
		"vpsadmin_ssh_key",
		"vpsadmin_vps",
	})

	for name, dataSource := range provider.DataSourcesMap {
		if dataSource == nil {
			t.Fatalf("data source %q is nil", name)
		}
		if dataSource.Read == nil {
			t.Fatalf("data source %q has no Read function", name)
		}
	}
}

func TestProviderRegistersExpectedResources(t *testing.T) {
	provider := Provider()
	if provider == nil {
		t.Fatal("Provider() returned nil")
	}

	assertMapKeys(t, provider.ResourcesMap, []string{
		"vpsadmin_dataset",
		"vpsadmin_mount",
		"vpsadmin_ssh_key",
		"vpsadmin_vps",
	})

	for name, resource := range provider.ResourcesMap {
		if resource == nil {
			t.Fatalf("resource %q is nil", name)
		}
		if resource.Create == nil {
			t.Fatalf("resource %q has no Create function", name)
		}
		if resource.Read == nil {
			t.Fatalf("resource %q has no Read function", name)
		}
		if resource.Update == nil {
			t.Fatalf("resource %q has no Update function", name)
		}
		if resource.Delete == nil {
			t.Fatalf("resource %q has no Delete function", name)
		}
	}
}

func assertMapKeys[V any](t *testing.T, got map[string]V, want []string) {
	t.Helper()

	gotKeys := make([]string, 0, len(got))
	for key := range got {
		gotKeys = append(gotKeys, key)
	}
	sort.Strings(gotKeys)

	wantKeys := append([]string(nil), want...)
	sort.Strings(wantKeys)

	if !reflect.DeepEqual(gotKeys, wantKeys) {
		t.Fatalf("keys = %v, want %v", gotKeys, wantKeys)
	}
}
