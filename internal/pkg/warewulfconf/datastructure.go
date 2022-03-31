package warewulfconf

import (
	"github.com/creasty/defaults"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

type ControllerConf struct {
	Comment  string        `yaml:"comment,omitempty"`
	Ipaddr   string        `yaml:"ipaddr"`
	Ipaddr6  string        `yaml:"ipaddr6,omitempty"`
	Netmask  string        `yaml:"netmask"`
	Network  string        `yaml:"network,omitempty"`
	Ipv6net  string        `yaml:"ipv6net,omitempty"`
	Fqdn     string        `yaml:"fqdn,omitempty"`
	Warewulf *WarewulfConf `yaml:"warewulf"`
	Dhcp     *DhcpConf     `yaml:"dhcp"`
	Tftp     *TftpConf     `yaml:"tftp"`
	Nfs      *NfsConf      `yaml:"nfs"`
	current  bool
}

type WarewulfConf struct {
	Port              int    `yaml:"port"`
	Secure            bool   `yaml:"secure"`
	UpdateInterval    int    `yaml:"update interval"`
	AutobuildOverlays bool   `yaml:"autobuild overlays"`
	EnableHostOverlay bool   `yaml:"host overlay"`
	Syslog            bool   `yaml:"syslog"`
	DataStore         string `yaml:"datastore"`
}

type DhcpConf struct {
	Enabled     bool   `yaml:"enabled"`
	Template    string `yaml:"template"`
	RangeStart  string `yaml:"range start"`
	RangeEnd    string `yaml:"range end"`
	SystemdName string `yaml:"systemd name"`
}

type TftpConf struct {
	Enabled     bool   `yaml:"enabled"`
	TftpRoot    string `yaml:"tftproot"`
	SystemdName string `yaml:"systemd name"`
}

type NfsConf struct {
	Enabled         bool             `default:"true" yaml:"enabled"`
	ExportsExtended []*NfsExportConf `yaml:"export paths"`
	SystemdName     string           `yaml:"systemd name"`
}

type NfsExportConf struct {
	Path          string `yaml:"path"`
	ExportOptions string `default:"rw,sync,no_subtree_check" yaml:"export options"`
	MountOptions  string `default:"defaults" yaml:"mount options"`
	Mount         bool   `default:"true" yaml:"mount"`
}

func (s *NfsConf) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(s); err != nil {
		return err
	}

	type plain NfsConf
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}

	return nil
}

func init() {
	if !util.IsFile(ConfigFile) {
		wwlog.Printf(wwlog.ERROR, "Configuration file not found: %s\n", ConfigFile)
		// fail silently as this also called by bash_completion
	}
	_, err := New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not read Warewulf configuration file: %s\n", err)
	}
}

// Waste processor cycles to make code more readable

func DataStore() string {
	return cachedConf.Warewulf.DataStore
}
