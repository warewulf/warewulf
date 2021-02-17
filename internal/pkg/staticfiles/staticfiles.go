package staticfiles

import (
	"io/ioutil"

	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func WriteData(source string, dest string) error {
	bytes, err := getResource(source)
	err = ioutil.WriteFile(dest, bytes, 0644)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Failed writing %s to: %s\n", dest, err)
		return err
	}
	return nil
}
