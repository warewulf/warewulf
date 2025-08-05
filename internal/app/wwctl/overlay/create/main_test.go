package create

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func Test_OverlayCreate(t *testing.T) {
	tests := []struct {
		name        string
		overlayName string
		setup       func(t *testing.T, overlayDir string)
		wantErr     bool
	}{
		{
			name:        "simple create",
			overlayName: "test-overlay",
			setup:       func(t *testing.T, overlayDir string) {},
			wantErr:     false,
		},
		{
			name:        "create existing overlay",
			overlayName: "test-overlay-exists",
			setup: func(t *testing.T, overlayDir string) {
				err := os.MkdirAll(overlayDir, 0755)
				assert.NoError(t, err)
			},
			wantErr: true,
		},
	}

	warewulfd.SetNoDaemon()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testenv.New(t)
			conf := env.Configure()

			overlayDir := filepath.Join(conf.Paths.WWOverlaydir, tt.overlayName)
			tt.setup(t, overlayDir)

			baseCmd := GetCommand()
			baseCmd.SetArgs([]string{tt.overlayName})
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			err := baseCmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				_, err := os.Stat(overlayDir)
				assert.NoError(t, err, "overlay directory should have been created")
			}
		})
	}
}
