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
	KernelOverride string              `yaml:"kernel override,omitempty"`
	KernelArgs     string              `yaml:"kernel args,omitempty"`
	IpmiUserName   string              `yaml:"ipmi username,omitempty"`
	IpmiPassword   string              `yaml:"ipmi password,omitempty"`
	IpmiIpaddr     string              `yaml:"ipmi ipaddr,omitempty"`
	IpmiNetmask    string              `yaml:"ipmi netmask,omitempty"`
	IpmiPort       string              `yaml:"ipmi port,omitempty"`
	IpmiGateway    string              `yaml:"ipmi gateway,omitempty"`
	IpmiInterface  string              `yaml:"ipmi interface,omitempty"`
	IpmiWrite      bool                `yaml:"ipmi write,omitempty"`
	RuntimeOverlay string              `yaml:"runtime overlay,omitempty"`
	SystemOverlay  string              `yaml:"system overlay,omitempty"`
	Init           string              `yaml:"init,omitempty"`
	Root           string              `yaml:"root,omitempty"`
	AssetKey       string              `yaml:"asset key,omitempty"`
	Discoverable   string              `yaml:"discoverable,omitempty"`
	Profiles       []string            `yaml:"profiles,omitempty"`
	NetDevs        map[string]*NetDevs `yaml:"network devices,omitempty"`
	Tags           map[string]string   `yaml:"tags,omitempty"`
	Keys           map[string]string   `yaml:"keys,omitempty"` // Reverse compatibility
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

type Entry struct {
	value    string
	altvalue string
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
	KernelOverride Entry
	KernelArgs     Entry
	IpmiIpaddr     Entry
	IpmiNetmask    Entry
	IpmiPort       Entry
	IpmiGateway    Entry
	IpmiUserName   Entry
	IpmiPassword   Entry
	IpmiInterface  Entry
	IpmiWrite      Entry
	RuntimeOverlay Entry
	SystemOverlay  Entry
	Root           Entry
	Discoverable   Entry
	Init           Entry //TODO: Finish adding this...
	AssetKey       Entry
	Profiles       []string
	GroupProfiles  []string
	NetDevs        map[string]*NetDevEntry
	Tags           map[string]*Entry
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
