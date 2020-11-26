package node

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/errors"
	"github.com/hpcng/warewulf/internal/pkg/vnfs"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"regexp"
)

const ConfigFile = "/etc/warewulf/nodes.conf"

func init() {
	//TODO: Check to make sure nodes.conf is found

}

type nodeYaml struct {
	Profiles 		map[string]*profileConfig
	NodeGroups 		map[string]*nodeGroup
	Hwaddrs 		map[string]NodeInfo
}

type profileConfig struct {
	Vnfs           	string `yaml:"vnfs"`
	Ipxe           	string `yaml:"ipxe template,omitempty"`
	KernelVersion  	string `yaml:"kernel version"`
	KernelArgs     	string `yaml:"kernel args"`
	IpmiUserName   	string `yaml:"ipmi username,omitempty"`
	IpmiPassword   	string `yaml:"ipmi password,omitempty"`
	DomainSuffix   	string `yaml:"domain suffix,omitempty"`
	RuntimeOverlay 	[]OverlayEntry `yaml:"runtime overlay files,omitempty"`
	SystemOverlay  	[]OverlayEntry `yaml:"system overlay files,omitempty"`
}

type OverlayEntry struct {
	Path 			string	`yaml:"path,omitempty"`
	File		 	bool	`yaml:"file,omitempty"`
	Dir		 		bool	`yaml:"dir,omitempty"`
	Link 			bool	`yaml:"link,omitempty"`
	Template 		bool	`yaml:"template,omitempty"`
	Mode 			int32 	`yaml:"mode,omitempty"`
	Owner 			string 	`yaml:"owner,omitempty"`
	Group 			string 	`yaml:"group,omitempty"`
	Source 			string 	`yaml:"source,omitempty"`
	Sources 		[]string `yaml:"sources,omitempty"`
}

type nodeGroup struct {
	Comment        	string
	DomainSuffix   	string `yaml:"domain suffix"`
	Profiles   		[]string `yaml:"profiles"`
	Nodes          	map[string]*nodeEntry
}

type nodeEntry struct {
	Hostname       	string `yaml:"hostname,omitempty"`
	Vnfs           	string `yaml:"vnfs"`
	Ipxe           	string `yaml:"ipxe template,omitempty"`
	KernelVersion  	string `yaml:"kernel version"`
	KernelArgs     	string `yaml:"kernel args"`
	IpmiUserName   	string `yaml:"ipmi username,omitempty"`
	IpmiPassword   	string `yaml:"ipmi password,omitempty"`
	DomainSuffix   	string `yaml:"domain suffix,omitempty"`
	IpmiIpaddr     	string `yaml:"ipmi ipaddr,omitempty"`
	Profiles   		[]string `yaml:"profiles"`
	RuntimeOverlay 	[]OverlayEntry `yaml:"system overlay files,omitempty"`
	SystemOverlay  	[]OverlayEntry `yaml:"runtime overlay files,omitempty"`
	NetDevs        	map[string]*NetDevs
}

type NetDevs struct {
	Type    		string `yaml:"type,omitempty"`
	Hwaddr  		string
	Ipaddr  		string
	Netmask 		string
	Gateway 		string `yaml:"gateway,omitempty"`
}

type EntryInfo struct {
	Value 			string
	override		string
	def		 		string
}

type NodeInfo struct {
	Id             	string
	Gid 	       	string
	Uid 		   	string
	GroupName      	string
	HostName       	string
	DomainName     	string
	Fqdn           	string
	Vnfs           	string
	VnfsRoot		string
	Ipxe           	string
	KernelVersion  	string
	KernelArgs     	string
	IpmiIpaddr     	string
	IpmiUserName   	string
	IpmiPassword   	string
	Profiles 		[]string
	RuntimeOverlay 	[]OverlayEntry
	SystemOverlay  	[]OverlayEntry
	NetDevs        	map[string]*NetDevs
}

type GroupInfo struct {
	Id             	string
	GroupName      	string
	DomainName     	string
	Vnfs           	string
	Ipxe           	string
	SystemOverlay  	string
	RuntimeOverlay 	string
	KernelVersion  	string
	KernelArgs     	string
	IpmiUserName   	string
	IpmiPassword   	string
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

		g.Id = groupname
		g.GroupName = groupname
		g.DomainName = group.DomainSuffix

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

			n.Id = nodename
			n.Gid = groupname
			n.GroupName = groupname
			n.HostName = node.Hostname
			n.Profiles = node.Profiles
			n.IpmiIpaddr = node.IpmiIpaddr

			if len(n.Profiles) == 0 {
				n.Profiles = append(n.Profiles, "default")
			}

			for _, p := range n.Profiles {
				if _, ok := self.Profiles[p]; ok {
				} else {
					continue
				}
				if self.Profiles[p].Vnfs != "" {
					n.Vnfs = self.Profiles[p].Vnfs
				}
				if self.Profiles[p].KernelVersion != "" {
					n.KernelVersion = self.Profiles[p].KernelVersion
				}
				if self.Profiles[p].KernelArgs != "" {
					n.KernelArgs = self.Profiles[p].KernelArgs
				}
				if self.Profiles[p].Ipxe != "" {
					n.Ipxe = self.Profiles[p].Ipxe
				}
				if self.Profiles[p].IpmiUserName != "" {
					n.IpmiUserName = self.Profiles[p].IpmiUserName
				}
				if self.Profiles[p].IpmiPassword != "" {
					n.IpmiPassword = self.Profiles[p].IpmiPassword
				}
				if self.Profiles[p].DomainSuffix != "" {
					n.DomainName = self.Profiles[p].DomainSuffix
				}

				for _, ro := range self.Profiles[p].RuntimeOverlay {
					n.RuntimeOverlay = append(n.RuntimeOverlay, ro)
				}
				for _, so := range self.Profiles[p].SystemOverlay {
					n.SystemOverlay = append(n.SystemOverlay, so)
				}
			}

			if node.DomainSuffix != "" {
				n.DomainName = node.DomainSuffix
			} else if group.DomainSuffix != "" {
				n.DomainName = group.DomainSuffix
			}
			if node.Vnfs != "" {
				n.Vnfs = node.Vnfs
			}
			if node.KernelVersion != "" {
				n.KernelVersion = node.KernelVersion
			}
			if node.KernelArgs != "" {
				n.KernelArgs = node.KernelArgs
			}
			if node.Ipxe != "" {
				n.Ipxe = node.Ipxe
			}
			if node.IpmiUserName != "" {
				n.IpmiUserName = node.IpmiUserName
			}
			if node.IpmiPassword != "" {
				n.IpmiPassword = node.IpmiPassword
			}

			if n.Ipxe == "" {
				n.Ipxe = "default"
			}
			if n.KernelArgs == "" {
				n.KernelArgs = "crashkernel=no quiet"
			}

			config := config.New()
			n.VnfsRoot = config.VnfsChroot(vnfs.CleanName(n.Vnfs))

			if n.DomainName != "" {
				n.Fqdn = node.Hostname + "." + n.DomainName
			} else {
				n.Fqdn = node.Hostname
			}

			n.NetDevs = node.NetDevs

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