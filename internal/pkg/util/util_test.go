package util

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
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

			files := FindFiles(env.GetPath("/test"))
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

			files, err := FindFilterFiles(env.GetPath("/test"), tt.include, tt.exclude, true)
			assert.NoError(t, err)
			assert.Equal(t, tt.findFiles, files)
		})
	}
}

func Test_Overwrite(t *testing.T) {
	t.Run("overwrite a file", func(t *testing.T) {
		env := testenv.New(t)
		defer env.RemoveAll()
		env.CreateFile("file")
		env.WriteFile("file", "hello world")

		assert.Equal(t, "hello world", env.ReadFile("file"))

		err := OverwriteFile(env.GetPath("file"), []byte("hello warewulf"))
		assert.NoError(t, err)

		assert.Equal(t, "hello warewulf", env.ReadFile("file"))
	})
}
