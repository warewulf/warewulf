package assets

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/vnfs"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	//	"os"

	//	"os"
	"regexp"

	"github.com/hpcng/warewulf/internal/pkg/errors"
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
	SystemOverlay  string `yaml:"system system-overlay""`
	RuntimeOverlay string `yaml:"runtime system-overlay""`
	DomainSuffix   string `yaml:"domain suffix"`
	KernelVersion  string `yaml:"kernel version"`
	Nodes          map[string]nodeEntry
}

type nodeEntry struct {
	Hostname       string
	Vnfs           string
	Ipxe           string `yaml:"ipxe template"`
	SystemOverlay  string `yaml:"system system-overlay"`
	RuntimeOverlay string `yaml:"runtime system-overlay"`
	DomainSuffix   string `yaml:"domain suffix"`
	KernelVersion  string `yaml:"kernel version"`
	IpmiIpaddr     string `yaml:"ipmi ipaddr"`
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
	NetDevs        map[string]netDevs
}

func FindAllNodes() ([]NodeInfo, error) {
	var c nodeYaml
	var ret []NodeInfo
	config := config.New()


	data, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		fmt.Printf("error reading node configuration file\n")
		return nil, err
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}

	for groupname, group := range c.NodeGroups {
		for _, node := range group.Nodes {
			var n NodeInfo

			n.GroupName = groupname
			n.HostName = node.Hostname

			n.Vnfs = group.Vnfs
			n.SystemOverlay = group.SystemOverlay
			n.RuntimeOverlay = group.RuntimeOverlay
			n.KernelVersion = group.KernelVersion
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

			v := vnfs.New(n.Vnfs)
			n.VnfsDir = config.VnfsChroot(v.NameClean())

			ret = append(ret, n)
		}
	}

	return ret, nil
}

func FindByHwaddr(hwa string) (NodeInfo, error) {
	var ret NodeInfo

	nodeList, err := FindAllNodes()
	if err != nil {
		return ret, err
	}

	for _, node := range nodeList {
		for _, dev := range node.NetDevs {
			if dev.Hwaddr == hwa {
				return node, nil
			}
		}
	}

	return ret, errors.New("No nodes found with HW Addr: " + hwa)
}

func FindByIpaddr(ipaddr string) (NodeInfo, error) {
	var ret NodeInfo

	nodeList, err := FindAllNodes()
	if err != nil {
		return ret, err
	}

	for _, node := range nodeList {
		for _, dev := range node.NetDevs {
			if dev.Ipaddr == ipaddr {
				return node, nil
			}
		}
	}

	return ret, errors.New("No nodes found with IP Addr: " + ipaddr)
}

func SearchByName(search string) ([]NodeInfo, error) {
	var ret []NodeInfo

	nodeList, err := FindAllNodes()
	if err != nil {
		return ret, err
	}

	for _, node := range nodeList {
		b, _ := regexp.MatchString(search, node.Fqdn)
		if b == true {
			ret = append(ret, node)
		}
	}

	return ret, nil
}

/*
func FindAllVnfs() ([]string, error) {
	var ret []string
	set := make(map[string]bool)

	nodeList, err := FindAllNodes()
	if err != nil {
		return ret, err
	}

	for _, node := range nodeList {
		if node.Vnfs != "" {
			set[node.Vnfs] = true
		}
	}

	for entry := range set {
		ret = append(ret, entry)
	}

	return ret, nil
}

func FindAllKernels() ([]string, error) {
	var ret []string
	set := make(map[string]bool)

	nodeList, err := FindAllNodes()
	if err != nil {
		return ret, err
	}

	for _, node := range nodeList {
		if node.KernelVersion != "" {
			set[node.KernelVersion] = true
		}
	}

	for entry := range set {
		ret = append(ret, entry)
	}

	return ret, nil
}

func ListSystemOverlays() ([]string, error) {
	var ret []string
	set := make(map[string]bool)

	nodeList, err := FindAllNodes()
	if err != nil {
		return ret, err
	}

	for _, node := range nodeList {
		if node.SystemOverlay != "" {
			set[node.SystemOverlay] = true
		}
	}

	for entry := range set {
		ret = append(ret, entry)
	}

	return ret, nil
}

func ListRuntimeOverlays() ([]string, error) {
	var ret []string
	set := make(map[string]bool)

	nodeList, err := FindAllNodes()
	if err != nil {
		return ret, err
	}

	for _, node := range nodeList {
		if node.RuntimeOverlay != "" {
			set[node.RuntimeOverlay] = true
		}
	}

	for entry := range set {
		ret = append(ret, entry)
	}

	return ret, nil
}

*/
