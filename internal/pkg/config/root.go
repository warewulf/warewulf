// Package config reads, parses, and represents the warewulf.conf
// config file.
//
// warewulf.conf is a yaml-formatted configuration file that includes
// configuration for the Warewulf daemon and commands, as well as the
// DHCP, TFTP and NFS services that Warewulf manages.
package config

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"reflect"

	"github.com/creasty/defaults"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v3"
)

var cachedConf WarewulfYaml

// WarewulfYaml is the main Warewulf configuration structure. It stores
// some information about the Warewulf server locally, and has
// [WarewulfConf], [DHCPConf], [TFTPConf], and [NFSConf] sub-sections.
type WarewulfYaml struct {
	Comment     string        `yaml:"comment,omitempty"`
	Ipaddr      string        `yaml:"ipaddr,omitempty"`
	Ipaddr6     string        `yaml:"ipaddr6,omitempty"`
	Netmask     string        `yaml:"netmask,omitempty"`
	Network     string        `yaml:"network,omitempty"`
	Ipv6net     string        `yaml:"ipv6net,omitempty"`
	Fqdn        string        `yaml:"fqdn,omitempty"`
	Warewulf    *WarewulfConf `yaml:"warewulf,omitempty"`
	DHCP        *DHCPConf     `yaml:"dhcp,omitempty"`
	TFTP        *TFTPConf     `yaml:"tftp,omitempty"`
	NFS         *NFSConf      `yaml:"nfs,omitempty"`
	SSH         *SSHConf      `yaml:"ssh,omitempty"`
	MountsImage []*MountEntry `yaml:"image mounts,omitempty" default:"[{\"source\": \"/etc/resolv.conf\", \"dest\": \"/etc/resolv.conf\"}]"`
	Paths       *BuildConfig  `yaml:"paths,omitempty"`
	WWClient    *WWClientConf `yaml:"wwclient,omitempty"`

	warewulfconf string
	autodetected bool
}

// New caches and returns a new [WarewulfYaml] initialized with empty
// values, clearing replacing any previously cached value.
func New() *WarewulfYaml {
	cachedConf = WarewulfYaml{}
	cachedConf.warewulfconf = ""
	cachedConf.Warewulf = new(WarewulfConf)
	cachedConf.DHCP = new(DHCPConf)
	cachedConf.TFTP = new(TFTPConf)
	cachedConf.NFS = new(NFSConf)
	cachedConf.SSH = new(SSHConf)
	cachedConf.Paths = new(BuildConfig)
	if err := defaults.Set(&cachedConf); err != nil {
		panic(err)
	}
	return &cachedConf
}

// Get returns a previously cached [WarewulfYaml] if it exists, or returns
// a new WarewulfYaml.
func Get() *WarewulfYaml {
	// NOTE: This function can be called before any log level is set
	//       so using wwlog.Verbose or wwlog.Debug won't work
	if reflect.ValueOf(cachedConf).IsZero() {
		cachedConf = *New()
	}
	return &cachedConf
}

// Read populates [WarewulfYaml] with the values from a configuration
// file.
func (conf *WarewulfYaml) Read(confFileName string, autodetect bool) error {
	wwlog.Debug("Reading warewulf.conf from: %s", confFileName)
	conf.warewulfconf = confFileName
	if data, err := os.ReadFile(confFileName); err != nil {
		return err
	} else if err := conf.Parse(data, autodetect); err != nil {
		return err
	} else {
		return nil
	}
}

// Parse populates [WarewulfYaml] with the values from a yaml document.
func (conf *WarewulfYaml) Parse(data []byte, autodetect bool) error {
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

	if ip, network, err := net.ParseCIDR(conf.Ipaddr); err == nil {
		conf.Ipaddr = ip.String()
		if conf.Network == "" {
			conf.Network = network.IP.String()
		}
		if conf.Netmask == "" {
			conf.Netmask = net.IP(network.Mask).String()
		}
	}

	if autodetect {
		if conf.Ipaddr == "" {
			if ip := GetOutboundIP(); ip != nil {
				conf.Ipaddr = ip.String()
				conf.autodetected = true
			}
		}

		if conf.Netmask == "" {
			if ip := net.ParseIP(conf.Ipaddr); ip != nil {
				if network, err := GetIPNetForIP(ip); err == nil {
					conf.Netmask = net.IP(network.Mask).String()
					conf.autodetected = true
				}
			}
		}

		if conf.Network == "" {
			if ip := net.ParseIP(conf.Ipaddr); ip != nil {
				if mask := net.IPMask(net.ParseIP(conf.Netmask)); mask != nil {
					conf.Network = ip.Mask(mask).String()
					conf.autodetected = true
				}
			}
		}
	}
	if conf.Ipaddr6 != "" {
		if conf.Ipv6net == "" {
			wwlog.Error("Ipv6 network has not been set in warewulf.conf: ipv6net")
			return fmt.Errorf("Invalid Ipv6 network")
		}
		_, ipv6net, err := net.ParseCIDR(conf.Ipv6net)
		if err != nil {
			wwlog.Error("Invalid Ipv6 address specified, must be CIDR notation: %s", conf.Ipv6net)
			return fmt.Errorf("Invalid Ipv6 network")
		}
		if msize, _ := ipv6net.Mask.Size(); msize > 64 {
			wwlog.Error("ipv6 mask size must be smaller than 64")
			return fmt.Errorf("Invalid Ipv6 network size")
		}
	}

	return nil
}

func (config *WarewulfYaml) NetworkCIDR() string {
	if config.Network == "" || config.Netmask == "" {
		return ""
	}
	cidr := net.IPNet{
		IP:   net.ParseIP(config.Network),
		Mask: net.IPMask(net.ParseIP(config.Netmask)),
	}
	if cidr.IP == nil || cidr.Mask == nil {
		return ""
	}
	return cidr.String()
}

func (config *WarewulfYaml) IpCIDR() string {
	if config.Ipaddr == "" || config.Netmask == "" {
		return ""
	}
	cidr := net.IPNet{
		IP:   net.ParseIP(config.Ipaddr),
		Mask: net.IPMask(net.ParseIP(config.Netmask)),
	}
	if cidr.IP == nil || cidr.Mask == nil {
		return ""
	}
	return cidr.String()
}

// InitializedFromFile returns true if [WarewulfYaml] memory was read from
// a file, or false otherwise.
func (conf *WarewulfYaml) InitializedFromFile() bool {
	return conf.warewulfconf != ""
}

func (conf *WarewulfYaml) GetWarewulfConf() string {
	return conf.warewulfconf
}

func (conf *WarewulfYaml) Autodetected() bool {
	return conf.autodetected
}

func (config *WarewulfYaml) Dump() ([]byte, error) {
	var buf bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&buf)
	yamlEncoder.SetIndent(2)
	err := yamlEncoder.Encode(config)
	return buf.Bytes(), err
}

func (config *WarewulfYaml) PersistToFile(configFile string) error {
	out, dumpErr := config.Dump()
	if dumpErr != nil {
		wwlog.Error("%s", dumpErr)
		return dumpErr
	}
	file, err := os.OpenFile(configFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		wwlog.Error("%s", err)
		return err
	}
	defer file.Close()
	_, err = file.WriteString(string(out))
	if err != nil {
		return err
	}
	wwlog.Debug("persisted: %s", configFile)
	return nil
}
