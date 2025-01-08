package container

import (
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"

	"github.com/stretchr/testify/assert"

	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

func Test_ImportContainerDir(t *testing.T) {
	var tests = map[string]struct {
		files   []string
		sockets []string
	}{
		"empty container": {
			files:   nil,
			sockets: nil,
		},
		"sockets": {
			files: nil,
			sockets: []string{
				"/tmp/testsocket",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			src := "/tmp/testcontainer"
			env.CreateFile(filepath.Join(src, "/bin/sh"))
			for _, file := range tt.files {
				env.CreateFile(filepath.Join(src, file))
			}
			for _, socket := range tt.sockets {
				env.MkdirAll(filepath.Dir(filepath.Join(src, socket)))
				assert.NoError(t, unix.Mknod(env.GetPath(filepath.Join(src, socket)), unix.S_IFSOCK|0777, 0))
			}
			assert.NoError(t, ImportDirectory(env.GetPath(src), "testcontainer"))
			rootfs := "/var/lib/warewulf/chroots/testcontainer/rootfs"
			for _, file := range tt.files {
				assert.True(t, util.IsFile(env.GetPath(filepath.Join(rootfs, file))))
			}
			for _, socket := range tt.sockets {
				assert.True(t, util.IsFile(env.GetPath(filepath.Join(rootfs, socket))))
			}
		})
	}
}
