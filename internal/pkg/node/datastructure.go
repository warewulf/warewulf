package node

import (
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

/******
 * YAML data representations
 ******/

type NodeYaml struct {
	WWInternal   int `yaml:"WW_INTERNAL"`
	NodeProfiles map[string]*NodeConf
	Nodes        map[string]*NodeConf
}

/*
NodeConf is the datastructure which is stored on disk.
*/
type NodeConf struct {
	Comment       string `yaml:"comment,omitempty" lopt:"comment" comment:"Set arbitrary string comment"`
	ClusterName   string `yaml:"cluster name,omitempty" lopt:"cluster" sopt:"c" comment:"Set cluster group"`
	ContainerName string `yaml:"container name,omitempty" lopt:"container" sopt:"C" comment:"Set container name"`
	Ipxe          string `yaml:"ipxe template,omitempty" lopt:"ipxe" comment:"Set the iPXE template name"`
	// Deprecated start
	// Kernel settings here are deprecated and here for backward comptability
	KernelVersion  string `yaml:"kernel version,omitempty"`
	KernelOverride string `yaml:"kernel override,omitempty"`
	KernelArgs     string `yaml:"kernel args,omitempty"`
	// Ipmi settings herer are deprecated and here for backward comptability
	IpmiUserName  string `yaml:"ipmi username,omitempty"`
	IpmiPassword  string `yaml:"ipmi password,omitempty"`
	IpmiIpaddr    string `yaml:"ipmi ipaddr,omitempty"`
	IpmiNetmask   string `yaml:"ipmi netmask,omitempty"`
	IpmiPort      string `yaml:"ipmi port,omitempty"`
	IpmiGateway   string `yaml:"ipmi gateway,omitempty"`
	IpmiInterface string `yaml:"ipmi interface,omitempty"`
	IpmiWrite     string `yaml:"ipmi write,omitempty"`
	// Deprecated end
	RuntimeOverlay []string            `yaml:"runtime overlay,omitempty" lopt:"runtime" sopt:"R" comment:"Set the runtime overlay"`
	SystemOverlay  []string            `yaml:"system overlay,omitempty" lopt:"wwinit" sopt:"O" comment:"Set the system overlay"`
	Kernel         *KernelConf         `yaml:"kernel,omitempty"`
	Ipmi           *IpmiConf           `yaml:"ipmi,omitempty"`
	Init           string              `yaml:"init,omitempty" lopt:"init" sopt:"i" comment:"Define the init process to boot the container"`
	Root           string              `yaml:"root,omitempty" lopt:"root" comment:"Define the rootfs" `
	AssetKey       string              `yaml:"asset key,omitempty" lopt:"asset" comment:"Set the node's Asset tag (key)"`
	Discoverable   string              `yaml:"discoverable,omitempty" lopt:"discoverable" comment:"Make discoverable in given network (yes/no)"`
	Profiles       []string            `yaml:"profiles,omitempty" lopt:"profile" sopt:"P" comment:"Set the node's profile members (comma separated)"`
	NetDevs        map[string]*NetDevs `yaml:"network devices,omitempty"`
	Tags           map[string]string   `yaml:"tags,omitempty" lopt:"tagadd" comment:"base key"`
	TagsDel        []string            `yaml:"tagsdel,omitempty" lopt:"tagdel" comment:"remove this tags"` // should not go to disk only to wire
	Keys           map[string]string   `yaml:"keys,omitempty"`                                             // Reverse compatibility
}

type IpmiConf struct {
	UserName  string            `yaml:"username,omitempty" lopt:"ipmiuser" comment:"Set the IPMI username"`
	Password  string            `yaml:"password,omitempty" lopt:"ipmipass" comment:"Set the IPMI password"`
	Ipaddr    string            `yaml:"ipaddr,omitempty" lopt:"ipmiaddr" comment:"Set the IPMI IP address"`
	Netmask   string            `yaml:"netmask,omitempty" lopt:"ipminetmask" comment:"Set the IPMI netmask"`
	Port      string            `yaml:"port,omitempty" lopt:"ipmiport" comment:"Set the IPMI port"`
	Gateway   string            `yaml:"gateway,omitempty" lopt:"ipmigateway" comment:"Set the IPMI gateway"`
	Interface string            `yaml:"interface,omitempty" lopt:"ipmiinterface" comment:"Set the node's IPMI interface (defaults: 'lan')"`
	Write     string            `yaml:"write,omitempty" lopt:"ipmiwrite" comment:"Enable the write of impi configuration (yes/no)"`
	Tags      map[string]string `yaml:"tags,omitempty" lopt:"ipmitagadd" comment:"add ipmitags"`
	TagsDel   []string          `yaml:"tagsdel,omitempty" lopt:"ipmitagdel" comment:"remove ipmitags"` // should not go to disk only to wire
}
type KernelConf struct {
	Version  string `yaml:"version,omitempty"`
	Override string `yaml:"override,omitempty" lopt:"kerneloverride" sopt:"K" comment:"Set kernel override version"`
	Args     string `yaml:"args,omitempty" lopt:"kernelargs" sopt:"A" comment:"Set Kernel argument"`
}

type NetDevs struct {
	Type    string            `yaml:"type,omitempty" lopt:"type" sopt:"T" comment:"Set device type of given network"`
	OnBoot  string            `yaml:"onboot,omitempty" lopt:"onboot" comment:"Enable/disable network device (yes/no)"`
	Device  string            `yaml:"device,omitempty" lopt:"netdev" sopt:"N" comment:"Set the device for given network"`
	Hwaddr  string            `yaml:"hwaddr,omitempty" lopt:"hwaddr" sopt:"H" comment:"Set the device's HW address for given network"`
	Ipaddr  string            `yaml:"ipaddr,omitempty" comment:"IPv4 address in given network" sopt:"I" lopt:"ipaddr"`
	IpCIDR  string            `yaml:"ipcidr,omitempty"`
	Ipaddr6 string            `yaml:"ip6addr,omitempty" lopt:"ipaddr6" comment:"IPv6 address"`
	Prefix  string            `yaml:"prefix,omitempty"`
	Netmask string            `yaml:"netmask,omitempty" lopt:"netmask" sopt:"M" comment:"Set the networks netmask"`
	Gateway string            `yaml:"gateway,omitempty" lopt:"gateway" sopt:"G" comment:"Set the node's network device gateway"`
	Primary string            `yaml:"primary,omitempty" lopt:"primary" comment:"Enable/disable network device as primary (yes/no)"`
	Default string            `yaml:"default,omitempty"` /* backward compatibility */
	Tags    map[string]string `yaml:"tags,omitempty" lopt:"nettagadd" comment:"network tags"`
	TagsDel []string          `yaml:"tagsdel,omitempty" lopt:"nettagdel" comment:"delete network tags"` // should not go to disk only to wire
}

/******
 * Internal code data representations
 ******/
/*
Holds string values, when accessed via Get, its value
is returned which is the default or if set the value
from the profile or if set the value of the node itself
*/
type Entry struct {
	value    []string
	altvalue []string
	from     string
	def      []string
}

/*
NodeInfo is the in memory datastructure, which can containe
a default value, which is overwritten by the overlay from the
overlay (altvalue) which is overwitten by the value of the
node itself, for all values of type Entry.
*/
type NodeInfo struct {
	Id             Entry
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
	Profiles       Entry
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
	Tags      map[string]*Entry
}

type KernelEntry struct {
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
	Primary Entry
	Tags    map[string]*Entry
}

// string which is printed if no value is set
const NoValue = "--"

func init() {
	// Check that nodes.conf is found
	if !util.IsFile(ConfigFile) {
		wwlog.Warn("Missing node configuration file")
		// just return silently, as init is also called for bash_completion
		return
	}
}
