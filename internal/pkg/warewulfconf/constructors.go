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

		wwlog.Printf(wwlog.DEBUG, "Unmarshalling the Warewulf configuration\n")
		if err = yaml.Unmarshal(data, &ret); err != nil {
		    return ret, err
        }

		if ret.Ipaddr == "" && ret.HostBinding == "" {
			wwlog.Printf(wwlog.WARN, "IP address is not configured in warewulfd.conf\n")
		}
		if ret.Netmask == "" && ret.HostBinding == "" {
			wwlog.Printf(wwlog.WARN, "Netmask is not configured in warewulfd.conf\n")
		}

		if ret.Network == "" {
		    var deviceIP string
		    if ret.HostBinding != "" {
                if deviceIP, err = util.HostnameToV4(ret.HostBinding); err != nil {
                    wwlog.Printf(wwlog.WARN, "Error converting HostBinding to ipv4: %v\n", err)
                    return ret, err
                }
                if len(ret.Ipaddr)+len(ret.Netmask) > 0 {
                    if !net.ParseIP(deviceIP).Equal(net.ParseIP(ret.Ipaddr)) {
                        wwlog.Printf(wwlog.WARN, "IP resolved from device hostname (%s) does not match '%s'\n", deviceIP, ret.Ipaddr)
                        return ret, fmt.Errorf("IP resolved from device hostname (%s) does not match '%s'", deviceIP, ret.Ipaddr)
                    }
                } else {
                    ret.Ipaddr = net.ParseIP(deviceIP).String()
                }
            }

            Addresses, err := util.GetLocalAddresses()
            if err != nil {
                wwlog.Printf(wwlog.WARN, "Error getting local addresses: %v\n", err)
                return ret, err
            }

            if _, ok := Addresses[ret.Ipaddr]; !ok {
                wwlog.Printf(wwlog.WARN, "Specified IP could not be found locally\n")
                return ret, fmt.Errorf("specified IP could not be found locally")
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
