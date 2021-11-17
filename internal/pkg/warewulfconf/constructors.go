package warewulfconf

import (
	"fmt"
	"github.com/brotherpowers/ipsubnet"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
)

var singleton ControllerConf

func New() (ret ControllerConf, err error) {

	if (ControllerConf{}) == singleton {
		wwlog.Printf(wwlog.DEBUG, "Opening Warewulf configuration file: %s\n", ConfigFile)
		data, err := ioutil.ReadFile(ConfigFile)
		if err != nil {
			fmt.Printf("error reading Warewulf configuration file\n")
			return ret, err
		}

		wwlog.Printf(wwlog.DEBUG, "Unmarshaling the Warewulf configuration\n")
		if err = yaml.Unmarshal(data, &ret); err != nil {
			return ret, err
		}

		if ret.Ipaddr == "" {
			wwlog.Printf(wwlog.WARN, "IP address is not configured in warewulfd.conf\n")
		}
		if ret.Netmask == "" {
			wwlog.Printf(wwlog.WARN, "Netmask is not configured in warewulfd.conf\n")
		}

		switch {
		case ret.DeviceName != "":
			iface, err := net.InterfaceByName(ret.DeviceName)
			if err != nil {
				return ret, err
			}
			ipv4, err := util.GetFirstIPv4(iface)
			if err != nil {
				return ret, err
			}
			ret.Ipaddr = ipv4.To4().String()
			mask := net.IPMask(net.ParseIP(ret.Netmask).To4())
			size, _ := mask.Size()

			sub := ipsubnet.SubnetCalculator(ret.Ipaddr, size)

			ret.Network = sub.GetNetworkPortion()
		case ret.Network == "":
			ip := net.ParseIP(ret.Ipaddr)
			if ip == nil {
				if ip, err = util.HostnameToV4(ret.Ipaddr); err != nil {
					return ret, err
				}
				ret.Ipaddr = ip.String()
				wwlog.Printf(wwlog.DEBUG, "Resolved DNS name '%s' to ipv4: '%s'\n", ret.Ipaddr, ip.String())
			}

			if !util.AddressExists(ip) {
				return ret, fmt.Errorf("Address '%s' does not exist on any local interface\n", ret.Ipaddr)
			}

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
