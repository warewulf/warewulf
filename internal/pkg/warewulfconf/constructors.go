package warewulfconf

import (
	"fmt"
	"io/ioutil"
	"net"
	"errors"

	"github.com/brotherpowers/ipsubnet"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"

	"gopkg.in/yaml.v2"
)

var cachedConf ControllerConf

func New() (ControllerConf, error) {
	var ret ControllerConf = *defaultConfig()

	// Check if cached config is old before re-reading config file
	if !cachedConf.current {
		wwlog.Printf(wwlog.DEBUG, "Opening Warewulf configuration file: %s\n", ConfigFile)
		data, err := ioutil.ReadFile(ConfigFile)
		if err != nil {
			fmt.Printf("Error reading Warewulf configuration file\n")
			return ret, err
		}

		wwlog.Printf(wwlog.DEBUG, "Unmarshaling the Warewulf configuration\n")
		err = yaml.Unmarshal(data, &ret)
		if err != nil {
			return ret, err
		}

// TODO: Need to add comprehensive config file validator
// TODO: Change function to guess default IP address and/or mask from local system
		if ret.Ipaddr == "" {
			wwlog.Printf(wwlog.ERROR, "IP address is not configured in warewulfd.conf\n")
			return ret, errors.New("no IP Address")
		}

		if ret.Netmask == "" {
			wwlog.Printf(wwlog.ERROR, "Netmask is not configured in warewulfd.conf\n")
			return ret, errors.New("no netmask")
		}

		if ret.Network == "" {
			mask := net.IPMask(net.ParseIP(ret.Netmask).To4())
			size, _ := mask.Size()

			sub := ipsubnet.SubnetCalculator(ret.Ipaddr, size)

			ret.Network = sub.GetNetworkPortion()
		}

		if ret.Warewulf.Port == 0 {
			ret.Warewulf.Port = defaultPort
		}

		wwlog.Printf(wwlog.DEBUG, "Returning warewulf config object\n")
		cachedConf = ret
		cachedConf.current = true

	} else {
		wwlog.Printf(wwlog.DEBUG, "Returning cached warewulf config object\n")
		// If cached struct isn't empty, use it as the return value 
		ret = cachedConf
	}

	return ret, nil
}
