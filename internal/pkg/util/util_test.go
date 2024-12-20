package util

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

func Test_FindFiles(t *testing.T) {
	var tests = map[string]struct {
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
			defer env.RemoveAll(t)
			env.MkdirAll(t, "/test")
			for _, file_ := range tt.createFiles {
				env.CreateFile(t, filepath.Join("/test", file_))
			}

			files := FindFiles(env.GetPath("/test"))
			assert.Equal(t, tt.findFiles, files)
		})
	}
}

func Test_FindFilterFiles(t *testing.T) {
	var tests = map[string]struct {
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
			defer env.RemoveAll(t)
			env.MkdirAll(t, "/test")
			for _, file_ := range tt.createFiles {
				env.CreateFile(t, filepath.Join("/test", file_))
			}

			files, err := FindFilterFiles(env.GetPath("/test"), tt.include, tt.exclude, true)
			assert.NoError(t, err)
			assert.Equal(t, tt.findFiles, files)
		})
	}
}
