package staticfiles

import (
	"io/ioutil"

	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
)

func WriteData(source string, dest string) error {
	bytes, err := getResource(source)
	if err != nil {
		return errors.Wrap(err, "failed to get resource")
	}

	err = ioutil.WriteFile(dest, bytes, 0644)
	if err != nil {
		// TODO: remove log message if appropriate
		wwlog.Printf(wwlog.ERROR, "Failed writing %s to: %s\n", dest, err)
		return errors.Wrap(err, "failed to write to file")
	}

	return nil
}
