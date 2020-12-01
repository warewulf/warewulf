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

	for controllername, controller := range self.Controllers {
		for groupname, group := range controller.NodeGroups {
			for nodename, node := range group.Nodes {
				var n NodeInfo
				var allProfiles []string

				if node.Disabled == true || group.Disabled == true {
					wwlog.Printf(wwlog.VERBOSE, "Skipping disabled node: %s/%s\n", groupname, nodename)
					continue
				}

				n.Id.Set(nodename)
				n.Gid.Set(groupname)
				n.Cid.Set(controllername)
				n.HostName.Set(node.Hostname)
				n.IpmiIpaddr.Set(node.IpmiIpaddr)
				n.DomainName.Set(node.DomainName)
				n.Vnfs.Set(node.Vnfs)
				n.KernelVersion.Set(node.KernelVersion)
				n.KernelArgs.Set(node.KernelArgs)
				n.Ipxe.Set(node.Ipxe)
				n.IpmiUserName.Set(node.IpmiUserName)
				n.IpmiPassword.Set(node.IpmiPassword)
				n.SystemOverlay.Set(node.SystemOverlay)
				n.RuntimeOverlay.Set(node.RuntimeOverlay)

				n.DomainName.SetGroup(group.DomainName)
				n.Vnfs.SetGroup(group.Vnfs)
				n.KernelVersion.SetGroup(group.KernelVersion)
				n.KernelArgs.SetGroup(group.KernelArgs)
				n.Ipxe.SetGroup(group.Ipxe)
				n.IpmiUserName.SetGroup(group.IpmiUserName)
				n.IpmiPassword.SetGroup(group.IpmiPassword)
				n.SystemOverlay.SetGroup(group.SystemOverlay)
				n.RuntimeOverlay.SetGroup(group.RuntimeOverlay)

				n.RuntimeOverlay.SetDefault("default")
				n.SystemOverlay.SetDefault("default")
				n.Ipxe.SetDefault("default")
				n.KernelArgs.SetDefault("crashkernel=no quiet")

				n.GroupProfiles = group.Profiles
				n.Profiles = node.Profiles

				allProfiles = append(allProfiles, group.Profiles...)
				allProfiles = append(allProfiles, node.Profiles...)

				for _, p := range allProfiles {
					if _, ok := self.NodeProfiles[p]; !ok {
						wwlog.Printf(wwlog.WARN, "Profile not found for node '%s': %s\n", nodename, p)
						continue
					}

					n.DomainName.SetProfile(self.NodeProfiles[p].DomainName)
					n.Vnfs.SetProfile(self.NodeProfiles[p].Vnfs)
					n.KernelVersion.SetProfile(self.NodeProfiles[p].KernelVersion)
					n.KernelArgs.SetProfile(self.NodeProfiles[p].KernelArgs)
					n.Ipxe.SetProfile(self.NodeProfiles[p].Ipxe)
					n.IpmiUserName.SetProfile(self.NodeProfiles[p].IpmiUserName)
					n.IpmiPassword.SetProfile(self.NodeProfiles[p].IpmiPassword)
					n.SystemOverlay.SetProfile(self.NodeProfiles[p].SystemOverlay)
					n.RuntimeOverlay.SetProfile(self.NodeProfiles[p].RuntimeOverlay)
				}

				if n.DomainName.Defined() == true {
					n.Fqdn.Set(node.Hostname + "." + n.DomainName.Get())
				} else {
					n.Fqdn.Set(node.Hostname)
				}

				n.NetDevs = node.NetDevs

				ret = append(ret, n)
			}
		}
	}

	return ret, nil
}

func (self *nodeYaml) FindAllGroups() ([]GroupInfo, error) {
	var ret []GroupInfo

	for controllername, controller := range self.Controllers {
		for groupname, group := range controller.NodeGroups {
			var g GroupInfo

			g.Id = groupname
			g.Cid = controllername
			g.DomainName = group.DomainName
			g.Comment = group.Comment
			g.Vnfs = group.Vnfs
			g.KernelVersion = group.KernelVersion
			g.KernelArgs = group.KernelArgs
			g.IpmiPassword = group.IpmiPassword
			g.IpmiUserName = group.IpmiUserName
			g.SystemOverlay = group.SystemOverlay
			g.RuntimeOverlay = group.RuntimeOverlay

			g.Profiles = group.Profiles

			// TODO: Validate or die on all inputs

			ret = append(ret, g)
		}
	}
	return ret, nil
}

func (self *nodeYaml) FindAllControllers() ([]ControllerInfo, error) {
	var ret []ControllerInfo

	for controllername, controller := range self.Controllers {
		var c ControllerInfo

		c.Id = controllername
		c.Ipaddr = controller.Ipaddr
		c.Services = struct {
			Warewulfd struct {
				Port       string
				Secure     string
				StartCmd   string
				RestartCmd string
				EnableCmd  string
			}
			Dhcp struct {
				Enabled    bool
				RangeStart string
				RangeEnd   string
				ConfigFile string
				StartCmd   string
				RestartCmd string
				EnableCmd  string
			}
			Tftp struct {
				Enabled    bool
				TftpRoot   string
				StartCmd   string
				RestartCmd string
				EnableCmd  string
			}
			Nfs struct {
				Enabled    bool
				Exports    []string
				StartCmd   string
				RestartCmd string
				EnableCmd  string
			}
		}(controller.Services)

		// TODO: Validate or die on all inputs

		ret = append(ret, c)
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
		b, _ := regexp.MatchString(search, node.Fqdn.Get())
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
			b, _ := regexp.MatchString(search, node.Fqdn.Get())
			if b == true {
				ret = append(ret, node)
			}
		}
	}

	return ret, nil
}
