package container

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

func TestFindInitramfs(t *testing.T) {
	conf := warewulfconf.Get()
	temp, err := os.MkdirTemp(os.TempDir(), "ww-conf-*")
	assert.NoError(t, err)
	defer os.RemoveAll(temp)
	conf.Paths.WWChrootdir = temp

	assert.NoError(t, os.MkdirAll(filepath.Join(RootFsDir("image"), "boot"), 0700))

	tests := map[string]struct {
		name      string
		initramfs []string
		ver       string
		path      string
	}{
		"ok case 1": {
			initramfs: []string{"/boot/initramfs-1.1.1.aarch64.img"},
			ver:       "1.1.1",
			path:      "/boot/initramfs-1.1.1.aarch64.img",
		},
		"ok case 2": {
			initramfs: []string{"/boot/initrd-1.1.1.aarch64"},
			ver:       "1.1.1",
			path:      "/boot/initrd-1.1.1.aarch64",
		},
		"ok case 3": {
			initramfs: []string{"/boot/initramfs-1.1.1.aarch64"},
			ver:       "1.1.1",
			path:      "/boot/initramfs-1.1.1.aarch64",
		},
		"ok case 4": {
			initramfs: []string{"/boot/initrd-1.1.1.aarch64.img"},
			ver:       "1.1.1",
			path:      "/boot/initrd-1.1.1.aarch64.img",
		},
		"prefix match": {
			initramfs: []string{"/boot/initrd-1.1.1.aarch64.img"},
			ver:       "1.1",
			path:      "/boot/initrd-1.1.1.aarch64.img",
		},
		"error case, wrong init name": {
			initramfs: []string{"/boot/initrr-1.1.1.aarch64.img"},
			ver:       "1.1.1",
			path:      "",
		},
		"error case, wrong ver": {
			initramfs: []string{"/boot/initrd-1.1.1.aarch64.img"},
			ver:       "1.1.2",
			path:      "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll(t)
			for _, init := range tt.initramfs {
				env.CreateFile(t, filepath.Join("/var/lib/warewulf/chroots/image/rootfs", init))
			}

			initramfs := FindInitramfs("image", tt.ver)
			if tt.path == "" {
				assert.Nil(t, initramfs)
			} else {
				assert.NotNil(t, initramfs)
				assert.Equal(t, tt.path, initramfs.Path)
			}
		})
	}
}
