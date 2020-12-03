package node

import (
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"os"
)

/******
 * YAML data representations
 ******/

type nodeYaml struct {
	Controllers  map[string]*ControllerConf `yaml:"controlers"`
	NodeProfiles map[string]*NodeConf
	Nodes        map[string]*NodeConf
}

type NodeConf struct {
	Comment  string `yaml:"comment,omitempty"`
	Disabled bool   `yaml:"disabled,omitempty"`
	//	Hostname       string   `yaml:"hostname,omitempty"`
	DomainName     string   `yaml:"domain name,omitempty"`
	Vnfs           string   `yaml:"vnfs,omitempty"`
	Ipxe           string   `yaml:"ipxe template,omitempty"`
	KernelVersion  string   `yaml:"kernel version,omitempty"`
	KernelArgs     string   `yaml:"kernel args,omitempty"`
	IpmiUserName   string   `yaml:"ipmi username,omitempty"`
	IpmiPassword   string   `yaml:"ipmi password,omitempty"`
	IpmiIpaddr     string   `yaml:"ipmi ipaddr,omitempty"`
	IpmiNetmask    string   `yaml:"ipmi netmask,omitempty"`
	RuntimeOverlay string   `yaml:"runtime overlay files,omitempty"`
	SystemOverlay  string   `yaml:"system overlay files,omitempty"`
	Profiles       []string `yaml:"profiles,omitempty"`
	NetDevs        map[string]*NetDevs
}

type ControllerConf struct {
	Comment  string `yaml:"comment"`
	Ipaddr   string `yaml:"ipaddr"`
	Fqdn     string `yaml:"fqdn"`
	Services struct {
		Warewulfd struct {
			Port       string `yaml:"port"`
			Secure     bool   `yaml:"secure,omitempty"`
			StartCmd   string `yaml:"start command,omitempty"`
			RestartCmd string `yaml:"restart command,omitempty"`
			EnableCmd  string `yaml:"enable command,omitempty"`
		} `yaml:"warewulf"`
		Dhcp struct {
			Enabled    bool   `yaml:"enabled,omitempty"`
			Template   string `yaml:"template,omitempty"`
			RangeStart string `yaml:"range start,omitempty"`
			RangeEnd   string `yaml:"range end,omitempty"`
			ConfigFile string `yaml:"config file,omitempty"`
			StartCmd   string `yaml:"start command,omitempty"`
			RestartCmd string `yaml:"restart command,omitempty"`
			EnableCmd  string `yaml:"enable command,omitempty"`
		} `yaml:"dhcp,omitempty"`
		Tftp struct {
			Enabled    bool   `yaml:"enabled,omitempty"`
			TftpRoot   string `yaml:"tftp root,omitempty"`
			StartCmd   string `yaml:"start command,omitempty"`
			RestartCmd string `yaml:"restart command,omitempty"`
			EnableCmd  string `yaml:"enable command,omitempty"`
		} `yaml:"tftp,omitempty"`
		Nfs struct {
			Enabled    bool     `yaml:"enabled,omitempty"`
			Exports    []string `yaml:"exports,omitempty"`
			StartCmd   string   `yaml:"start command,omitempty"`
			RestartCmd string   `yaml:"restart command,omitempty"`
			EnableCmd  string   `yaml:"enable command,omitempty"`
		} `yaml:"nfs,omitempty"`
	} `yaml:"services"`
}

type NetDevs struct {
	Type    string `yaml:"type,omitempty"`
	Default bool   `yaml:"default"`
	Hwaddr  string
	Ipaddr  string
	Netmask string
	Gateway string `yaml:"gateway,omitempty"`
}

/******
 * Internal code data representations
 ******/

type Entry struct {
	value    string
	altvalue string
	bool     bool
	altbool  bool
	from     string
}

type NodeInfo struct {
	Id      Entry
	Cid     Entry
	Comment Entry
	//	HostName       Entry
	//	Fqdn           Entry
	DomainName     Entry
	Vnfs           Entry
	Ipxe           Entry
	KernelVersion  Entry
	KernelArgs     Entry
	IpmiIpaddr     Entry
	IpmiNetmask    Entry
	IpmiUserName   Entry
	IpmiPassword   Entry
	RuntimeOverlay Entry
	SystemOverlay  Entry
	Profiles       []string
	GroupProfiles  []string
	NetDevs        map[string]*NetDevEntry
}

type NetDevEntry struct {
	Type    Entry `yaml:"type,omitempty"`
	Default Entry `yaml:"default"`
	Hwaddr  Entry
	Ipaddr  Entry
	Netmask Entry
	Gateway Entry `yaml:"gateway,omitempty"`
}

const ConfigFile = "/etc/warewulf/nodes.conf"

func init() {
	//TODO: Check to make sure nodes.conf is found
	if util.IsFile(ConfigFile) == false {
		wwlog.Printf(wwlog.ERROR, "Configuration file not found: %s\n", ConfigFile)
		os.Exit(1)
	}
}
