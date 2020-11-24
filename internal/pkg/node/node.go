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
	NodeGroups 		map[string]*nodeGroup //`yaml:"nodegroups"`
	Hwaddrs 		map[string]NodeInfo
}

type nodeGroup struct {
	Comment        string
	Vnfs           string `yaml:"vnfs"`
	Ipxe           string `yaml:"ipxe template,omitempty"`
	SystemOverlay  string `yaml:"system overlay,omitempty""`
	RuntimeOverlay string `yaml:"runtime overlay""`
	DomainSuffix   string `yaml:"domain suffix"`
	KernelVersion  string `yaml:"kernel version"`
	KernelArgs     string `yaml:"kernel args"`
	IpmiUserName   string `yaml:"ipmi username,omitempty"`
	IpmiPassword   string `yaml:"ipmi password,omitempty"`
	Nodes          map[string]*nodeEntry
}

type nodeEntry struct {
	Hostname       string `yaml:"hostname,omitempty"`
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
	NetDevs        map[string]*netDevs
}

type netDevs struct {
	Type    string `yaml:"type,omitempty"`
	Hwaddr  string
	Ipaddr  string
	Netmask string
	Gateway string `yaml:"gateway,omitempty"`
}

type EntryInfo struct {
	Value 			string
	override		string
	def		 		string
}

type NodeInfo struct {
	Id             EntryInfo
	Gid 	       EntryInfo
	Uid 		   EntryInfo
	GroupName      EntryInfo
	HostName       EntryInfo
	DomainName     EntryInfo
	Fqdn           EntryInfo
	Vnfs           EntryInfo
	Ipxe           EntryInfo
	SystemOverlay  EntryInfo
	RuntimeOverlay EntryInfo
	KernelVersion  EntryInfo
	KernelArgs     EntryInfo
	IpmiIpaddr     EntryInfo
	IpmiUserName   EntryInfo
	IpmiPassword   EntryInfo
	NetDevs        map[string]*netDevs
}

type GroupInfo struct {
	Id             EntryInfo
	GroupName      EntryInfo
	DomainName     EntryInfo
	Vnfs           EntryInfo
	Ipxe           EntryInfo
	SystemOverlay  EntryInfo
	RuntimeOverlay EntryInfo
	KernelVersion  EntryInfo
	KernelArgs     EntryInfo
	IpmiUserName   EntryInfo
	IpmiPassword   EntryInfo
}

func (i *EntryInfo) Default (value string) {
	if value == "" {
		return
	}
	i.def = value
}

func (i *EntryInfo) Set (value string) {
	if value == "" {
		return
	}
	i.Value = value
	i.override = ""
}

func (i *EntryInfo) Override (value string) {
	if value == "" {
		return
	}
	i.override = value
}

func (i *EntryInfo) String () string {
	if i.override != "" {
		return i.override
	} else if i.Value != "" {
		return i.Value
	}
	return i.def
}

func (i *EntryInfo) Defined () bool {
	if i.String() == "" {
		return false
	}
	return true
}

func (i *EntryInfo) Fprint () string {
	if i.override != "" {
		return "["+ i.override +"]"
	} else if i.Value != "" {
		return i.Value
	} else if i.def != "" {
		return "("+ i.def +")"
	}
	return "--"
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


func (self *nodeYaml) FindAllGroups() ([]GroupInfo, error) {
	var ret []GroupInfo

	for groupname, group := range self.NodeGroups {
		var g GroupInfo

		g.Id.Set(groupname)
		g.GroupName.Set(groupname)
		g.RuntimeOverlay.Set(group.RuntimeOverlay)
		g.SystemOverlay.Set(group.SystemOverlay)
		g.Ipxe.Set(group.Ipxe)
		g.KernelVersion.Set(group.KernelVersion)
		g.KernelArgs.Set(group.KernelArgs)
		g.Vnfs.Set(group.Vnfs)
		g.IpmiUserName.Set(group.IpmiUserName)
		g.IpmiPassword.Set(group.IpmiPassword)
		g.DomainName.Set(group.DomainSuffix)

		// TODO: Validate or die on all inputs

		ret = append(ret, g)
	}
	return ret, nil
}


func (self *nodeYaml) FindAllNodes() ([]NodeInfo, error) {
	var ret []NodeInfo

	for groupname, group := range self.NodeGroups {
		for nodename, node := range group.Nodes {
			var n NodeInfo

			n.Id.Set(nodename)
			n.Gid.Set(groupname)
			n.GroupName.Set(groupname)
			n.HostName.Set(node.Hostname)
			n.IpmiIpaddr.Set(node.IpmiIpaddr)

			n.Vnfs.Set(group.Vnfs)
			n.SystemOverlay.Set(group.SystemOverlay)
			n.RuntimeOverlay.Set(group.RuntimeOverlay)
			n.KernelVersion.Set(group.KernelVersion)
			n.KernelArgs.Set(group.KernelArgs)
			n.DomainName.Set(group.DomainSuffix)
			n.Ipxe.Set(group.Ipxe)
			n.IpmiUserName.Set(group.IpmiUserName)
			n.IpmiPassword.Set(group.IpmiPassword)

			n.Vnfs.Override(node.Vnfs)
			n.SystemOverlay.Override(node.SystemOverlay)
			n.RuntimeOverlay.Override(node.RuntimeOverlay)
			n.KernelVersion.Override(node.KernelVersion)
			n.KernelArgs.Override(node.KernelArgs)
			n.DomainName.Override(node.DomainSuffix)
			n.Ipxe.Override(node.Ipxe)
			n.IpmiUserName.Override(node.IpmiUserName)
			n.IpmiPassword.Override(node.IpmiPassword)

			n.RuntimeOverlay.Default("default")
			n.SystemOverlay.Default("default")
			n.Ipxe.Default("default")

			if n.DomainName.Defined() == true {
				if group.DomainSuffix != "" {
					n.Fqdn.Set(node.Hostname + "." + group.DomainSuffix)
				} else if node.DomainSuffix != "" {
					n.Fqdn.Set(node.Hostname + "." + node.DomainSuffix)
				} else {
					n.Fqdn.Set(node.Hostname)
				}
			}

			n.NetDevs = node.NetDevs

			util.ValidateOrDie(n.Fqdn.String() +":group name", n.GroupName.String(), "^[a-zA-Z0-9-._]*$")
			util.ValidateOrDie(n.Fqdn.String() +":vnfs", n.Vnfs.String(), "^[a-zA-Z0-9-._:/]*$")
			util.ValidateOrDie(n.Fqdn.String() +":system overlay", n.SystemOverlay.String(), "^[a-zA-Z0-9-._]*$")
			util.ValidateOrDie(n.Fqdn.String() +":runtime overlay", n.RuntimeOverlay.String(), "^[a-zA-Z0-9-._]*$")
			util.ValidateOrDie(n.Fqdn.String() +":domain suffix", n.DomainName.String(), "^[a-zA-Z0-9-._]*$")
			util.ValidateOrDie(n.Fqdn.String() +":hostname", n.HostName.String(), "^[a-zA-Z0-9-_]*$")
			util.ValidateOrDie(n.Fqdn.String() +":kernel version", n.KernelVersion.String(), "^[a-zA-Z0-9-._]*$")

			ret = append(ret, n)
		}
	}

	return ret, nil
}

func (self *nodeYaml) FindByHwaddr(hwa string) (NodeInfo, error) {
	var ret NodeInfo

	n, _ := self.FindAllNodes()

	for _, node := range n {
		for _, dev := range node.NetDevs {
			if dev.Hwaddr == hwa {
				return node, nil
			}
		}
	}

	return ret, errors.New("No nodes found with HW Addr: " + hwa)
}

func (self *nodeYaml) FindByIpaddr(ipaddr string) (NodeInfo, error) {
	var ret NodeInfo

	n, _ := self.FindAllNodes()

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
		b, _ := regexp.MatchString(search, node.Fqdn.String())
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
			b, _ := regexp.MatchString(search, node.Fqdn.String())
			if b == true {
				ret = append(ret, node)
			}
		}
	}

	return ret, nil
}
