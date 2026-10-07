package util_test

import (
	"encoding/hex"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/util"

	"github.com/pkg/xattr"
)

func Test_FindFiles(t *testing.T) {
	tests := map[string]struct {
		createFiles []string
		findFiles   []string
	}{
		"no files": {
			createFiles: []string{},
			findFiles:   nil,
		},
		"single file": {
			createFiles: []string{"testfile"},
			findFiles:   []string{"testfile"},
		},
		"nested file": {
			createFiles: []string{"testdir/testfile"},
			findFiles:   []string{"testdir/", "testdir/testfile"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			env.MkdirAll("/test")
			for _, file_ := range tt.createFiles {
				env.CreateFile(filepath.Join("/test", file_))
			}

			files := util.FindFiles(env.GetPath("/test"))
			assert.Equal(t, tt.findFiles, files)
		})
	}
}

func Test_FindFilterFiles(t *testing.T) {
	tests := map[string]struct {
		createFiles []string
		include     []string
		exclude     []string
		findFiles   []string
	}{
		"no files": {
			createFiles: []string{},
			include:     []string{"*"},
			findFiles:   nil,
		},
		"single file": {
			createFiles: []string{"testfile"},
			include:     []string{"*"},
			findFiles:   []string{"testfile"},
		},
		"nested file": {
			createFiles: []string{"testdir/testfile"},
			include:     []string{"*"},
			findFiles:   []string{"testdir", "testdir/testfile"},
		},
		"multiple files": {
			createFiles: []string{"test1/testfile", "test2/testfile"},
			include:     []string{"*"},
			findFiles:   []string{"test1", "test1/testfile", "test2", "test2/testfile"},
		},
		"excluded files": {
			createFiles: []string{"test1/test1", "test1/test2", "test2/test1", "test2/test2"},
			include:     []string{"*"},
			exclude:     []string{"test1/*2", "test2"},
			findFiles:   []string{"test1", "test1/test1"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			env.MkdirAll("/test")
			for _, file_ := range tt.createFiles {
				env.CreateFile(filepath.Join("/test", file_))
			}

			files, err := util.FindFilterFiles(env.GetPath("/test"), tt.include, tt.exclude, true)
			assert.NoError(t, err)
			assert.Equal(t, tt.findFiles, files)
		})
	}
}

func Test_SGetXattrsR(t *testing.T) {
	currentUser, err := user.Current()
	if currentUser.Uid != "0" || err != nil {
		t.Skip("Skipping test, security.capability xattr can only be set when running as root")
	}
	tests := map[string]struct {
		fileXattrs map[string]string
	}{
		"no xattrs": {
			fileXattrs: map[string]string{"test/test1": "", "test/test2": ""},
		},
		"xattrs": {
			fileXattrs: map[string]string{"test/test1": "security.capability=0x0000000200200000000000000000000000000000", "test/test2": "system.posix_acl_access=0x0200000001000600ffffffff02000600e903000004000400ffffffff10000600ffffffff20000400ffffffff"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mask := "security\\.capability|system\\.posix_acl.*"
			env := testenv.New(t)
			defer env.RemoveAll()
			env.MkdirAll("/test")
			for file_, xattr_ := range tt.fileXattrs {
				env.CreateFile(filepath.Join("/", file_))
				if xattr_ == "" {
					break
				}
				xattr_ := strings.SplitN(xattr_, "=", 2)
				xattrName := xattr_[0]
				xattrValue := xattr_[1]
				xattrHex, _ := hex.DecodeString(xattrValue[2:])
				err := xattr.LSet(env.GetPath("/"+file_), xattrName, xattrHex)
				assert.NoError(t, err)
			}
			xattrs, err := util.SGetXattrsR(env.GetPath("/"), mask)
			assert.NoError(t, err)
			for file_, xattr_ := range tt.fileXattrs {
				if xattr_ == "" {
					assert.Nil(t, xattrs)
				} else {
					assert.Contains(t, xattrs, "# file: "+file_+"\n"+tt.fileXattrs[file_]+"\n\n")
				}
			}
		})
	}
}
func Test_CopyXattrs(t *testing.T) {
	t.Run("copy xattrs", func(t *testing.T) {
		currentUser, err := user.Current()
		if currentUser.Uid != "0" || err != nil {
			t.Skip("Skipping test, security.capability xattr can only be set when running as root")
		}
		xattrs := map[string][]byte{"security.capability": []byte{0x0, 0x0, 0x0, 0x2, 0x0, 0x20, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, "system.posix_acl_access": []byte{0x2, 0x0, 0x0, 0x0, 0x1, 0x0, 0x6, 0x0, 0xff, 0xff, 0xff, 0xff, 0x2, 0x0, 0x6, 0x0, 0xe9, 0x3, 0x0, 0x0, 0x4, 0x0, 0x4, 0x0, 0xff, 0xff, 0xff, 0xff, 0x10, 0x0, 0x6, 0x0, 0xff, 0xff, 0xff, 0xff, 0x20, 0x0, 0x4, 0x0, 0xff, 0xff, 0xff, 0xff}}
		files := []string{"test1", "test2"}
		dirs := []string{"/src", "/dest"}
		env := testenv.New(t)
		defer env.RemoveAll()
		env.MkdirAll("/src")
		for _, file := range files {
			env.CreateFile(filepath.Join("/src", file))
		}
		env.MkdirAll("/dest")
		for xattrName, value := range xattrs {
			err := xattr.LSet(env.GetPath(filepath.Join(dirs[0], files[0])), xattrName, value)
			assert.NoError(t, err)
		}
		err = xattr.LSet(env.GetPath(filepath.Join(dirs[0], files[1])), "system.posix_acl_access", xattrs["system.posix_acl_access"])
		assert.NoError(t, err)

		err = exec.Command("cp", "-r", env.GetPath(dirs[0])+"/.", env.GetPath(dirs[1])).Run()
		assert.NoError(t, err)
		for _, file := range files {
			err := util.CopyXattrs(env.GetPath(filepath.Join(dirs[0], file)), env.GetPath(filepath.Join(dirs[1], file)))
			assert.NoError(t, err)
		}
		for xattrName, value := range xattrs {
			r, err := xattr.LGet(env.GetPath(filepath.Join(dirs[1], files[0])), xattrName)
			assert.NoError(t, err)
			assert.Equal(t, value, r)
		}
		r, err := xattr.LGet(env.GetPath(filepath.Join(dirs[1], files[1])), "system.posix_acl_access")
		assert.NoError(t, err)
		assert.Equal(t, r, xattrs["system.posix_acl_access"])
	})
}
