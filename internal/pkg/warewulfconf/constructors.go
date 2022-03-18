package warewulfconf

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"path"

	"github.com/brotherpowers/ipsubnet"
	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"

	"gopkg.in/yaml.v2"
)

var cachedConf ControllerConf

var ConfigFile string

func init() {
	if ConfigFile == "" {
		ConfigFile = path.Join(buildconfig.SYSCONFDIR(), "warewulf/warewulf.conf")
	}
}

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
		// check validity of ipv6 net
		if ret.Ipaddr6 != "" {
			_, ipv6net, err := net.ParseCIDR(ret.Ipaddr6)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "Invalid ipv6 address specified, mut be CIDR notation: %s\n", ret.Ipaddr6)
				return ret, errors.New("invalid ipv6 network")
			}
			if msize, _ := ipv6net.Mask.Size(); msize > 64 {
				wwlog.Printf(wwlog.ERROR, "ipv6 mask size must be smaller than 64\n")
				return ret, errors.New("invalid ipv6 network size")
			}
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
