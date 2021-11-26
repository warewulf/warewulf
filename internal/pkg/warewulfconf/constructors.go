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

        if (ret.Ipaddr == "" || ret.Netmask == "") && ret.DeviceName == "" {
            return ret, fmt.Errorf("'[ipaddr/hostname]:netmask' or a standalone device name are not configured in warewulfd.conf")
        }

        switch {
        case ret.Network == "" && ret.DeviceName == "":
            ip := net.ParseIP(ret.Ipaddr).To4()
            if ip == nil {
                if ip, err = util.HostnameToV4(ret.Ipaddr); err != nil {
                    return ret, err
                }
                wwlog.Printf(wwlog.DEBUG, "Resolved DNS name '%s' to ipv4: '%s'\n", ret.Ipaddr, ip.String())
            }

            ret.Ipaddr = ip.String()
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
            ret.Netmask = ipv4.To4().DefaultMask().String()
        default:
            return ret, fmt.Errorf("")
        }

        mask := net.IPMask(net.ParseIP(ret.Netmask).To4())
        size, _ := mask.Size()

        sub := ipsubnet.SubnetCalculator(ret.Ipaddr, size)

        ret.Network = sub.GetNetworkPortion()

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
