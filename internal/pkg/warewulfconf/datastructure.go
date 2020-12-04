package warewulfconf

import (
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"os"
)

const ConfigFile = "/etc/warewulf/warewulf.conf"

type ControllerConf struct {
	Comment  string        `yaml:"comment"`
	Ipaddr   string        `yaml:"ipaddr"`
	Netmask  string        `yaml:"netmask,omitempty"`
	Fqdn     string        `yaml:"fqdn,omitempty"`
	Warewulf *WarewulfConf `yaml:"warewulf"`
	Dhcp     *DhcpConf     `yaml:"dhcp"`
	Tftp     *TftpConf     `yaml:"tftp"`
	Nfs      *NfsConf      `yaml:"nfs"`
}

type WarewulfConf struct {
	Port    int    `yaml:"port,omitempty"`
	Secure  bool   `yaml:"secure,omitempty"`
	Enable  string `yaml:"enable command,omitempty"`
	Restart string `yaml:"restart command,omitempty"`
}

type DhcpConf struct {
	Enabled    bool   `yaml:"enabled"`
	Template   string `yaml:"template,omitempty"`
	RangeStart string `yaml:"range start,omitempty"`
	RangeEnd   string `yaml:"range end,omitempty"`
	ConfigFile string `yaml:"config file,omitempty"`
	Enable     string `yaml:"enable command,omitempty"`
	Restart    string `yaml:"restart command,omitempty"`
}

type TftpConf struct {
	Enabled bool   `yaml:"enabled"`
	Root    string `yaml:"root,omitempty"`
	Enable  string `yaml:"enable command,omitempty"`
	Restart string `yaml:"restart command,omitempty"`
}

type NfsConf struct {
	Enabled bool     `yaml:"enabled"`
	Exports []string `yaml:"exports,omitempty"`
	Enable  string   `yaml:"enable command,omitempty"`
	Restart string   `yaml:"restart command,omitempty"`
}

func init() {
	//TODO: Check to make sure nodes.conf is found
	if util.IsFile(ConfigFile) == false {
		wwlog.Printf(wwlog.ERROR, "Configuration file not found: %s\n", ConfigFile)
		os.Exit(1)
	}
}
