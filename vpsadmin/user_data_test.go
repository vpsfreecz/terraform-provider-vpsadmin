package vpsadmin

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vpsfreecz/vpsadmin-go-client/client"
)

func TestUserDataStateHash(t *testing.T) {
	t.Parallel()

	got := userDataStateHash("provider user data\n")
	want := "78d1d76bc8e9f797a0c19a397d6c25648ccb99bbe7920ab4d256d5fa23f413df"

	if got != want {
		t.Fatalf("userDataStateHash() = %q, want %q", got, want)
	}
}

func TestResourceVpsInlineUserDataSchema(t *testing.T) {
	t.Parallel()

	resource := resourceVps()
	content := resource.Schema["user_data"]
	format := resource.Schema["user_data_format"]

	if !content.Optional || !content.ForceNew || !content.Sensitive || content.StateFunc == nil {
		t.Fatalf("unexpected user_data schema: %#v", content)
	}
	if !format.Optional || !format.ForceNew {
		t.Fatalf("unexpected user_data_format schema: %#v", format)
	}

	for _, value := range supportedUserDataFormats {
		if _, errors := format.ValidateFunc(value, "user_data_format"); len(errors) != 0 {
			t.Errorf("user_data_format rejected %q: %v", value, errors)
		}
	}
	if _, errors := format.ValidateFunc("invalid", "user_data_format"); len(errors) == 0 {
		t.Error("user_data_format accepted an unsupported format")
	}
	if _, errors := content.ValidateFunc("", "user_data"); len(errors) == 0 {
		t.Error("user_data accepted empty content")
	}
	const secretMarker = "inline-user-data-secret-marker"
	_, errors := content.ValidateFunc(
		strings.Repeat("x", maxUserDataBytes+1)+secretMarker,
		"user_data",
	)
	if len(errors) == 0 {
		t.Error("user_data accepted content over the API limit")
	}
	if strings.Contains(fmt.Sprint(errors), secretMarker) {
		t.Fatalf("user_data validation exposed content: %v", errors)
	}
}

func TestValidateUserDataContentCountsBytes(t *testing.T) {
	t.Parallel()

	withinLimit := strings.Repeat("ž", maxUserDataBytes/2) + "x"
	if len(withinLimit) != maxUserDataBytes {
		t.Fatalf("test content has %d bytes, want %d", len(withinLimit), maxUserDataBytes)
	}
	if _, errors := validateUserDataContent(withinLimit, "user_data"); len(errors) != 0 {
		t.Fatalf("content at the byte limit was rejected: %v", errors)
	}

	overLimit := withinLimit + "ž"
	if _, errors := validateUserDataContent(overLimit, "user_data"); len(errors) == 0 {
		t.Error("multibyte content over the byte limit was accepted")
	}
}

func TestResourceVpsUserDataFormatRequiresContent(t *testing.T) {
	t.Parallel()

	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"location":            "test",
		"install_os_template": "debian-latest",
		"cpu":                 1,
		"memory":              1024,
		"diskspace":           4096,
		"user_data_format":    "script",
	})
	diagnostics := schema.InternalMap(resourceVps().Schema).Validate(config)

	if len(diagnostics) == 0 {
		t.Fatal("schema accepted user_data_format without user_data")
	}
	if !strings.Contains(diagnostics[0].Detail, "user_data") {
		t.Fatalf("unexpected diagnostic: %#v", diagnostics[0])
	}
}

func TestResourceVpsUserDataScriptDefaultIsEquivalent(t *testing.T) {
	t.Parallel()

	field := resourceVps().Schema["user_data_format"]
	d := schema.TestResourceDataRaw(t, resourceVps().Schema, nil)

	if !field.DiffSuppressFunc("user_data_format", "", defaultUserDataFormat, d) {
		t.Error("adding the explicit script default was not suppressed")
	}
	if !field.DiffSuppressFunc("user_data_format", defaultUserDataFormat, "", d) {
		t.Error("removing the explicit script default was not suppressed")
	}
	if field.DiffSuppressFunc("user_data_format", "script", "cloudinit_config", d) {
		t.Error("a real user-data format change was suppressed")
	}
}

func TestConfigureVpsInlineUserData(t *testing.T) {
	t.Parallel()

	const content = "#!/bin/sh\nprintf inline > /root/provider-user-data\n"

	for _, test := range []struct {
		name       string
		config     map[string]interface{}
		wantFormat string
	}{
		{
			name:       "default format",
			config:     map[string]interface{}{"user_data": content},
			wantFormat: "script",
		},
		{
			name: "explicit format",
			config: map[string]interface{}{
				"user_data":        content,
				"user_data_format": "cloudinit_script",
			},
			wantFormat: "cloudinit_script",
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			d := schema.TestResourceDataRaw(t, resourceVps().Schema, test.config)
			input := &client.ActionVpsCreateInput{}
			configureVpsInlineUserData(input, d)

			if input.UserDataContent != content {
				t.Fatalf("UserDataContent = %q, want raw content", input.UserDataContent)
			}
			if input.UserDataFormat != test.wantFormat {
				t.Fatalf("UserDataFormat = %q, want %q", input.UserDataFormat, test.wantFormat)
			}
		})
	}
}

func TestConfigureVpsInlineUserDataLeavesInputUnset(t *testing.T) {
	t.Parallel()

	d := schema.TestResourceDataRaw(t, resourceVps().Schema, nil)
	input := &client.ActionVpsCreateInput{}
	configureVpsInlineUserData(input, d)

	if input.UserDataContent != "" || input.UserDataFormat != "" {
		t.Fatalf("unexpected user-data input: %#v", input)
	}
}

func TestRedactedVpsCreateInputHidesUserData(t *testing.T) {
	t.Parallel()

	const content = "provider-user-data-secret"
	input := (&client.ActionVpsCreateInput{}).SetUserDataContent(content)
	redacted := redactedVpsCreateInput(input)
	formatted := fmt.Sprintf("%#v", redacted)

	if strings.Contains(formatted, content) {
		t.Fatalf("redacted input contains raw user data: %s", formatted)
	}
	if redacted.UserDataContent != "[REDACTED]" {
		t.Fatalf("UserDataContent = %q, want redaction marker", redacted.UserDataContent)
	}
	if input.UserDataContent != content {
		t.Fatal("redaction modified the API input")
	}
}
