package vpsadmin

import "testing"

func TestIsSupportedVpsFeature(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		feature string
		want    bool
	}{
		{name: "fuse", feature: "fuse", want: true},
		{name: "kvm", feature: "kvm", want: true},
		{name: "lxc", feature: "lxc", want: true},
		{name: "ppp", feature: "ppp", want: true},
		{name: "tun", feature: "tun", want: true},
		{name: "unsupported", feature: "nfs", want: false},
		{name: "case-sensitive", feature: "FUSE", want: false},
		{name: "empty", feature: "", want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isSupportedVpsFeature(tt.feature)
			if got != tt.want {
				t.Fatalf("isSupportedVpsFeature(%q) = %t, want %t", tt.feature, got, tt.want)
			}
		})
	}
}
