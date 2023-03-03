package warewulfconf

import (
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/brotherpowers/ipsubnet"
	"github.com/creasty/defaults"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v2"
)

var cachedConf ControllerConf

var ConfigFile string

/*
Creates a new empty ControllerConf object, returns a cached
one if called in a nother context.
*/
func New() (ret ControllerConf) {
	// NOTE: This function can be called before any log level is set
	//       so using wwlog.Verbose or wwlog.Debug won't work
	if !cachedConf.current {
		ret.Warewulf = new(WarewulfConf)
		ret.Dhcp = new(DhcpConf)
		ret.Tftp = new(TftpConf)
		ret.Nfs = new(NfsConf)
		ret.Paths = new(BuildConfig)
		_ = defaults.Set(&ret)
		ret.setDynamicDefaults()

		cachedConf = ret
		cachedConf.current = true

	} else {
		// If cached struct isn't empty, use it as the return value
		ret = cachedConf
	}
	return ret
}

/*
Populate the configuration with the values from the configuration file.
*/
func (conf *ControllerConf) ReadConf(confFileName string) (err error) {
	fileHandle, err := os.ReadFile(confFileName)
	if err != nil {
		return err
	}
	return conf.Read(fileHandle)
}

/*
Populate the configuration with the values from the given yaml information
*/
func (conf *ControllerConf) Read(data []byte) (err error) {
	// ipxe binaries are merged not overwritten, store defaults separate
	defIpxe := make(map[string]string)
	for k, v := range conf.Tftp.IpxeBinaries {
		defIpxe[k] = v
		delete(conf.Tftp.IpxeBinaries, k)
	}
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		return
	}
	if len(conf.Tftp.IpxeBinaries) == 0 {
		conf.Tftp.IpxeBinaries = defIpxe
	}
	cachedConf = *conf
	cachedConf.current = true
	return
}

/*
Set the runtime defaults like IP address of running system to the config
*/
func (ret *ControllerConf) setDynamicDefaults() (err error) {
	if ret.Ipaddr == "" || ret.Netmask == "" {
		conn, error := net.Dial("udp", "8.8.8.8:80")
		if error != nil {
			return err
		}
		defer conn.Close()
		localIp := conn.LocalAddr().(*net.UDPAddr)
		if ret.Ipaddr == "" {
			ret.Ipaddr = localIp.IP.String()
			wwlog.Verbose("IP address is not configured in warewulfd.conf, using %s", ret.Ipaddr)
		}
		if ret.Netmask == "" {
			mask := localIp.IP.DefaultMask()
			ret.Netmask = fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3])
			wwlog.Verbose("Netmask address is not configured in warewulfd.conf, using %s", ret.Netmask)
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
			wwlog.Error("Invalid ipv6 address specified, mut be CIDR notation: %s", ret.Ipaddr6)
			return errors.New("invalid ipv6 network")
		}
		if msize, _ := ipv6net.Mask.Size(); msize > 64 {
			wwlog.Error("ipv6 mask size must be smaller than 64")
			return errors.New("invalid ipv6 network size")
		}
	}
	return
}
