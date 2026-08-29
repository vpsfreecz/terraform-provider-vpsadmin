package vpsadmin

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

const (
	defaultUserDataFormat = "script"
	maxUserDataBytes      = 65535
)

var supportedUserDataFormats = []string{
	"script",
	"cloudinit_config",
	"cloudinit_script",
	"nixos_configuration",
	"nixos_flake_configuration",
	"nixos_flake_uri",
}

var supportedUserDataFormatsText = "`" + strings.Join(supportedUserDataFormats, "`, `") + "`"

func validateUserDataContent(value interface{}, key string) ([]string, []error) {
	content, ok := value.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %s to be string", key)}
	}

	if len(content) < 1 || len(content) > maxUserDataBytes {
		return nil, []error{fmt.Errorf(
			"expected byte length of %s to be in the range (1 - %d), got %d",
			key,
			maxUserDataBytes,
			len(content),
		)}
	}

	return nil, nil
}

func userDataStateHash(value interface{}) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(value.(string))))
}
