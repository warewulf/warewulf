package node

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/errors"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/vnfs"
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
	NodeGroups map[string]nodeGroup //`yaml:"nodegroups"`
}

type nodeGroup struct {
	Comment        string
	Vnfs           string
	Ipxe           string `yaml:"ipxe template"`
	SystemOverlay  string `yaml:"system overlay""`
	RuntimeOverlay string `yaml:"runtime overlay""`
	DomainSuffix   string `yaml:"domain suffix"`
	KernelVersion  string `yaml:"kernel version"`
	KernelArgs     string `yaml:"kernel args"`
	Nodes          map[string]nodeEntry
}

type nodeEntry struct {
	Hostname       string
	Vnfs           string
	Ipxe           string `yaml:"ipxe template"`
	SystemOverlay  string `yaml:"system overlay"`
	RuntimeOverlay string `yaml:"runtime overlay"`
	DomainSuffix   string `yaml:"domain suffix"`
	KernelVersion  string `yaml:"kernel version"`
	KernelArgs     string `yaml:"kernel args"`
	IpmiIpaddr     string `yaml:"ipmi ipaddr"`
	IpmiUserName   string `yaml:"ipmi username"`
	IpmiPassword   string `yaml:"ipmi password"`
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

type nodeInfoContainer struct {
	Nodes			[]NodeInfo
}


func New() (nodeInfoContainer, error) {
	var c nodeYaml
	var ret nodeInfoContainer

	config := config.New()

	wwlog.Printf(wwlog.DEBUG, "Opening configuration file: %s\n", ConfigFile)
	data, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		fmt.Printf("error reading node configuration file\n")
		return ret, err
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return ret, err
	}

	for groupname, group := range c.NodeGroups {
		for _, node := range group.Nodes {
			var n NodeInfo

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

			util.ValidateOrDie(n.Fqdn, "group name", 		n.GroupName, "^[a-zA-Z0-9-._]+$")
			util.ValidateOrDie(n.Fqdn, "vnfs", 				n.Vnfs, "^[a-zA-Z0-9-._:/]+$")
			util.ValidateOrDie(n.Fqdn, "system overlay", 	n.SystemOverlay, "^[a-zA-Z0-9-._]+$")
			util.ValidateOrDie(n.Fqdn, "runtime overlay", 	n.RuntimeOverlay, "^[a-zA-Z0-9-._]+$")
			util.ValidateOrDie(n.Fqdn, "domain suffix", 	n.DomainName, "^[a-zA-Z0-9-._]+$")
			util.ValidateOrDie(n.Fqdn, "hostname", 			n.HostName, "^[a-zA-Z0-9-_]+$")
			util.ValidateOrDie(n.Fqdn, "kernel version", 	n.KernelVersion, "^[a-zA-Z0-9-._]+$")

			v := vnfs.New(n.Vnfs)
			n.VnfsDir = config.VnfsChroot(v.NameClean())

			ret.Nodes = append(ret.Nodes, n)
		}
	}

	return ret, nil
}




func (nodes *nodeInfoContainer) FindAllNodes() ([]NodeInfo, error) {
	return nodes.Nodes, nil
}


func (nodes *nodeInfoContainer) FindByHwaddr(hwa string) (NodeInfo, error) {
	var ret NodeInfo

	for _, node := range nodes.Nodes {
		for _, dev := range node.NetDevs {
			if dev.Hwaddr == hwa {
				return node, nil
			}
		}
	}

	return ret, errors.New("No nodes found with HW Addr: " + hwa)
}

func (nodes *nodeInfoContainer) FindByIpaddr(ipaddr string) (NodeInfo, error) {
	var ret NodeInfo

	for _, node := range nodes.Nodes {
		for _, dev := range node.NetDevs {
			if dev.Ipaddr == ipaddr {
				return node, nil
			}
		}
	}

	return ret, errors.New("No nodes found with IP Addr: " + ipaddr)
}

func (nodes *nodeInfoContainer) SearchByName(search string) ([]NodeInfo, error) {
	var ret []NodeInfo

	for _, node := range nodes.Nodes {
		b, _ := regexp.MatchString(search, node.Fqdn)
		if b == true {
			ret = append(ret, node)
		}
	}

	return ret, nil
}

func (nodes *nodeInfoContainer) SearchByNameList(searchList []string) ([]NodeInfo, error) {
	var ret []NodeInfo

	for _, search := range searchList {
		for _, node := range nodes.Nodes {
			b, _ := regexp.MatchString(search, node.Fqdn)
			if b == true {
				ret = append(ret, node)
			}
		}
	}

	return ret, nil
}
