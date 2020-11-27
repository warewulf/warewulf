package node

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/errors"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"regexp"
)

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

func (self *nodeYaml) FindAllNodes() ([]NodeInfo, error) {
	var ret []NodeInfo

	for groupname, group := range self.NodeGroups {
		for nodename, node := range group.Nodes {
			var n NodeInfo
			var allProfiles []string

			if node.Disabled == true || group.Disabled == true {
				wwlog.Printf(wwlog.VERBOSE, "Skipping disabled node: %s/%s\n", groupname, nodename)
				continue
			}

			n.Id = nodename
			n.Gid = groupname
			n.GroupName = groupname
			n.HostName = node.Hostname
			n.IpmiIpaddr.Set(node.IpmiIpaddr)

			n.Profiles = node.Profiles

			allProfiles = append(allProfiles, group.Profiles...)
			allProfiles = append(allProfiles, node.Profiles...)

			for _, p := range allProfiles {
				if _, ok := self.NodeProfiles[p]; !ok {
					wwlog.Printf(wwlog.WARN, "Profile not found for node '%s': %s\n", nodename, p)
					continue
				}

				n.Vnfs.profile = self.NodeProfiles[p].Vnfs
				n.KernelVersion.profile = self.NodeProfiles[p].KernelVersion
				n.KernelArgs.profile = self.NodeProfiles[p].KernelArgs
				n.Ipxe.profile = self.NodeProfiles[p].Ipxe
				n.IpmiUserName.profile = self.NodeProfiles[p].IpmiUserName
				n.IpmiPassword.profile = self.NodeProfiles[p].IpmiPassword
				n.DomainName.profile = self.NodeProfiles[p].DomainName
				n.SystemOverlay.profile = self.NodeProfiles[p].SystemOverlay
				n.RuntimeOverlay.profile = self.NodeProfiles[p].RuntimeOverlay
			}

			n.DomainName.value = node.DomainName
			n.DomainName.group = group.DomainName
			n.Vnfs.value = node.Vnfs
			n.KernelVersion.value = node.KernelVersion
			n.KernelArgs.value = node.KernelArgs
			n.Ipxe.value = node.Ipxe
			n.IpmiUserName.value = node.IpmiUserName
			n.IpmiPassword.value = node.IpmiPassword

			n.RuntimeOverlay.def = "default"
			n.SystemOverlay.def = "default"
			n.Ipxe.def = "default"
			n.KernelArgs.def = "crashkernel=no quiet"

			//			config := config.New()
			//			n.VnfsRoot = config.VnfsChroot(vnfs.CleanName(n.Vnfs))

			if n.DomainName.Defined() == true {
				n.Fqdn = n.HostName + "." + n.DomainName.Get()
			} else {
				n.Fqdn = node.Hostname
			}

			n.NetDevs = node.NetDevs

			ret = append(ret, n)
		}
	}

	return ret, nil
}

func (self *nodeYaml) FindAllGroups() ([]GroupInfo, error) {
	var ret []GroupInfo

	for groupname, group := range self.NodeGroups {
		var g GroupInfo

		g.Id = groupname
		g.DomainName = group.DomainName
		g.Disabled = group.Disabled
		g.Comment = group.Comment

		g.Profiles = group.Profiles

		// TODO: Validate or die on all inputs

		ret = append(ret, g)
	}
	return ret, nil
}

func (self *nodeYaml) FindAllProfiles() ([]ProfileInfo, error) {
	var ret []ProfileInfo

	for name, profile := range self.NodeProfiles {
		var p ProfileInfo

		p.Id = name
		p.Comment = profile.Comment
		p.Vnfs = profile.Vnfs
		p.Ipxe = profile.Ipxe
		p.KernelVersion = profile.KernelVersion
		p.KernelArgs = profile.KernelArgs
		p.IpmiUserName = profile.IpmiUserName
		p.IpmiPassword = profile.IpmiPassword
		p.DomainName = profile.DomainName
		p.RuntimeOverlay = profile.RuntimeOverlay
		p.SystemOverlay = profile.SystemOverlay

		// TODO: Validate or die on all inputs

		ret = append(ret, p)
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
