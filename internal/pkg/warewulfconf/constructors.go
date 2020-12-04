package warewulfconf

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func New() (ControllerConf, error) {
	var ret ControllerConf

	wwlog.Printf(wwlog.DEBUG, "Opening Warewulf configuration file: %s\n", ConfigFile)
	data, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		fmt.Printf("error reading node configuration file\n")
		return ret, err
	}

	wwlog.Printf(wwlog.DEBUG, "Unmarshaling the Warewulf configuration\n")
	err = yaml.Unmarshal(data, &ret)
	if err != nil {
		return ret, err
	}

	if ret.Warewulf.Port == 0 {
		ret.Warewulf.Port = 9873
	}

	wwlog.Printf(wwlog.DEBUG, "Returning node object\n")

	return ret, nil
}
