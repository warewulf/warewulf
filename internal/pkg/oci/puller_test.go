package oci

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetReference(t *testing.T) {
	temp, err := os.MkdirTemp(os.TempDir(), "ww-archive-*")
	assert.NoError(t, err)
	defer os.RemoveAll(temp)

	tests := []struct {
		name string
		uri  string
		err  error
	}{
		{
			name: "file archive ok case",
			uri:  "test.tar",
			err:  nil,
		},
		{
			name: "file archive ko case, because having colon in file name",
			uri:  "test:latest.tar",
			err:  fmt.Errorf("/test:latest.tar should not contain colon"),
		},
		{
			name: "file archive ko case, because having colon in path name",
			uri:  "/test:/test.tar",
			err:  fmt.Errorf("/test:/test.tar should not contain colon"),
		},
		{
			name: "docker ok case",
			uri:  "docker://test:latest",
			err:  nil,
		},
		{
			name: "docker daemon ok case",
			uri:  "docker-daemon://test:latest",
			err:  nil,
		},
	}

	for _, tt := range tests {
		t.Logf("running test: %s", tt.name)
		if !strings.HasPrefix(tt.uri, "docker") && !strings.HasPrefix(tt.uri, "docker-daemon") {
			tt.uri = filepath.Join(temp, tt.uri)
			parent := filepath.Dir(tt.uri)
			err := os.MkdirAll(parent, 0o755)
			assert.NoError(t, err)
			f, err := os.Create(tt.uri)
			assert.NoError(t, err)
			assert.NoError(t, f.Close())
		}

		_, err := getReference(tt.uri)
		if tt.err == nil {
			assert.NoError(t, err)
		} else {
			assert.ErrorContains(t, err, tt.err.Error())
		}
	}

}
