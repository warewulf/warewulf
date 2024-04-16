package container

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
)

func TestInitramfsBootPath(t *testing.T) {
	conf := warewulfconf.Get()
	temp, err := os.MkdirTemp(os.TempDir(), "ww-conf-*")
	assert.NoError(t, err)
	defer os.RemoveAll(temp)
	conf.Paths.WWChrootdir = temp

	assert.NoError(t, os.MkdirAll(filepath.Join(RootFsDir("image"), "boot"), 0700))

	tests := []struct {
		name      string
		initramfs []string
		ver       string
		err       error
		retName   string
	}{
		{
			name:      "ok case 1",
			initramfs: []string{"initramfs-1.1.1.aarch64.img"},
			ver:       "1.1.1.aarch64",
			err:       nil,
		},
		{
			name:      "ok case 2",
			initramfs: []string{"initrd-1.1.1.aarch64"},
			ver:       "1.1.1.aarch64",
			err:       nil,
		},
		{
			name:      "ok case 3",
			initramfs: []string{"initramfs-1.1.1.aarch64"},
			ver:       "1.1.1.aarch64",
			err:       nil,
		},
		{
			name:      "ok case 4",
			initramfs: []string{"initrd-1.1.1.aarch64.img"},
			ver:       "1.1.1.aarch64",
			err:       nil,
		},
		{
			name:      "error case, wrong init name",
			initramfs: []string{"initrr-1.1.1.aarch64.img"},
			ver:       "1.1.1.aarch64",
			err:       fmt.Errorf("Failed to find a target kernel version initramfs"),
		},
		{
			name:      "error case, wrong ver",
			initramfs: []string{"initrr-1.1.1.aarch64.img"},
			ver:       "1.1.2.aarch64",
			err:       fmt.Errorf("Failed to find a target kernel version initramfs"),
		},
	}

	for _, tt := range tests {
		t.Logf("running test: %s", tt.name)
		for _, init := range tt.initramfs {
			assert.NoError(t, os.WriteFile(filepath.Join(RootFsDir("image"), "boot", init), []byte(""), 0600))
		}
		initPath, err := InitramfsBootPath("image", tt.ver)
		assert.Equal(t, tt.err, err)
		if err == nil {
			assert.NotEmpty(t, initPath)
		} else {
			assert.Empty(t, initPath)
		}

		if tt.retName != "" {
			assert.Equal(t, filepath.Base(initPath), tt.retName)
		}
		// remove the file
		for _, init := range tt.initramfs {
			assert.NoError(t, os.Remove(filepath.Join(RootFsDir("image"), "boot", init)))
		}
	}
}
