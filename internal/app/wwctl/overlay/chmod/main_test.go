package chmod

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func Test_OverlayChmod(t *testing.T) {
	tests := []struct {
		name        string
		overlayName string
		fileName    string
		perm        string
		wantErr     bool
		system      bool
		startPerm   fs.FileMode
	}{
		{
			name:        "simple chmod",
			overlayName: "test-overlay",
			fileName:    "test.txt",
			perm:        "0600",
			wantErr:     false,
			system:      false,
			startPerm:   0644,
		},
		{
			name:        "system overlay chmod",
			overlayName: "wwinit", // A known system overlay
			fileName:    "init.sh",
			perm:        "0700",
			wantErr:     false,
			system:      true,
			startPerm:   0755,
		},
	}

	warewulfd.SetNoDaemon()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testenv.New(t)
			conf := env.Configure()

			// Setup overlay
			var overlayDir string
			if tt.system {
				overlayDir = filepath.Join(conf.Paths.DistributionOverlaydir(), tt.overlayName)
			} else {
				overlayDir = filepath.Join(conf.Paths.WWOverlaydir, tt.overlayName)
			}
			err := os.MkdirAll(overlayDir, 0755)
			assert.NoError(t, err)
			filePath := filepath.Join(overlayDir, tt.fileName)
			err = os.WriteFile(filePath, []byte("test"), tt.startPerm)
			assert.NoError(t, err)

			baseCmd := GetCommand()
			baseCmd.SetArgs([]string{tt.overlayName, tt.fileName, tt.perm})
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			err = baseCmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check permissions
			// for system overlays, the file is copied to a site overlay
			if tt.system {
				overlayDir = filepath.Join(conf.Paths.WWOverlaydir, tt.overlayName)
				filePath = filepath.Join(overlayDir, tt.fileName)
			}
			stat, err := os.Stat(filePath)
			assert.NoError(t, err)
			if !tt.wantErr {
				mode, err := strconv.ParseUint(tt.perm, 8, 32)
				assert.NoError(t, err)
				assert.Equal(t, fs.FileMode(mode), stat.Mode().Perm())
			} else {
				assert.Equal(t, tt.startPerm, stat.Mode().Perm())
			}
		})
	}
}
