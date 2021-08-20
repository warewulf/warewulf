package node

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/errors"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v2"
)

const ConfigFile = "/etc/warewulf/nodes.conf"

func New() (nodeYaml, error) {
	var ret nodeYaml

	wwlog.Printf(wwlog.DEBUG, "Opening node configuration file: %s\n", ConfigFile)
	data, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		fmt.Printf("error reading node configuration file\n")
		return ret, err
	}

	wwlog.Printf(wwlog.DEBUG, "Unmarshaling the node configuration\n")
	err = yaml.Unmarshal(data, &ret)
	if err != nil {
		return ret, err
	}

	wwlog.Printf(wwlog.DEBUG, "Returning node object\n")

	return ret, nil
}

func (config *nodeYaml) FindAllNodes() ([]NodeInfo, error) {
	var ret []NodeInfo

	wwlog.Printf(wwlog.DEBUG, "Finding all nodes...\n")
	for nodename, node := range config.Nodes {
		var n NodeInfo

		wwlog.Printf(wwlog.DEBUG, "In node loop: %s\n", nodename)
		n.NetDevs = make(map[string]*NetDevEntry)
		n.Keys = make(map[string]*Entry)
		n.SystemOverlay.SetDefault("default")
		n.RuntimeOverlay.SetDefault("default")
		n.Ipxe.SetDefault("default")
		n.Init.SetDefault("/sbin/init")
		n.Root.SetDefault("initramfs")
		n.KernelArgs.SetDefault("quiet crashkernel=no vga=791 rootfstype=rootfs")

		fullname := strings.SplitN(nodename, ".", 2)
		if len(fullname) > 1 {
			n.ClusterName.SetDefault(fullname[1])
		}

		if len(node.Profiles) == 0 {
			n.Profiles = []string{"default"}
		} else {
			n.Profiles = node.Profiles
		}

		n.Id.Set(nodename)
		n.Comment.Set(node.Comment)
		n.ContainerName.Set(node.ContainerName)
		n.KernelVersion.Set(node.KernelVersion)
		n.KernelArgs.Set(node.KernelArgs)
		n.ClusterName.Set(node.ClusterName)
		n.Ipxe.Set(node.Ipxe)
		n.Init.Set(node.Init)
		n.IpmiIpaddr.Set(node.IpmiIpaddr)
		n.IpmiNetmask.Set(node.IpmiNetmask)
		n.IpmiGateway.Set(node.IpmiGateway)
		n.IpmiUserName.Set(node.IpmiUserName)
		n.IpmiPassword.Set(node.IpmiPassword)
		n.IpmiInterface.Set(node.IpmiInterface)
		n.SystemOverlay.Set(node.SystemOverlay)
		n.RuntimeOverlay.Set(node.RuntimeOverlay)
		n.Root.Set(node.Root)

		n.Discoverable.SetB(node.Discoverable)

		for devname, netdev := range node.NetDevs {
			if _, ok := n.NetDevs[devname]; !ok {
				var netdev NetDevEntry
				n.NetDevs[devname] = &netdev
			}

			n.NetDevs[devname].Ipaddr.Set(netdev.Ipaddr)
			n.NetDevs[devname].Netmask.Set(netdev.Netmask)
			n.NetDevs[devname].Hwaddr.Set(netdev.Hwaddr)
			n.NetDevs[devname].Gateway.Set(netdev.Gateway)
			n.NetDevs[devname].Type.Set(netdev.Type)
			n.NetDevs[devname].Default.SetB(netdev.Default)
		}

		for keyname, key := range node.Keys {
			if _, ok := n.Keys[keyname]; !ok {
				var key Entry
				n.Keys[keyname] = &key
			}
			n.Keys[keyname].Set(key)
		}

		for _, p := range n.Profiles {
			if _, ok := config.NodeProfiles[p]; !ok {
				wwlog.Printf(wwlog.WARN, "Profile not found for node '%s': %s\n", nodename, p)
				continue
			}

			wwlog.Printf(wwlog.VERBOSE, "Merging profile into node: %s <- %s\n", nodename, p)

			n.Comment.SetAlt(config.NodeProfiles[p].Comment, p)
			n.ClusterName.SetAlt(config.NodeProfiles[p].ClusterName, p)
			n.ContainerName.SetAlt(config.NodeProfiles[p].ContainerName, p)
			n.KernelVersion.SetAlt(config.NodeProfiles[p].KernelVersion, p)
			n.KernelArgs.SetAlt(config.NodeProfiles[p].KernelArgs, p)
			n.Ipxe.SetAlt(config.NodeProfiles[p].Ipxe, p)
			n.Init.SetAlt(config.NodeProfiles[p].Init, p)
			n.IpmiIpaddr.SetAlt(config.NodeProfiles[p].IpmiIpaddr, p)
			n.IpmiNetmask.SetAlt(config.NodeProfiles[p].IpmiNetmask, p)
			n.IpmiGateway.SetAlt(config.NodeProfiles[p].IpmiGateway, p)
			n.IpmiUserName.SetAlt(config.NodeProfiles[p].IpmiUserName, p)
			n.IpmiPassword.SetAlt(config.NodeProfiles[p].IpmiPassword, p)
			n.IpmiInterface.SetAlt(config.NodeProfiles[p].IpmiInterface, p)
			n.SystemOverlay.SetAlt(config.NodeProfiles[p].SystemOverlay, p)
			n.RuntimeOverlay.SetAlt(config.NodeProfiles[p].RuntimeOverlay, p)
			n.Root.SetAlt(config.NodeProfiles[p].Root, p)

			n.Discoverable.SetAltB(config.NodeProfiles[p].Discoverable, p)

			for devname, netdev := range config.NodeProfiles[p].NetDevs {
				if _, ok := n.NetDevs[devname]; !ok {
					var netdev NetDevEntry
					n.NetDevs[devname] = &netdev
				}
				wwlog.Printf(wwlog.DEBUG, "Updating profile (%s) netdev: %s\n", p, devname)

				n.NetDevs[devname].Ipaddr.SetAlt(netdev.Ipaddr, p)
				n.NetDevs[devname].Netmask.SetAlt(netdev.Netmask, p)
				n.NetDevs[devname].Hwaddr.SetAlt(netdev.Hwaddr, p)
				n.NetDevs[devname].Gateway.SetAlt(netdev.Gateway, p)
				n.NetDevs[devname].Type.SetAlt(netdev.Type, p)
				n.NetDevs[devname].Default.SetAltB(netdev.Default, p)
			}

			for keyname, key := range config.NodeProfiles[p].Keys {
				if _, ok := n.Keys[keyname]; !ok {
					var key Entry
					n.Keys[keyname] = &key
				}
				n.Keys[keyname].SetAlt(key, p)
			}
		}

		ret = append(ret, n)

	}

	sort.Slice(ret, func(i, j int) bool {
		if ret[i].ClusterName.Get() < ret[j].ClusterName.Get() {
			return true
		} else if ret[i].ClusterName.Get() == ret[j].ClusterName.Get() {
			if ret[i].Id.Get() < ret[j].Id.Get() {
				return true
			}
		}
		return false
	})

	return ret, nil
}

func (config *nodeYaml) FindAllProfiles() ([]NodeInfo, error) {
	var ret []NodeInfo

	for name, profile := range config.NodeProfiles {
		var p NodeInfo
		p.NetDevs = make(map[string]*NetDevEntry)
		p.Keys = make(map[string]*Entry)

		p.Id.Set(name)
		p.Comment.Set(profile.Comment)
		p.ContainerName.Set(profile.ContainerName)
		p.Ipxe.Set(profile.Ipxe)
		p.Init.Set(profile.Init)
		p.KernelVersion.Set(profile.KernelVersion)
		p.KernelArgs.Set(profile.KernelArgs)
		p.IpmiNetmask.Set(profile.IpmiNetmask)
		p.IpmiGateway.Set(profile.IpmiGateway)
		p.IpmiUserName.Set(profile.IpmiUserName)
		p.IpmiPassword.Set(profile.IpmiPassword)
		p.IpmiInterface.Set(profile.IpmiInterface)
		p.RuntimeOverlay.Set(profile.RuntimeOverlay)
		p.SystemOverlay.Set(profile.SystemOverlay)
		p.Root.Set(profile.Root)

		p.Discoverable.SetB(profile.Discoverable)

		for devname, netdev := range profile.NetDevs {
			if _, ok := p.NetDevs[devname]; !ok {
				var netdev NetDevEntry
				p.NetDevs[devname] = &netdev
			}

			wwlog.Printf(wwlog.DEBUG, "Updating profile netdev: %s\n", devname)

			p.NetDevs[devname].Ipaddr.Set(netdev.Ipaddr)
			p.NetDevs[devname].Netmask.Set(netdev.Netmask)
			p.NetDevs[devname].Hwaddr.Set(netdev.Hwaddr)
			p.NetDevs[devname].Gateway.Set(netdev.Gateway)
			p.NetDevs[devname].Type.Set(netdev.Type)
			p.NetDevs[devname].Default.SetB(netdev.Default)
		}

		for keyname, key := range profile.Keys {
			if _, ok := p.Keys[keyname]; !ok {
				var key Entry
				p.Keys[keyname] = &key
			}
			p.Keys[keyname].Set(key)
		}

		// TODO: Validate or die on all inputs

		ret = append(ret, p)
	}

	sort.Slice(ret, func(i, j int) bool {
		if ret[i].ClusterName.Get() < ret[j].ClusterName.Get() {
			return true
		} else if ret[i].ClusterName.Get() == ret[j].ClusterName.Get() {
			if ret[i].Id.Get() < ret[j].Id.Get() {
				return true
			}
		}
		return false
	})

	return ret, nil
}

func (config *nodeYaml) FindDiscoverableNode() (NodeInfo, string, error) {
	var ret NodeInfo

	nodes, _ := config.FindAllNodes()

	for _, node := range nodes {
		if !node.Discoverable.GetB() {
			continue
		}
		for netdev, dev := range node.NetDevs {
			if !dev.Hwaddr.Defined() {
				return node, netdev, nil
			}
		}
	}

	return ret, "", errors.New("No unconfigured nodes found")
}

func (config *nodeYaml) FindByHwaddr(hwa string) (NodeInfo, error) {
	var ret NodeInfo

	n, _ := config.FindAllNodes()

	for _, node := range n {
		for _, dev := range node.NetDevs {
			if dev.Hwaddr.Get() == hwa {
				return node, nil
			}
		}
	}

	return ret, errors.New("No nodes found with HW Addr: " + hwa)
}

func (config *nodeYaml) FindByIpaddr(ipaddr string) (NodeInfo, error) {
	var ret NodeInfo

	n, _ := config.FindAllNodes()

	for _, node := range n {
		for _, dev := range node.NetDevs {
			if dev.Ipaddr.Get() == ipaddr {
				return node, nil
			}
		}
	}

	return ret, errors.New("No nodes found with IP Addr: " + ipaddr)
}
