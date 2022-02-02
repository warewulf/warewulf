package warewulfconf

import (
	"path"

	"github.com/creasty/defaults"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

type ControllerConf struct {
	Comment  string        `yaml:"comment,omitempty"`
	Ipaddr   string        `yaml:"ipaddr"`
	Netmask  string        `yaml:"netmask"`
	Network  string        `yaml:"network,omitempty"`
	Fqdn     string        `yaml:"fqdn,omitempty"`
	Chroots  string        `yaml:"chroots,omitempty"`
	Overlays string        `yaml:"overlays,omitempty"`
	Warewulf *WarewulfConf `yaml:"warewulf"`
	Dhcp     *DhcpConf     `yaml:"dhcp"`
	Tftp     *TftpConf     `yaml:"tftp"`
	Nfs      *NfsConf      `yaml:"nfs"`

	current bool
}

type WarewulfConf struct {
	Port              int    `yaml:"port"`
	Secure            bool   `yaml:"secure"`
	UpdateInterval    int    `yaml:"update interval"`
	AutobuildOverlays bool   `yaml:"autobuild overlays"`
	Syslog            bool   `yaml:"syslog"`
	DataStore         string `yaml:"datastore,omitempty"`
}

type DhcpConf struct {
	Enabled     bool   `yaml:"enabled"`
	Template    string `yaml:"template"`
	RangeStart  string `yaml:"range start"`
	RangeEnd    string `yaml:"range end"`
	SystemdName string `yaml:"systemd name"`
	ConfigFile  string `yaml:"config file,omitempty"`
}

type TftpConf struct {
	Enabled     bool   `yaml:"enabled"`
	TftpRoot    string `yaml:"tftproot"`
	SystemdName string `yaml:"systemd name"`
}

type NfsConf struct {
	Enabled         bool             `default:"true" yaml:"enabled"`
	Exports         []string         `yaml:"exports"`
	ExportsExtended []*NfsExportConf `yaml:"export paths"`
	SystemdName     string           `yaml:"systemd name"`
}

type NfsExportConf struct {
	Path          string `yaml:"path"`
	ExportOptions string `yaml:"export options"`
	MountOptions  string `yaml:"mount options"`
	Mount         bool   `yaml:"mount"`
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

// Global application and server configuration GET function
func Config(key string) string {
	// Simplify access to other configuration settings here.
	switch key {
	case "datastore":
		value := path.Join(cachedConf.Warewulf.DataStore, buildData["WAREWULF"])
		wwlog.Printf(wwlog.DEBUG, "%s = '%s'\n", key, value)
		return value
	case "tftproot":
		value := path.Join(cachedConf.Tftp.TftpRoot, buildData["WAREWULF"])
		wwlog.Printf(wwlog.DEBUG, "%s = '%s'\n", key, value)
		return value
	}
	// Return data from the build configuration map
	if value, exists := buildData[key]; exists {
		wwlog.Printf(wwlog.DEBUG, "%s = '%s'\n", key, value)
		return value
	} else {
		wwlog.Printf(wwlog.ERROR, "%s is undefined\n", key)
		return "UNDEF"
	}
}
