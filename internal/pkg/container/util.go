package container

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func ValidName(name string) bool {
	if !util.ValidString(name, "^[\\w\\-\\.\\:]+$") {
		wwlog.Printf(wwlog.WARN, "VNFS name has illegal characters: %s\n", name)
		return false
	}
	return true
}

func ListSources() ([]string, error) {
	var ret []string

	err := os.MkdirAll(SourceParentDir(), 0755)
	if err != nil {
		return ret, errors.New("Could not create VNFS source parent directory: " + SourceParentDir())
	}
	wwlog.Printf(wwlog.DEBUG, "Searching for VNFS Rootfs directories: %s\n", SourceParentDir())

	sources, err := ioutil.ReadDir(SourceParentDir())
	if err != nil {
		return ret, err
	}

	for _, source := range sources {
		wwlog.Printf(wwlog.VERBOSE, "Found VNFS source: %s\n", source.Name())

		if !ValidName(source.Name()) {
			continue
		}

		if !ValidSource(source.Name()) {
			continue
		}

		ret = append(ret, source.Name())
	}

	return ret, nil
}

func ValidSource(name string) bool {
	fullPath := RootFsDir(name)

	if !ValidName(name) {
		return false
	}

	if !util.IsDir(fullPath) {
		wwlog.Printf(wwlog.VERBOSE, "Location is not a VNFS source directory: %s\n", name)
		return false
	}

	return true
}

func DeleteSource(name string) error {
	fullPath := SourceDir(name)

	wwlog.Printf(wwlog.VERBOSE, "Removing path: %s\n", fullPath)
	return os.RemoveAll(fullPath)
}
