package util

import "testing"

func TestParseVersion(t *testing.T) {
	tests := map[string]struct {
		path     string
		expected string
	}{
		"el9_4 module path":             {"/lib/modules/5.14.0-427.37.1.el9_4.aarch64/vmlinuz", "5.14.0-427.37.1"},
		"el8_6 boot path":               {"/boot/vmlinuz-4.18.0-372.13.1.el8_6.x86_64", "4.18.0-372.13.1"},
		"el10_2 with suffix after dist": {"/lib/modules/6.12.0-211.16.1.el10_2.0.1.x86_64/vmlinuz", "6.12.0-211.16.1"},
		"el9_8 with suffix after dist":  {"/lib/modules/5.14.0-687.10.1.el9_8.0.1.x86_64/vmlinuz", "5.14.0-687.10.1"},
		"el9_8 plain dist tag":          {"/lib/modules/5.14.0-687.10.1.el9_8.x86_64/vmlinuz", "5.14.0-687.10.1"},
		"boot path with .gz":            {"/boot/vmlinuz-5.14.0-427.24.1.el9_4.x86_64.gz", "5.14.0-427.24.1"},
		"no version in path":            {"/boot/vmlinuz-linux", ""},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := ""
			if v := ParseVersion(tt.path); v != nil {
				got = v.String()
			}
			if got != tt.expected {
				t.Errorf("ParseVersion(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}
