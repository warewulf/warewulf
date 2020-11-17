package overlay

import (
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/errors"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"io/ioutil"
	"os"
)

func FindAllRuntimeOverlays() ([]string, error) {
	config := config.New()
	var ret []string

	wwlog.Printf(wwlog.DEBUG, "Looking for runtime overlays...")
	files, err := ioutil.ReadDir(config.RuntimeOverlayDir())
	if err != nil {
		return ret, err
	}

	for _, file := range files {
		wwlog.Printf(wwlog.DEBUG, "Evaluating runtime overlay: %s\n", file.Name())
		if file.IsDir() == true {
			ret = append(ret, file.Name())
		}
	}

	return ret, nil
}


func RuntimeOverlayInit(name string) error {
	config := config.New()

	path := config.RuntimeOverlaySource(name)

	if util.IsDir(path) == true {
		return errors.New("Runtime overlay already exists: "+name)
	}

	err := os.MkdirAll(path, 0755)

	return err
}