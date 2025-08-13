package chown

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
)

func Test_OverlayChown(t *testing.T) {
	currentUser := strconv.Itoa(os.Getuid())
	currentGroup := strconv.Itoa(os.Getgid())

	tests := []struct {
		name        string
		overlayName string
		fileName    string
		chownSpec   string
		wantErr     bool
		system      bool
		checkUser   bool
		checkGroup  bool
	}{
		{
			name:        "simple chown user and group",
			overlayName: "test-overlay-1",
			fileName:    "test.txt",
			chownSpec:   currentUser + ":" + currentGroup,
			wantErr:     false,
			system:      false,
			checkUser:   true,
			checkGroup:  true,
		},
		{
			name:        "simple chown user only",
			overlayName: "test-overlay-2",
			fileName:    "test.txt",
			chownSpec:   currentUser,
			wantErr:     false,
			system:      false,
			checkUser:   true,
			checkGroup:  false, // group should not change
		},
		{
			name:        "simple chown group only",
			overlayName: "test-overlay-3",
			fileName:    "test.txt",
			chownSpec:   ":" + currentGroup,
			wantErr:     false,
			system:      false,
			checkUser:   false, // user should not change
			checkGroup:  true,
		},
		{
			name:        "system overlay chown",
			overlayName: "wwinit", // A known system overlay
			fileName:    "init.sh",
			chownSpec:   currentUser + ":" + currentGroup,
			wantErr:     false,
			system:      true,
			checkUser:   true,
			checkGroup:  true,
		},
		{
			name:        "bad chown spec",
			overlayName: "test-overlay-4",
			fileName:    "test.txt",
			chownSpec:   "bad:bad",
			wantErr:     true,
			system:      false,
			checkUser:   false,
			checkGroup:  false,
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
			f, err := os.Create(filePath)
			assert.NoError(t, err)
			f.Close()
			// get initial owner
			stat, err := os.Stat(filePath)
			assert.NoError(t, err)
			sysStat, ok := stat.Sys().(*syscall.Stat_t)
			assert.True(t, ok)
			startUID := int(sysStat.Uid)
			startGID := int(sysStat.Gid)

			baseCmd := GetCommand()
			baseCmd.SetArgs([]string{tt.overlayName, tt.fileName, tt.chownSpec})
			buf := new(bytes.Buffer)
			baseCmd.SetOut(buf)
			baseCmd.SetErr(buf)
			err = baseCmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check ownership
			// for system overlays, the file is copied to a site overlay
			if tt.system {
				overlayDir = filepath.Join(conf.Paths.WWOverlaydir, tt.overlayName)
				filePath = filepath.Join(overlayDir, tt.fileName)
			}
			stat, err = os.Stat(filePath)
			assert.NoError(t, err)
			sysStat, ok = stat.Sys().(*syscall.Stat_t)
			assert.True(t, ok)
			endUID := int(sysStat.Uid)
			endGID := int(sysStat.Gid)

			if !tt.wantErr {
				if tt.checkUser {
					u, err := strconv.Atoi(currentUser)
					assert.NoError(t, err)
					assert.Equal(t, u, endUID)
				} else {
					assert.Equal(t, startUID, endUID)
				}
				if tt.checkGroup {
					g, err := strconv.Atoi(currentGroup)
					assert.NoError(t, err)
					assert.Equal(t, g, endGID)
				} else {
					assert.Equal(t, startGID, endGID)
				}
			} else {
				assert.Equal(t, startUID, endUID)
				assert.Equal(t, startGID, endGID)
			}
		})
	}
}
