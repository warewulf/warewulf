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
	IpmiUserName   string `yaml:"ipmi username,omitempty"`
	IpmiPassword   string `yaml:"ipmi password,omitempty"`
	DomainName     string `yaml:"domain name,omitempty"`
	RuntimeOverlay string `yaml:"runtime overlay files,omitempty"`
	SystemOverlay  string `yaml:"system overlay files,omitempty"`
}

type ControllerConf struct {
	Comment  string `yaml:"comment"`
	Ipaddr   string `yaml:"ipaddr"`
	Services struct {
		Warewulfd struct {
			Port       string `yaml:"port"`
			Secure     string `yaml:"secure,omitempty"`
			StartCmd   string `yaml:"start command,omitempty"`
			RestartCmd string `yaml:"restart command,omitempty"`
			EnableCmd  string `yaml:"enable command,omitempty"`
		} `yaml:"warewulf"`
		Dhcp struct {
			Enabled    bool   `yaml:"enabled,omitempty"`
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
	RuntimeOverlay string   `yaml:"runtime overlay files,omitempty"`
	SystemOverlay  string   `yaml:"system overlay files,omitempty"`
	Profiles       []string `yaml:"profiles,omitempty"`
	NetDevs        map[string]*NetDevs
}

type NetDevs struct {
	Type    string `yaml:"type,omitempty"`
	Hwaddr  string
	Ipaddr  string
	Netmask string
	Gateway string `yaml:"gateway,omitempty"`
}

/******
 * Code internal data representations
 ******/

type NodeInfoEntry struct {
	value      string
	profile    string
	group      string
	controller string
	def        string
}

type NodeInfo struct {
	Id             NodeInfoEntry
	Gid            NodeInfoEntry
	Cid            NodeInfoEntry
	Comment        NodeInfoEntry
	HostName       NodeInfoEntry
	Fqdn           NodeInfoEntry
	DomainName     NodeInfoEntry
	Vnfs           NodeInfoEntry
	Ipxe           NodeInfoEntry
	KernelVersion  NodeInfoEntry
	KernelArgs     NodeInfoEntry
	IpmiIpaddr     NodeInfoEntry
	IpmiUserName   NodeInfoEntry
	IpmiPassword   NodeInfoEntry
	RuntimeOverlay NodeInfoEntry
	SystemOverlay  NodeInfoEntry
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
			Secure     string
			StartCmd   string
			RestartCmd string
			EnableCmd  string
		}
		Dhcp struct {
			Enabled    bool
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
	Id             string
	Cid            string
	Comment        string
	Vnfs           string
	Ipxe           string
	KernelVersion  string
	KernelArgs     string
	IpmiUserName   string
	IpmiPassword   string
	DomainName     string
	RuntimeOverlay string
	SystemOverlay  string
	Profiles       []string
}

type ProfileInfo struct {
	Id             string
	Comment        string
	Vnfs           string
	Ipxe           string
	KernelVersion  string
	KernelArgs     string
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
