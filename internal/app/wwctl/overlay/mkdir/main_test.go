package mkdir

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

func Test_OverlayMkdir(t *testing.T) {
	tests := []struct {
		name        string
		overlayName string
		dirName     string
		perm        string
		wantErr     bool
		system      bool
	}{
		{
			name:        "simple mkdir",
			overlayName: "test-overlay",
			dirName:     "testdir",
			perm:        "0755",
			wantErr:     false,
			system:      false,
		},
		{
			name:        "system overlay mkdir",
			overlayName: "wwinit", // A known system overlay
			dirName:     "init.d",
			perm:        "0700",
			wantErr:     false,
			system:      true,
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
				err := os.MkdirAll(overlayDir, 0755)
				assert.NoError(t, err)
			} else {
				overlayDir = filepath.Join(conf.Paths.WWOverlaydir, tt.overlayName)
				err := os.MkdirAll(overlayDir, 0755)
				assert.NoError(t, err)
			}

			baseCmd := GetCommand()
			baseCmd.SetArgs([]string{"-m", tt.perm, tt.overlayName, tt.dirName})
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			err := baseCmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check permissions
			// for system overlays, the file is copied to a site overlay
			if tt.system {
				overlayDir = filepath.Join(conf.Paths.WWOverlaydir, tt.overlayName)
			}
			dirPath := filepath.Join(overlayDir, tt.dirName)
			stat, err := os.Stat(dirPath)
			assert.NoError(t, err)
			if !tt.wantErr {
				mode, err := strconv.ParseUint(tt.perm, 8, 32)
				assert.NoError(t, err)
				assert.Equal(t, fs.FileMode(mode), stat.Mode().Perm())
			}
		})
	}
}
