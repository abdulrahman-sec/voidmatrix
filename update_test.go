package main

import "testing"

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
	}{
		// Stable releases
		{"v1.0.1", "v1.0.2", true},
		{"v1.0.1", "v1.1.0", true},
		{"v1.0.1", "v2.0.0", true},
		{"v1.0.1", "v1.0.1", false},
		{"v1.0.12", "v1.0.2", false},
		{"v1.0.2", "v1.0.12", true},

		// Pre-releases to stable
		{"v1.0.2-beta", "v1.0.2", true},
		{"v1.0.1-rc1", "v1.0.2", true},

		// Stable to pre-release (filtered out)
		{"v1.0.1", "v1.0.2-beta", false},
		{"v1.0.1", "v1.0.2-rc2", false},

		// Pre-release to pre-release (filtered out or same)
		{"v1.0.2-beta", "v1.0.2-rc1", false},
		{"v1.0.1-alpha", "v1.0.1-beta", false},
	}

	for _, tt := range tests {
		got := isNewerVersion(tt.current, tt.latest)
		if got != tt.want {
			t.Errorf("isNewerVersion(%q, %q) = %v; want %v", tt.current, tt.latest, got, tt.want)
		}
	}
}
