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
	NodeProfiles map[string]*ProfileConf
	Controllers  map[string]*ControllerConf `yaml:"control"`
}

type ProfileConf struct {
	Comment        string `yaml:"comment"`
	Vnfs           string `yaml:"vnfs,omitempty"`
	Ipxe           string `yaml:"ipxe template,omitempty"`
	KernelVersion  string `yaml:"kernel version,omitempty"`
	KernelArgs     string `yaml:"kernel args,omitempty"`
	IpmiNetmask    string `yaml:"ipmi netmask,omitempty"`
	IpmiUserName   string `yaml:"ipmi username,omitempty"`
	IpmiPassword   string `yaml:"ipmi password,omitempty"`
	DomainName     string `yaml:"domain name,omitempty"`
	RuntimeOverlay string `yaml:"runtime overlay files,omitempty"`
	SystemOverlay  string `yaml:"system overlay files,omitempty"`
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
	NodeGroups map[string]*GroupConf
}

type GroupConf struct {
	Comment        string   `yaml:"comment"`
	Disabled       bool     `yaml:"disabled,omitempty"`
	DomainName     string   `yaml:"domain name"`
	Vnfs           string   `yaml:"vnfs,omitempty"`
	Ipxe           string   `yaml:"ipxe template,omitempty"`
	KernelVersion  string   `yaml:"kernel version,omitempty"`
	KernelArgs     string   `yaml:"kernel args,omitempty"`
	IpmiNetmask    string   `yaml:"ipmi netmask,omitempty"`
	IpmiUserName   string   `yaml:"ipmi username,omitempty"`
	IpmiPassword   string   `yaml:"ipmi password,omitempty"`
	RuntimeOverlay string   `yaml:"runtime overlay files,omitempty"`
	SystemOverlay  string   `yaml:"system overlay files,omitempty"`
	Profiles       []string `yaml:"profiles,omitempty"`
	Nodes          map[string]*NodeConf
}

type NodeConf struct {
	Comment        string   `yaml:"comment,omitempty"`
	Disabled       bool     `yaml:"disabled,omitempty"`
	Hostname       string   `yaml:"hostname,omitempty"`
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

type NetDevs struct {
	Type    string `yaml:"type,omitempty"`
	Default bool   `yaml:"default"`
	Hwaddr  string
	Ipaddr  string
	Netmask string
	Gateway string `yaml:"gateway,omitempty"`
}

/******
 * Code internal data representations
 ******/

type Entry struct {
	Node       string
	Profile    string
	Group      string
	Controller string
	Default    string
}

type NodeInfo struct {
	Id             Entry
	Gid            Entry
	Cid            Entry
	Comment        Entry
	HostName       Entry
	Fqdn           Entry
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
	NetDevs        map[string]*NetDevs
}

type ControllerInfo struct {
	Id         string
	Comment    string
	Ipaddr     string
	Fqdn       string
	DomainName string
	Services   struct {
		Warewulfd struct {
			Port       string
			Secure     bool
			StartCmd   string
			RestartCmd string
			EnableCmd  string
		}
		Dhcp struct {
			Enabled    bool
			Template   string
			RangeStart string
			RangeEnd   string
			ConfigFile string
			StartCmd   string
			RestartCmd string
			EnableCmd  string
		}
		Tftp struct {
			Enabled    bool
			TftpRoot   string
			StartCmd   string
			RestartCmd string
			EnableCmd  string
		}
		Nfs struct {
			Enabled    bool
			Exports    []string
			StartCmd   string
			RestartCmd string
			EnableCmd  string
		}
	}
}

type GroupInfo struct {
	Id             Entry
	Cid            Entry
	Comment        Entry
	Vnfs           Entry
	Ipxe           Entry
	KernelVersion  Entry
	KernelArgs     Entry
	IpmiNetmask    Entry
	IpmiUserName   Entry
	IpmiPassword   Entry
	DomainName     Entry
	RuntimeOverlay Entry
	SystemOverlay  Entry
	Profiles       []string
}

type ProfileInfo struct {
	Id             string
	Comment        string
	Vnfs           string
	Ipxe           string
	KernelVersion  string
	KernelArgs     string
	IpmiNetmask    string
	IpmiUserName   string
	IpmiPassword   string
	DomainName     string
	RuntimeOverlay string
	SystemOverlay  string
}

const ConfigFile = "/etc/warewulf/nodes.conf"

func init() {
	//TODO: Check to make sure nodes.conf is found
	if util.IsFile(ConfigFile) == false {
		wwlog.Printf(wwlog.ERROR, "Configuration file not found: %s\n", ConfigFile)
		os.Exit(1)
	}
}
