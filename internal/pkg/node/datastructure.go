package node

import (
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

/******
 * YAML data representations
 ******/

type nodeYaml struct {
	WWInternal   int `yaml:"WW_INTERNAL"`
	NodeProfiles map[string]*NodeConf
	Nodes        map[string]*NodeConf
}

type NodeConf struct {
	Comment        string              `yaml:"comment,omitempty"`
	ClusterName    string              `yaml:"cluster name,omitempty"`
	ContainerName  string              `yaml:"container name,omitempty"`
	Ipxe           string              `yaml:"ipxe template,omitempty"`
	RuntimeOverlay []string            `yaml:"runtime overlay,omitempty"`
	SystemOverlay  []string            `yaml:"system overlay,omitempty"`
	Kernel         *KernelConf         `yaml:"kernel,omitempty"`
	Ipmi           *IpmiConf           `yaml:"ipmi,omitempty"`
	Init           string              `yaml:"init,omitempty"`
	Root           string              `yaml:"root,omitempty"`
	AssetKey       string              `yaml:"asset key,omitempty"`
	Discoverable   string              `yaml:"discoverable,omitempty"`
	Profiles       []string            `yaml:"profiles,omitempty"`
	NetDevs        map[string]*NetDevs `yaml:"network devices,omitempty"`
	Tags           map[string]string   `yaml:"tags,omitempty"`
	Keys           map[string]string   `yaml:"keys,omitempty"` // Reverse compatibility
}

type IpmiConf struct {
	UserName  string `yaml:"username,omitempty"`
	Password  string `yaml:"password,omitempty"`
	Ipaddr    string `yaml:"ipaddr,omitempty"`
	Netmask   string `yaml:"netmask,omitempty"`
	Port      string `yaml:"port,omitempty"`
	Gateway   string `yaml:"gateway,omitempty"`
	Interface string `yaml:"interface,omitempty"`
	Write     bool   `yaml:"write,omitempty"`
}
type KernelConf struct {
	Version  string `yaml:"version,omitempty"`
	Override string `yaml:"override,omitempty"`
	Args     string `yaml:"args,omitempty"`
}

type NetDevs struct {
	Type    string `yaml:"type,omitempty"`
	OnBoot  string `yaml:"onboot,omitempty"`
	Device  string `yaml:"device,omitempty"`
	Hwaddr  string `yaml:"hwaddr,omitempty"`
	Ipaddr  string `yaml:"ipaddr,omitempty"`
	IpCIDR  string `yaml:"ipcidr,omitempty"`
	Ipaddr6 string `yaml:"ip6addr,omitempty"`
	Prefix  string `yaml:"prefix,omitempty"`
	Netmask string `yaml:"netmask,omitempty"`
	Gateway string `yaml:"gateway,omitempty"`
	Default string `yaml:"default,omitempty"`
}

/******
 * Internal code data representations
 ******/
/*
Holds a strtng value, when accessed via Get, its value
is returned which is the default or if set the value
from the profile or if set the value of the node itself
*/
type Entry struct {
	value    []string
	altvalue []string
	from     string
	def      []string
}

type NodeInfo struct {
	Id             Entry
	Cid            Entry
	Comment        Entry
	ClusterName    Entry
	ContainerName  Entry
	Ipxe           Entry
	RuntimeOverlay Entry
	SystemOverlay  Entry
	Root           Entry
	Discoverable   Entry
	Init           Entry //TODO: Finish adding this...
	AssetKey       Entry
	Kernel         *KernelEntry
	Ipmi           *IpmiEntry
	Profiles       []string
	GroupProfiles  []string
	NetDevs        map[string]*NetDevEntry
	Tags           map[string]*Entry
}

type IpmiEntry struct {
	Ipaddr    Entry
	Netmask   Entry
	Port      Entry
	Gateway   Entry
	UserName  Entry
	Password  Entry
	Interface Entry
	Write     Entry
}

type KernelEntry struct {
	Version  Entry
	Override Entry
	Args     Entry
}

type NetDevEntry struct {
	Type    Entry
	OnBoot  Entry
	Device  Entry
	Hwaddr  Entry
	Ipaddr  Entry
	Ipaddr6 Entry
	IpCIDR  Entry
	Prefix  Entry
	Netmask Entry
	Gateway Entry
	Default Entry
}

func init() {
	// Check that nodes.conf is found
	if !util.IsFile(ConfigFile) {
		wwlog.Printf(wwlog.WARN, "Missing node configuration file\n")
		// just return silently, as init is also called for bash_completion
		return
	}
}
