package node

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"os"
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
	Disabled       bool                `yaml:"disabled,omitempty"`
	ClusterName    string              `yaml:"cluster name,omitempty"`
	ContainerName  string              `yaml:"container name,omitempty"`
	Ipxe           string              `yaml:"ipxe template,omitempty"`
	KernelVersion  string              `yaml:"kernel version,omitempty"`
	KernelArgs     string              `yaml:"kernel args,omitempty"`
	IpmiUserName   string              `yaml:"ipmi username,omitempty"`
	IpmiPassword   string              `yaml:"ipmi password,omitempty"`
	IpmiIpaddr     string              `yaml:"ipmi ipaddr,omitempty"`
	IpmiNetmask    string              `yaml:"ipmi netmask,omitempty"`
	RuntimeOverlay string              `yaml:"runtime overlay files,omitempty"`
	SystemOverlay  string              `yaml:"system overlay files,omitempty"`
	Init           string              `yaml:"init,omitempty"`
	Discoverable   bool                `yaml:"discoverable,omitempty"`
	Profiles       []string            `yaml:"profiles,omitempty"`
	NetDevs        map[string]*NetDevs `yaml:"network devices,omitempty"`
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
	IpmiUserName   Entry
	IpmiPassword   Entry
	RuntimeOverlay Entry
	SystemOverlay  Entry
	Discoverable   Entry
	Disabled       Entry
	Init           Entry //TODO: Finish adding this...
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

func init() {
	//TODO: Check to make sure nodes.conf is found
	if util.IsFile(ConfigFile) == false {
		c, err := os.OpenFile(ConfigFile, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not create new configuration file: %s\n", err)
			os.Exit(1)
		}

		fmt.Fprintf(c, "nodeprofiles:\n")
		fmt.Fprintf(c, "  default:\n")
		fmt.Fprintf(c, "    comment: This profile is automatically included for each node\n")
		fmt.Fprintf(c, "    kernel args: crashkernel=no quiet\n")
		fmt.Fprintf(c, "nodes: {}\n")

		c.Close()

		wwlog.Printf(wwlog.INFO, "Created default node configuration\n")
	}
}
