package warewulfconf

import (
	"fmt"
	"net"
	"os"

	"github.com/pkg/errors"

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
func New() (conf ControllerConf) {
	// NOTE: This function can be called before any log level is set
	//       so using wwlog.Verbose or wwlog.Debug won't work
	if !cachedConf.current {
		conf.Warewulf = new(WarewulfConf)
		conf.Dhcp = new(DhcpConf)
		conf.Tftp = new(TftpConf)
		conf.Nfs = new(NfsConf)
		conf.Paths = new(BuildConfig)
		_ = defaults.Set(&conf)

		cachedConf = conf
		cachedConf.current = true

	} else {
		// If cached struct isn't empty, use it as the return value
		conf = cachedConf
	}
	return conf
}

/*
Populate the configuration with the values from the configuration file.
*/
func (conf *ControllerConf) ReadConf(confFileName string) (err error) {
	wwlog.Debug("Reading warewulf.conf from: %s", confFileName)
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
	err = conf.SetDynamicDefaults()
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
func (conf *ControllerConf) SetDynamicDefaults() (err error) {
	if conf.Ipaddr == "" || conf.Netmask == "" || conf.Network == "" {
		var mask net.IPMask
		var network *net.IPNet
		var ipaddr net.IP

		if conf.Ipaddr == "" {
			wwlog.Verbose("Configuration has no valid network, going to dynamic values")
			conn, _ := net.Dial("udp", "8.8.8.8:80")
			defer conn.Close()
			ipaddr = conn.LocalAddr().(*net.UDPAddr).IP
			mask = ipaddr.DefaultMask()
			sz, _ := mask.Size()
			conf.Ipaddr = ipaddr.String() + fmt.Sprintf("/%d", sz)
		}
		_, network, err = net.ParseCIDR(conf.Ipaddr)
		if err == nil {
			mask = network.Mask
			fmt.Println(mask)
		} else {
			return errors.Wrap(err, "Couldn't parse IP address")
		}
		if conf.Netmask == "" {
			conf.Netmask = fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3])
			wwlog.Verbose("Netmask address is not configured in warewulf.conf, using %s", conf.Netmask)
		}
		if conf.Network == "" {
			conf.Network = network.IP.String()
			wwlog.Verbose("Network is not configured in warewulf.conf, using %s", conf.Network)
		}
	}
	if conf.Dhcp.RangeStart == "" && conf.Dhcp.RangeEnd == "" {
		start := net.ParseIP(conf.Network).To4()
		start[3] += 1
		if start.Equal(net.ParseIP(conf.Ipaddr)) {
			start[3] += 1
		}
		conf.Dhcp.RangeStart = start.String()
		wwlog.Verbose("dhpd start is not configured in warewulf.conf, using %s", conf.Dhcp.RangeStart)
		sz, _ := net.IPMask(net.ParseIP(conf.Netmask).To4()).Size()
		range_end := (1 << (32 - sz)) / 8
		if range_end > 127 {
			range_end = 127
		}
		end := net.ParseIP(conf.Network).To4()
		end[3] += byte(range_end)
		conf.Dhcp.RangeEnd = end.String()
		wwlog.Verbose("dhpd end is not configured in warewulf.conf, using %s", conf.Dhcp.RangeEnd)

	}
	// check validity of ipv6 net
	if conf.Ipaddr6 != "" {
		_, ipv6net, err := net.ParseCIDR(conf.Ipaddr6)
		if err != nil {
			wwlog.Error("Invalid ipv6 address specified, mut be CIDR notation: %s", conf.Ipaddr6)
			return errors.New("invalid ipv6 network")
		}
		if msize, _ := ipv6net.Mask.Size(); msize > 64 {
			wwlog.Error("ipv6 mask size must be smaller than 64")
			return errors.New("invalid ipv6 network size")
		}
	}
	cachedConf = *conf
	cachedConf.current = true
	return
}
