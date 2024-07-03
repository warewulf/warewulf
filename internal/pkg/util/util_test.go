package util

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func TryCreatePath(t *testing.T, elem ...string) {
	err := os.MkdirAll(filepath.Join(elem...), os.ModePerm)
	if err != nil {
		t.Errorf("Failed creating dir: %v", err)
		t.FailNow()
	}
}

func Test_FindFilterFiles(t *testing.T) {
	wwlog.SetLogLevel(wwlog.DEBUG)
	dir, err := os.MkdirTemp(os.TempDir(), "warewulf-test")
	if err != nil {
		t.Errorf("Failed creating tmpdir: %v", err)
		t.FailNow()
	}
	defer os.RemoveAll(dir)
	TryCreatePath(t, dir, "boot")
	TryCreatePath(t, dir, "usr", "local")
	TryCreatePath(t, dir, "usr", "bin")
	TryCreatePath(t, dir, "usr", "usr", "local")
	TryCreatePath(t, dir, "bin")
	TryCreatePath(t, dir, "lib")

	assert.NoError(t, os.Symlink("/path/to/target", filepath.Join(dir, "symlink")))

	files, err := FindFilterFiles(dir, []string{"boot", "usr", "bin", "symlink"}, []string{"/b*/", "/usr/local"}, true)

	if err != nil {
		t.Errorf("FindFilerFiles failed: %v", err)
		t.FailNow()
	}

	expected := []string{"usr", "usr/bin", "usr/usr", "usr/usr/local", "symlink"}
	sort.Strings(expected)
	sort.Strings(files)
	if !reflect.DeepEqual(files, expected) {
		t.Errorf("expected %v, got %v", expected, files)
		t.FailNow()
	}
}
