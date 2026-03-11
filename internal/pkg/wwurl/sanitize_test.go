package wwurl

import (
	"fmt"
	"strings"
	"testing"
)

func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "redacts assetkey",
			input: "https://192.168.3.1:9874/provision/00:0c:29:7c:49:6f?assetkey=secretvalue&compress=gz&stage=runtime&uuid=62184d56-6d53-9895-0b51-035f457c496f",
			want:  "https://192.168.3.1:9874/provision/00:0c:29:7c:49:6f?assetkey=REDACTED&compress=gz&stage=runtime&uuid=62184d56-6d53-9895-0b51-035f457c496f",
		},
		{
			name:  "no assetkey unchanged",
			input: "https://192.168.3.1:9874/provision/00:0c:29:7c:49:6f?compress=gz&stage=runtime",
			want:  "https://192.168.3.1:9874/provision/00:0c:29:7c:49:6f?compress=gz&stage=runtime",
		},
		{
			name:  "invalid URL returned as-is",
			input: "not a url ://bad",
			want:  "not a url ://bad",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeURL(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeURL(%q)\n got:  %q\n want: %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizeError(t *testing.T) {
	// Simulate a Go net/http transport error that embeds the full URL.
	errMsg := `Get "https://192.168.3.1:9874/provision/00:0c:29:7c:49:6f?assetkey=mysecret&compress=gz&stage=runtime&uuid=abc123": dial tcp 192.168.3.1:9874: connect: connection refused`
	err := fmt.Errorf("%s", errMsg)
	got := SanitizeError(err)
	if got == errMsg {
		t.Error("SanitizeError did not redact assetkey from error message")
	}
	if strings.Contains(got, "mysecret") {
		t.Errorf("SanitizeError still contains secret: %q", got)
	}
	if !strings.Contains(got, "REDACTED") {
		t.Errorf("SanitizeError missing REDACTED marker: %q", got)
	}
}
