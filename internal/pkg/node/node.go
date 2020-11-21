package node

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/errors"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"regexp"
)

const ConfigFile = "/etc/warewulf/nodes.conf"
const LocalStateDir = "/var/warewulf"

func init() {
	//TODO: Check to make sure nodes.conf is found

}

type nodeYaml struct {
	NodeGroups map[string]*nodeGroup //`yaml:"nodegroups"`
}

type nodeGroup struct {
	Comment        string
	Vnfs           string `yaml:"vnfs"`
	Ipxe           string `yaml:"ipxe template"`
	SystemOverlay  string `yaml:"system overlay""`
	RuntimeOverlay string `yaml:"runtime overlay""`
	DomainSuffix   string `yaml:"domain suffix"`
	KernelVersion  string `yaml:"kernel version"`
	KernelArgs     string `yaml:"kernel args"`
	Nodes          map[string]*nodeEntry
}

type nodeEntry struct {
	Hostname       string
	Vnfs           string `yaml:"vnfs,omitempty"`
	Ipxe           string `yaml:"ipxe template,omitempty"`
	SystemOverlay  string `yaml:"system overlay,omitempty"`
	RuntimeOverlay string `yaml:"runtime overlay,omitempty"`
	DomainSuffix   string `yaml:"domain suffix,omitempty"`
	KernelVersion  string `yaml:"kernel version,omitempty"`
	KernelArgs     string `yaml:"kernel args,omitempty"`
	IpmiIpaddr     string `yaml:"ipmi ipaddr,omitempty"`
	IpmiUserName   string `yaml:"ipmi username,omitempty"`
	IpmiPassword   string `yaml:"ipmi password,omitempty"`
	NetDevs        map[string]netDevs
}

type netDevs struct {
	Type    string
	Hwaddr  string
	Ipaddr  string
	Netmask string
	Gateway string
}

type NodeInfo struct {
	Id             string
	GroupName      string
	HostName       string
	DomainName     string
	Fqdn           string
	Vnfs           string
	VnfsDir        string
	Ipxe           string
	SystemOverlay  string
	RuntimeOverlay string
	KernelVersion  string
	KernelArgs     string
	IpmiIpaddr     string
	IpmiUserName   string
	IpmiPassword   string
	NetDevs        map[string]netDevs
}

func New() (nodeYaml, error) {
	var ret nodeYaml

	wwlog.Printf(wwlog.DEBUG, "Opening node configuration file: %s\n", ConfigFile)
	data, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		fmt.Printf("error reading node configuration file\n")
		return ret, err
	}

	err = yaml.Unmarshal(data, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func (self *nodeYaml) AddGroup(groupname string) error {
	var group nodeGroup

	self.NodeGroups[groupname] = &group
	self.NodeGroups[groupname].DomainSuffix = groupname
	return nil
}

func (self *nodeYaml) AddNode(groupname string, nodename string) error {
	var node nodeEntry

	self.NodeGroups[groupname].Nodes[nodename] = &node
	self.NodeGroups[groupname].Nodes[nodename].Hostname = nodename

	return nil
}

func (self *nodeYaml) SetGroupVal(groupID string, entry string, value string) (int, error) {
	var count int

	for gid, group := range self.NodeGroups {
		for nid := range group.Nodes {
			if groupID == gid {
				if entry == "vnfs" {
					self.NodeGroups[gid].Vnfs = value
					self.NodeGroups[gid].Nodes[nid].Vnfs = ""
					count++
				} else if entry == "kernel" {
					self.NodeGroups[gid].KernelVersion = value
					self.NodeGroups[gid].Nodes[nid].KernelVersion = ""
					count++
				} else if entry == "domainsuffix" {
					self.NodeGroups[gid].DomainSuffix = value
					self.NodeGroups[gid].Nodes[nid].DomainSuffix = ""
					count++
				} else if entry == "ipxe" {
					self.NodeGroups[gid].Ipxe = value
					self.NodeGroups[gid].Nodes[nid].Ipxe = ""
					count++
				} else if entry == "systemoverlay" {
					self.NodeGroups[gid].SystemOverlay = value
					self.NodeGroups[gid].Nodes[nid].SystemOverlay = ""
					count++
				} else if entry == "runtimeoverlay" {
					self.NodeGroups[gid].RuntimeOverlay = value
					self.NodeGroups[gid].Nodes[nid].RuntimeOverlay = ""
					count++
				} else if entry == "hostname" {
					self.NodeGroups[gid].Nodes[nid].Hostname = value
					count++
				}
			}
		}
	}

	return count, nil
}

func (self *nodeYaml) SetNodeVal(nodeID string, entry string, value string) (int, error) {
	var count int

	for gid, group := range self.NodeGroups {
		for nid := range group.Nodes {
			if nodeID == gid+":"+nid {
				if entry == "vnfs" {
					self.NodeGroups[gid].Nodes[nid].Vnfs = value
					count++
				} else if entry == "kernel" {
					self.NodeGroups[gid].Nodes[nid].KernelVersion = value
					count++
				} else if entry == "domainsuffix" {
					self.NodeGroups[gid].Nodes[nid].DomainSuffix = value
					count++
				} else if entry == "ipxe" {
					self.NodeGroups[gid].Nodes[nid].Ipxe = value
					count++
				} else if entry == "systemoverlay" {
					self.NodeGroups[gid].Nodes[nid].SystemOverlay = value
					count++
				} else if entry == "runtimeoverlay" {
					self.NodeGroups[gid].Nodes[nid].RuntimeOverlay = value
					count++
				} else if entry == "hostname" {
					self.NodeGroups[gid].Nodes[nid].Hostname = value
					count++
				}
			}
		}
	}

	return count, nil
}

func (self *nodeYaml) Persist() error {

	out, err := yaml.Marshal(self)
	if err != nil {
		return err
	}

	fmt.Println(string(out))

	return nil
}

func (self *nodeYaml) FindAllNodes() ([]NodeInfo, error) {
	var ret []NodeInfo

	for groupname, group := range self.NodeGroups {
		for nodename, node := range group.Nodes {
			var n NodeInfo

			n.Id = groupname + ":" + nodename
			n.GroupName = groupname
			n.HostName = node.Hostname
			n.IpmiIpaddr = node.IpmiIpaddr
			n.IpmiUserName = node.IpmiUserName
			n.IpmiPassword = node.IpmiPassword

			n.Vnfs = group.Vnfs
			n.SystemOverlay = group.SystemOverlay
			n.RuntimeOverlay = group.RuntimeOverlay
			n.KernelVersion = group.KernelVersion
			n.KernelArgs = group.KernelArgs
			n.DomainName = group.DomainSuffix
			n.Ipxe = group.Ipxe
			n.NetDevs = node.NetDevs

			if node.KernelVersion != "" {
				n.KernelVersion = node.KernelVersion
			}
			if node.Vnfs != "" {
				n.Vnfs = node.Vnfs
			}
			if node.SystemOverlay != "" {
				n.SystemOverlay = node.SystemOverlay
			}
			if node.RuntimeOverlay != "" {
				n.RuntimeOverlay = node.RuntimeOverlay
			}
			if node.DomainSuffix != "" {
				n.DomainName = node.DomainSuffix
			}
			if node.Ipxe != "" {
				n.Ipxe = node.Ipxe
			}
			if node.KernelArgs != "" {
				n.KernelArgs = node.KernelArgs
			}

			if n.RuntimeOverlay == "" {
				n.RuntimeOverlay = "default"
			}
			if n.SystemOverlay == "" {
				n.SystemOverlay = "default"
			}
			if n.Ipxe == "" {
				n.Ipxe = "default"
			}

			if n.DomainName != "" {
				n.Fqdn = node.Hostname + "." + n.DomainName
			} else {
				n.Fqdn = node.Hostname
			}

			util.ValidateOrDie(n.Fqdn, "group name", n.GroupName, "^[a-zA-Z0-9-._]+$")
			util.ValidateOrDie(n.Fqdn, "vnfs", n.Vnfs, "^[a-zA-Z0-9-._:/]+$")
			util.ValidateOrDie(n.Fqdn, "system overlay", n.SystemOverlay, "^[a-zA-Z0-9-._]+$")
			util.ValidateOrDie(n.Fqdn, "runtime overlay", n.RuntimeOverlay, "^[a-zA-Z0-9-._]+$")
			util.ValidateOrDie(n.Fqdn, "domain suffix", n.DomainName, "^[a-zA-Z0-9-._]+$")
			util.ValidateOrDie(n.Fqdn, "hostname", n.HostName, "^[a-zA-Z0-9-_]+$")
			util.ValidateOrDie(n.Fqdn, "kernel version", n.KernelVersion, "^[a-zA-Z0-9-._]+$")

			ret = append(ret, n)
		}
	}

	return ret, nil
}

func (nodes *nodeYaml) FindByHwaddr(hwa string) (NodeInfo, error) {
	var ret NodeInfo

	n, _ := nodes.FindAllNodes()

	for _, node := range n {
		for _, dev := range node.NetDevs {
			if dev.Hwaddr == hwa {
				return node, nil
			}
		}
	}

	return ret, errors.New("No nodes found with HW Addr: " + hwa)
}

func (nodes *nodeYaml) FindByIpaddr(ipaddr string) (NodeInfo, error) {
	var ret NodeInfo

	n, _ := nodes.FindAllNodes()

	for _, node := range n {
		for _, dev := range node.NetDevs {
			if dev.Ipaddr == ipaddr {
				return node, nil
			}
		}
	}

	return ret, errors.New("No nodes found with IP Addr: " + ipaddr)
}

func (nodes *nodeYaml) SearchByName(search string) ([]NodeInfo, error) {
	var ret []NodeInfo

	n, _ := nodes.FindAllNodes()

	for _, node := range n {
		b, _ := regexp.MatchString(search, node.Fqdn)
		if b == true {
			ret = append(ret, node)
		}
	}

	return ret, nil
}

func (nodes *nodeYaml) SearchByNameList(searchList []string) ([]NodeInfo, error) {
	var ret []NodeInfo

	n, _ := nodes.FindAllNodes()

	for _, search := range searchList {
		for _, node := range n {
			b, _ := regexp.MatchString(search, node.Fqdn)
			if b == true {
				ret = append(ret, node)
			}
		}
	}

	return ret, nil
}
