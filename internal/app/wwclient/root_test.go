package wwclient

import (
	"testing"
)

func TestParseWWIDFromCmdline(t *testing.T) {
	tests := []struct {
		cmdline     string
		expected    string
		expectError bool
	}{
		{"wwid=123", "123", false},
		{"foo wwid=abc bar", "abc", false},
		{"wwid=", "", true},
		{"foo bar", "", true},
		{"wwid=123 wwid=456", "123", false}, // it takes the first one
	}

	for _, tt := range tests {
		t.Run(tt.cmdline, func(t *testing.T) {
			actual, err := parseWWIDFromCmdline(tt.cmdline)
			if (err != nil) != tt.expectError {
				t.Errorf("parseWWIDFromCmdline(%q) error = %v; expectError %v", tt.cmdline, err, tt.expectError)
				return
			}
			if actual != tt.expected {
				t.Errorf("parseWWIDFromCmdline(%q) = %q; expected %q", tt.cmdline, actual, tt.expected)
			}
		})
	}
}

func TestParseTPMFromCmdline(t *testing.T) {
	tests := []struct {
		cmdline  string
		expected bool
	}{
		{"tpm", true},
		{"TPM", true},
		{"tpm=1", true},
		{"tpm=true", true},
		{"TPM=true", true},
		{"tpm=yes", true},
		{"tpm=on", true},
		{"tpm=0", false},
		{"tpm=false", false},
		{"TPM=false", false},
		{"tpm=no", false},
		{"tpm=off", false},
		{"foo tpm bar", true},
		{"foo tpm=1 bar", true},
		{"foo tpm=0 bar", false},
		{"tpm=1 tpm=0", false}, // last one wins
		{"tpm=0 tpm=1", true},  // last one wins
		{"", false},
		{"foo bar", false},
		{"tpm_extra", false},
	}

	for _, tt := range tests {
		t.Run(tt.cmdline, func(t *testing.T) {
			actual := parseTPMFromCmdline(tt.cmdline)
			if actual != tt.expected {
				t.Errorf("parseTPMFromCmdline(%q) = %v; expected %v", tt.cmdline, actual, tt.expected)
			}
		})
	}
}
