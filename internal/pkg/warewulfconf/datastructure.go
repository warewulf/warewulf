package warewulfconf

import (
	"github.com/creasty/defaults"
)

type ControllerConf struct {
	WWInternal      int           `yaml:"WW_INTERNAL"`
	Comment         string        `yaml:"comment,omitempty"`
	Ipaddr          string        `yaml:"ipaddr"`
	Ipaddr6         string        `yaml:"ipaddr6,omitempty"`
	Netmask         string        `yaml:"netmask"`
	Network         string        `yaml:"network,omitempty"`
	Ipv6net         string        `yaml:"ipv6net,omitempty"`
	Fqdn            string        `yaml:"fqdn,omitempty"`
	Warewulf        *WarewulfConf `yaml:"warewulf"`
	Dhcp            *DhcpConf     `yaml:"dhcp"`
	Tftp            *TftpConf     `yaml:"tftp"`
	Nfs             *NfsConf      `yaml:"nfs"`
	MountsContainer []*MountEntry `yaml:"container mounts" default:"[{\"source\": \"/etc/resolv.conf\", \"dest\": \"/etc/resolv.conf\"}]"`
	Paths           *BuildConfig  `yaml:"paths"`
	current         bool
	readConf        bool
}

type WarewulfConf struct {
	Port              int    `yaml:"port" default:"9983"`
	Secure            bool   `yaml:"secure" default:"true"`
	UpdateInterval    int    `yaml:"update interval" default:"60"`
	AutobuildOverlays bool   `yaml:"autobuild overlays" default:"true"`
	EnableHostOverlay bool   `yaml:"host overlay" default:"true"`
	Syslog            bool   `yaml:"syslog" default:"false"`
	DataStore         string `yaml:"datastore" default:"/var/lib/warewulf"`
}

type DhcpConf struct {
	Enabled     bool   `yaml:"enabled" default:"true"`
	Template    string `yaml:"template" default:"default"`
	RangeStart  string `yaml:"range start,omitempty"`
	RangeEnd    string `yaml:"range end,omitempty"`
	SystemdName string `yaml:"systemd name" default:"dhcpd"`
}

type TftpConf struct {
	Enabled     bool   `yaml:"enabled" default:"true"`
	TftpRoot    string `yaml:"tftproot" default:"/var/lib/tftpboot"`
	SystemdName string `yaml:"systemd name" default:"tftp"`
	// Path is relative to buildconfig.DATADIR()
	IpxeBinaries map[string]string `yaml:"ipxe" default:"{\"00:09\": \"x86_64.efi\",\"00:00\": \"x86_64.kpxe\",\"00:0B\": \"arm64.efi\",\"00:07\":  \"x86_64.efi\"}"`
}

type NfsConf struct {
	Enabled         bool             `yaml:"enabled" default:"true"`
	ExportsExtended []*NfsExportConf `yaml:"export paths" default:"[]"`
	SystemdName     string           `yaml:"systemd name" default:"nfsd"`
}

type NfsExportConf struct {
	Path          string `yaml:"path" default:"/dev/null"`
	ExportOptions string `default:"rw,sync,no_subtree_check" yaml:"export options"`
	MountOptions  string `default:"defaults" yaml:"mount options"`
	Mount         bool   `default:"true" yaml:"mount"`
}

/*
Describe a mount point for a container exec
*/
type MountEntry struct {
	Source   string `yaml:"source" default:"/etc/resolv.conf"`
	Dest     string `yaml:"dest,omitempty" default:"/etc/resolv.conf"`
	ReadOnly bool   `yaml:"readonly,omitempty" default:"false"`
	Options  string `yaml:"options,omitempty"` // ignored at the moment
}

func (s *NfsConf) Unmarshal(unmarshal func(interface{}) error) error {
	if err := defaults.Set(s); err != nil {
		return err
	}
	return nil
}

// Waste processor cycles to make code more readable

func DataStore() string {
	_ = New()
	return cachedConf.Warewulf.DataStore
}
