package overlay

import (
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/errors"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"io/ioutil"
	"os"
)

func FindAllSystemOverlays() ([]string, error) {
	config := config.New()
	var ret []string

	wwlog.Printf(wwlog.DEBUG, "Looking for system overlays...")
	files, err := ioutil.ReadDir(config.SystemOverlayDir())
	if err != nil {
		return ret, err
	}

	for _, file := range files {
		wwlog.Printf(wwlog.DEBUG, "Evaluating system overlay: %s\n", file.Name())
		if file.IsDir() == true {
			ret = append(ret, file.Name())
		}
	}

	return ret, nil
}


func SystemOverlayInit(name string) error {
	config := config.New()

	path := config.SystemOverlaySource(name)

	if util.IsDir(path) == true {
		return errors.New("Runtime overlay already exists: "+name)
	}

	err := os.MkdirAll(path, 0755)

	return err
}