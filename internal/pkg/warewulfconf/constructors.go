package warewulfconf

import (
	"fmt"
	"io/ioutil"
	"net"

	"github.com/brotherpowers/ipsubnet"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v2"
)

var singleton ControllerConf

func New() (ControllerConf, error) {
	var ret ControllerConf

	if (ControllerConf{}) == singleton {
		wwlog.Printf(wwlog.DEBUG, "Opening Warewulf configuration file: %s\n", ConfigFile)
		data, err := ioutil.ReadFile(ConfigFile)
		if err != nil {
			fmt.Printf("error reading Warewulf configuration file\n")
			return ret, err
		}

		wwlog.Printf(wwlog.DEBUG, "Unmarshaling the Warewulf configuration\n")
		err = yaml.Unmarshal(data, &ret)
		if err != nil {
			return ret, err
		}

		if ret.Ipaddr == "" {
			wwlog.Printf(wwlog.WARN, "IP address is not configured in warewulfd.conf\n")
		}
		if ret.Netmask == "" {
			wwlog.Printf(wwlog.WARN, "Netmask is not configured in warewulfd.conf\n")
		}

		if ret.Network == "" {
			mask := net.IPMask(net.ParseIP(ret.Netmask).To4())
			size, _ := mask.Size()

			sub := ipsubnet.SubnetCalculator(ret.Ipaddr, size)

			ret.Network = sub.GetNetworkPortion()
		}

		if ret.Warewulf.Port == 0 {
			ret.Warewulf.Port = 9873
		}

		wwlog.Printf(wwlog.DEBUG, "Returning warewulf config object\n")
		singleton = ret

	} else {
		wwlog.Printf(wwlog.DEBUG, "Returning cached warewulf config object\n")

		ret = singleton
	}

	return ret, nil
}
