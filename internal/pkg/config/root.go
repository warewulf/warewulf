// Package config reads, parses, and represents the warewulf.conf
// config file.
//
// warewulf.conf is a yaml-formatted configuration file that includes
// configuration for the Warewulf daemon and commands, as well as the
// DHCP, TFTP and NFS services that Warewulf manages.
package config


import (
	"fmt"
	"net"
	"os"
	"reflect"

	"github.com/pkg/errors"

	"github.com/creasty/defaults"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v2"
)


var cachedConf RootConf


// RootConf is the main Warewulf configuration structure. It stores
// some information about the Warewulf server locally, and has
// [WarewulfConf], [DHCPConf], [TFTPConf], and [NFSConf] sub-sections.
type RootConf struct {
	WWInternal      int           `yaml:"WW_INTERNAL"`
	Comment         string        `yaml:"comment,omitempty"`
	Ipaddr          string        `yaml:"ipaddr"`
	Ipaddr6         string        `yaml:"ipaddr6,omitempty"`
	Netmask         string        `yaml:"netmask"`
	Network         string        `yaml:"network,omitempty"`
	Ipv6net         string        `yaml:"ipv6net,omitempty"`
	Fqdn            string        `yaml:"fqdn,omitempty"`
	Warewulf        *WarewulfConf `yaml:"warewulf"`
	DHCP            *DHCPConf     `yaml:"dhcp"`
	TFTP            *TFTPConf     `yaml:"tftp"`
	NFS             *NFSConf      `yaml:"nfs"`
	MountsContainer []*MountEntry `yaml:"container mounts" default:"[{\"source\": \"/etc/resolv.conf\", \"dest\": \"/etc/resolv.conf\"}]"`
	Paths           *BuildConfig  `yaml:"paths"`

	fromFile        bool
}


// New caches and returns a new [RootConf] initialized with empty
// values, clearing replacing any previously cached value.
func New() (*RootConf) {
	cachedConf = RootConf{}
	cachedConf.fromFile = false
	cachedConf.Warewulf = new(WarewulfConf)
	cachedConf.DHCP = new(DHCPConf)
	cachedConf.TFTP = new(TFTPConf)
	cachedConf.NFS = new(NFSConf)
	cachedConf.Paths = new(BuildConfig)
	if err := defaults.Set(&cachedConf); err != nil {
		panic(err)
	}
	return &cachedConf
}


// Get returns a previously cached [RootConf] if it exists, or returns
// a new RootConf.
func Get() (*RootConf) {
	// NOTE: This function can be called before any log level is set
	//       so using wwlog.Verbose or wwlog.Debug won't work
	if reflect.ValueOf(cachedConf).IsZero() {
		cachedConf = *New()
	}
	return &cachedConf
}


// Read populates [RootConf] with the values from a configuration
// file.
func (conf *RootConf) Read(confFileName string) (error) {
	wwlog.Debug("Reading warewulf.conf from: %s", confFileName)
	if data, err := os.ReadFile(confFileName); err != nil {
		return err
	} else if err := conf.Parse(data); err != nil {
		return err
	} else {
		conf.fromFile = true
		return nil
	}
}


// Parse populates [RootConf] with the values from a yaml document.
func (conf *RootConf) Parse(data []byte) (error) {
	// ipxe binaries are merged not overwritten, store defaults separate
	defIpxe := make(map[string]string)
	for k, v := range conf.TFTP.IpxeBinaries {
		defIpxe[k] = v
		delete(conf.TFTP.IpxeBinaries, k)
	}
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return err
	}
	if len(conf.TFTP.IpxeBinaries) == 0 {
		conf.TFTP.IpxeBinaries = defIpxe
	}
	return nil
}


// SetDynamicDefaults populates [RootConf] with plausible defaults for
// the runtime environment.
func (conf *RootConf) SetDynamicDefaults() (err error) {
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
	if conf.DHCP.RangeStart == "" && conf.DHCP.RangeEnd == "" {
		start := net.ParseIP(conf.Network).To4()
		start[3] += 1
		if start.Equal(net.ParseIP(conf.Ipaddr)) {
			start[3] += 1
		}
		conf.DHCP.RangeStart = start.String()
		wwlog.Verbose("dhpd start is not configured in warewulf.conf, using %s", conf.DHCP.RangeStart)
		sz, _ := net.IPMask(net.ParseIP(conf.Netmask).To4()).Size()
		range_end := (1 << (32 - sz)) / 8
		if range_end > 127 {
			range_end = 127
		}
		end := net.ParseIP(conf.Network).To4()
		end[3] += byte(range_end)
		conf.DHCP.RangeEnd = end.String()
		wwlog.Verbose("dhpd end is not configured in warewulf.conf, using %s", conf.DHCP.RangeEnd)

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
	return
}


// InitializedFromFile returns true if [RootConf] memory was read from
// a file, or false otherwise.
func (conf *RootConf) InitializedFromFile() bool {
	return conf.fromFile
}
