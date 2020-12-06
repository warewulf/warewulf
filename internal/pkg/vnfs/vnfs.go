package vnfs

import (
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/errors"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"io/ioutil"
	"os"
	"path"
)

func ValidName(name string) bool {
	if util.ValidString(name, "^[a-zA-Z0-9.:-]+$") == false {
		wwlog.Printf(wwlog.WARN, "VNFS name has illegal characters: %s\n", name)
		return false
	}
	return true
}

func SourceParentDir() string {
	return path.Join(config.LocalStateDir, "chroots")
}

func SourceDir(name string) string {
	return path.Join(SourceParentDir(), name)
}

func RootFsDir(name string) string {
	return path.Join(SourceDir(name), "rootfs")
}

func ImageParentDir() string {
	return path.Join(config.LocalStateDir, "provision/vnfs/")
}

func ImageFile(name string) string {
	return path.Join(ImageParentDir(), name+".img.gz")
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

		if ValidName(source.Name()) == false {
			continue
		}

		if ValidSource(source.Name()) == false {
			continue
		}

		ret = append(ret, source.Name())
	}

	return ret, nil
}

func ValidSource(name string) bool {
	fullPath := RootFsDir(name)

	if ValidName(name) == false {
		return false
	}

	if util.IsDir(fullPath) == false {
		wwlog.Printf(wwlog.VERBOSE, "Location is not a VNFS source directory: %s\n", name)
		return false
	}

	if util.IsFile(path.Join(fullPath, "/sbin/init")) == false {
		wwlog.Printf(wwlog.VERBOSE, "VNFS Source does not have a valid /sbin/init: %s\n", name)
		return false
	}

	return true
}
