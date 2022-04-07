package warewulfconf

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"path"

	"github.com/brotherpowers/ipsubnet"
	"github.com/creasty/defaults"
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
	var ret ControllerConf
	var warewulfconf WarewulfConf
	var dhpdconf DhcpConf
	var tftpconf TftpConf
	var nfsConf NfsConf
	ret.Warewulf = &warewulfconf
	ret.Dhcp = &dhpdconf
	ret.Tftp = &tftpconf
	ret.Nfs = &nfsConf
	err := defaults.Set(&ret)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Coult initialize default variables\n")
		return ret, err
	}
	// Check if cached config is old before re-reading config file
	if !cachedConf.current {
		wwlog.Printf(wwlog.DEBUG, "Opening Warewulf configuration file: %s\n", ConfigFile)
		data, err := ioutil.ReadFile(ConfigFile)
		if err != nil {
			wwlog.Printf(wwlog.WARN, "Error reading Warewulf configuration file\n")
		}

		wwlog.Printf(wwlog.DEBUG, "Unmarshaling the Warewulf configuration\n")
		err = yaml.Unmarshal(data, &ret)
		if err != nil {
			return ret, err
		}

		if ret.Ipaddr == "" || ret.Netmask == "" {
			conn, error := net.Dial("udp", "8.8.8.8:80")
			if error != nil {
				return ret, err
			}
			defer conn.Close()
			localIp := conn.LocalAddr().(*net.UDPAddr)
			if ret.Ipaddr == "" {
				ret.Ipaddr = localIp.IP.String()
				wwlog.Printf(wwlog.WARN, "IP address is not configured in warewulfd.conf, using %s\n", ret.Ipaddr)
			}
			if ret.Netmask == "" {
				mask := localIp.IP.DefaultMask()
				ret.Netmask = fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3])
				wwlog.Printf(wwlog.WARN, "Netmask address is not configured in warewulfd.conf, using %s\n", ret.Netmask)
			}
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
