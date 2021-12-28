package node

import (
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

/******
 * YAML data representations
 ******/

type nodeYaml struct {
	NodeProfiles map[string]*NodeConf
	Nodes        map[string]*NodeConf
}

type NodeConf struct {
	Comment        string              `yaml:"comment,omitempty"`
	ClusterName    string              `yaml:"cluster name,omitempty"`
	ContainerName  string              `yaml:"container name,omitempty"`
	Ipxe           string              `yaml:"ipxe template,omitempty"`
	KernelVersion  string              `yaml:"kernel version,omitempty"`
	KernelArgs     string              `yaml:"kernel args,omitempty"`
	IpmiUserName   string              `yaml:"ipmi username,omitempty"`
	IpmiPassword   string              `yaml:"ipmi password,omitempty"`
	IpmiIpaddr     string              `yaml:"ipmi ipaddr,omitempty"`
	IpmiNetmask    string              `yaml:"ipmi netmask,omitempty"`
	IpmiPort       string              `yaml:"ipmi port,omitempty"`
	IpmiGateway    string              `yaml:"ipmi gateway,omitempty"`
	IpmiInterface  string              `yaml:"ipmi interface,omitempty"`
	RuntimeOverlay string              `yaml:"runtime overlay,omitempty"`
	SystemOverlay  string              `yaml:"system overlay,omitempty"`
	Init           string              `yaml:"init,omitempty"`
	Root           string              `yaml:"root,omitempty"`
	Discoverable   bool                `yaml:"discoverable,omitempty"`
	Profiles       []string            `yaml:"profiles,omitempty"`
	NetDevs        map[string]*NetDevs `yaml:"network devices,omitempty"`
	Keys           map[string]string   `yaml:"keys,omitempty"`
}

type NetDevs struct {
	Name    string
	Type    string `yaml:"type,omitempty"`
	OnBoot  bool   `yaml:"onboot"`
	Device  string `yaml:"device"`
	Hwaddr  string
	Ipaddr  string
	IpCIDR  string
	Prefix  string
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
	def      string
}

type NodeInfo struct {
	Id             Entry
	Cid            Entry
	Comment        Entry
	ClusterName    Entry
	ContainerName  Entry
	Ipxe           Entry
	KernelVersion  Entry
	KernelArgs     Entry
	IpmiIpaddr     Entry
	IpmiNetmask    Entry
	IpmiPort       Entry
	IpmiGateway    Entry
	IpmiUserName   Entry
	IpmiPassword   Entry
	IpmiInterface  Entry
	RuntimeOverlay Entry
	SystemOverlay  Entry
	Root           Entry
	Discoverable   Entry
	Init           Entry //TODO: Finish adding this...
	Profiles       []string
	GroupProfiles  []string
	NetDevs        map[string]*NetDevEntry
	Keys           map[string]*Entry
}

type NetDevEntry struct {
	Name    Entry
	Type    Entry `yaml:"type,omitempty"`
	OnBoot  Entry `yaml:"onboot"`
	Device  Entry `yaml:"device"`
	Hwaddr  Entry
	Ipaddr  Entry
	IpCIDR  Entry
	Prefix  Entry
	Netmask Entry
	Gateway Entry `yaml:"gateway,omitempty"`
}

func init() {
	// Check that nodes.conf is found
	if !util.IsFile(ConfigFile) {
		wwlog.Printf(wwlog.WARN, "Missing node configuration file\n")
		// just return silently, as init is also called for bash_completion
		return
	}
}
