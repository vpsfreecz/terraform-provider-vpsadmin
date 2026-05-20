package vpsadmin

import "testing"

func TestGetPrimaryConnectionAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		publicIpv4  string
		privateIpv4 string
		publicIpv6  string
		want        string
	}{
		{
			name:        "prefers public IPv4",
			publicIpv4:  "198.51.100.10",
			privateIpv4: "10.0.0.10",
			publicIpv6:  "2001:db8::10",
			want:        "198.51.100.10",
		},
		{
			name:        "uses public IPv6 before private IPv4",
			privateIpv4: "10.0.0.10",
			publicIpv6:  "2001:db8::10",
			want:        "2001:db8::10",
		},
		{
			name:        "uses private IPv4 without public addresses",
			privateIpv4: "10.0.0.10",
			want:        "10.0.0.10",
		},
		{
			name: "returns empty string without addresses",
			want: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := getPrimaryConnectionAddress(
				tt.publicIpv4,
				tt.privateIpv4,
				tt.publicIpv6,
			)
			if got != tt.want {
				t.Fatalf("getPrimaryConnectionAddress() = %q, want %q", got, tt.want)
			}
		})
	}
}
