package kernel

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

func Test_FindKernel(t *testing.T) {
	var tests = map[string]struct {
		files   []string
		version string
		path    string
	}{
		"/boot/vmlinuz-* (1)": {
			files: []string{
				"/boot/vmlinuz-5.14.0-427.18.1.el9_4.x86_64",
				"/boot/vmlinuz-5.14.0-427.24.1.el9_4.x86_64",
				"/boot/vmlinuz-4.14.0-427.18.1.el8_4.x86_64",
			},
			version: "5.14.0-427.24.1",
			path:    "/boot/vmlinuz-5.14.0-427.24.1.el9_4.x86_64",
		},
		"/boot/vmlinuz-* (2)": {
			files: []string{
				"/boot/vmlinuz-5.15.0-119-generic",
				"/boot/vmlinuz-5.14.0-427.24.1.el9_4.x86_64",
				"/boot/vmlinuz-6.15.0-119-generic",
			},
			version: "6.15.0-119",
			path:    "/boot/vmlinuz-6.15.0-119-generic",
		},
		"/boot/vmlinuz-* (3)": {
			files: []string{
				"/boot/vmlinuz-5.15.0-0-vanilla",
				"/boot/vmlinuz-5.14.0-427.24.1.el9_4.x86_64",
			},
			version: "5.15.0-0",
			path:    "/boot/vmlinuz-5.15.0-0-vanilla",
		},
		"/boot/vmlinuz-* (4)": {
			files: []string{
				"/boot/vmlinuz-5.15.0-generic",
				"/boot/vmlinuz-5.14.0-427.24.1.el9_4.x86_64",
				"/boot/vmlinuz-5.13.0-427.24.1.el9_4.x86_64",
			},
			version: "5.15.0",
			path:    "/boot/vmlinuz-5.15.0-generic",
		},
		"/lib/modules/*/vmlinuz": {
			files: []string{
				"/lib/modules/5.14.0-427.18.1.el9_4.x86_64/vmlinuz",
				"/lib/modules/5.14.0-427.24.1.el9_4.x86_64/vmlinuz",
			},
			version: "5.14.0-427.24.1",
			path:    "/lib/modules/5.14.0-427.24.1.el9_4.x86_64/vmlinuz",
		},
		"/boot/vmlinuz-*.gz": {
			files: []string{
				"/boot/vmlinuz-5.14.0-427.18.1.el9_4.x86_64.gz",
				"/boot/vmlinuz-5.14.0-427.24.1.el9_4.x86_64.gz",
			},
			version: "5.14.0-427.24.1",
			path:    "/boot/vmlinuz-5.14.0-427.24.1.el9_4.x86_64.gz",
		},
		"ignore rescue and debug kernels": {
			files: []string{
				"/boot/vmlinuz-0-rescue-eb46964329b146e39518c625feab3ea0",
				"/boot/vmlinuz-5.14.0-362.24.1.el9_3.aarch64",
				"/boot/vmlinuz-5.14.0-427.31.1.el9_4.aarch64+debug",
				"/boot/vmlinuz-5.14.0-284.30.1.el9_2.aarch64",
				"/boot/vmlinuz-5.14.0-427.31.1.el9_4.aarch64",
			},
			version: "5.14.0-427.31.1",
			path:    "/boot/vmlinuz-5.14.0-427.31.1.el9_4.aarch64",
		},
		"no kernels": {
			files:   []string{},
			version: "",
			path:    "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			rootfs := "/var/lib/warewulf/chroots/testimage/rootfs"
			for _, file := range tt.files {
				env.CreateFile(filepath.Join(rootfs, file))
			}

			kernels := FindKernels("testimage")
			assert.Equal(t, len(tt.files), len(kernels))
			kernel := kernels.Default()
			if tt.version == "" && tt.path == "" {
				assert.Nil(t, kernel)
			} else {
				assert.Equal(t, "testimage", kernel.ImageName)
				assert.Equal(t, tt.version, kernel.Version())
				assert.Equal(t, tt.path, kernel.Path)
				assert.Equal(t, env.GetPath(filepath.Join(rootfs, tt.path)), kernel.FullPath())
			}
		})
	}
}

func Test_FromNode(t *testing.T) {
	tests := map[string]struct {
		files   []string
		version string
		path    string
	}{
		"default": {
			files: []string{
				"/boot/vmlinuz-5.14.0-427.18.1.el9_4.x86_64",
				"/boot/vmlinuz-5.14.0-427.24.1.el9_4.x86_64",
				"/boot/vmlinuz-4.14.0-427.18.1.el8_4.x86_64",
			},
			version: "",
			path:    "/boot/vmlinuz-5.14.0-427.24.1.el9_4.x86_64",
		},
		"path": {
			files: []string{
				"/boot/vmlinuz-5.14.0-427.18.1.el9_4.x86_64",
				"/boot/vmlinuz-5.14.0-427.24.1.el9_4.x86_64",
				"/boot/vmlinuz-4.14.0-427.18.1.el8_4.x86_64",
			},
			version: "/boot/vmlinuz-4.14.0-427.18.1.el8_4.x86_64",
			path:    "/boot/vmlinuz-4.14.0-427.18.1.el8_4.x86_64",
		},
		"version": {
			files: []string{
				"/boot/vmlinuz-5.14.0-427.18.1.el9_4.x86_64",
				"/boot/vmlinuz-5.14.0-427.24.1.el9_4.x86_64",
				"/boot/vmlinuz-4.14.0-427.18.1.el8_4.x86_64",
			},
			version: "4.14.0-427.18.1",
			path:    "/boot/vmlinuz-4.14.0-427.18.1.el8_4.x86_64",
		},
		"none": {
			files:   []string{},
			version: "",
			path:    "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			rootfs := "/var/lib/warewulf/chroots/testimage/rootfs"
			for _, file := range tt.files {
				env.CreateFile(filepath.Join(rootfs, file))
			}
			node := node.EmptyNode()
			node.ImageName = "testimage"
			node.Kernel.Version = tt.version
			kernel := FromNode(&node)
			if tt.path == "" {
				assert.Nil(t, kernel)
			} else {
				assert.Equal(t, "testimage", kernel.ImageName)
				assert.Equal(t, tt.path, kernel.Path)
			}
		})
	}
}

func Test_FindAllKernels(t *testing.T) {
	tests := map[string]struct {
		files map[string][]string
		count int
	}{
		"two images": {
			files: map[string][]string{
				"image1": []string{
					"/boot/vmlinuz-5.14.0-427.18.1.el9_4.x86_64",
					"/boot/vmlinuz-5.14.0-427.24.1.el9_4.x86_64",
					"/boot/vmlinuz-4.14.0-427.18.1.el8_4.x86_64",
				},
				"image2": []string{
					"/boot/vmlinuz-0-rescue-eb46964329b146e39518c625feab3ea0",
					"/boot/vmlinuz-5.14.0-362.24.1.el9_3.aarch64",
					"/boot/vmlinuz-5.14.0-427.31.1.el9_4.aarch64+debug",
					"/boot/vmlinuz-5.14.0-284.30.1.el9_2.aarch64",
					"/boot/vmlinuz-5.14.0-427.31.1.el9_4.aarch64",
				},
			},
			count: 8,
		},
		"empty": {
			files: map[string][]string{},
			count: 0,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			for image, files := range tt.files {
				rootfs := filepath.Join(filepath.Join("/var/lib/warewulf/chroots", image), "rootfs")
				for _, file := range files {
					env.CreateFile(filepath.Join(rootfs, file))
				}
			}
			kernels := FindAllKernels()
			assert.Equal(t, tt.count, len(kernels))
		})
	}
}

func Test_IsDebugOrRescue(t *testing.T) {
	tests := map[string]struct {
		path   string
		debug  bool
		rescue bool
	}{
		"default": {
			path:   "/boot/vmlinuz-1.0.0",
			debug:  false,
			rescue: false,
		},
		"debug": {
			path:   "/boot/vmlinuz-5.14.0-427.31.1.el9_4.aarch64+debug",
			debug:  true,
			rescue: false,
		},
		"rescue": {
			path:   "/boot/vmlinuz-0-rescue-eb46964329b146e39518c625feab3ea0",
			debug:  false,
			rescue: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			kernel := &Kernel{ImageName: "", Path: tt.path}
			assert.Equal(t, tt.debug, kernel.IsDebug())
			assert.Equal(t, tt.rescue, kernel.IsRescue())
		})
	}
}
