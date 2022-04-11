package node

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"path"
	"sort"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
	"github.com/hpcng/warewulf/internal/pkg/kernel"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"

	"gopkg.in/yaml.v2"
)

var ConfigFile string

func init() {
	if ConfigFile == "" {
		ConfigFile = path.Join(buildconfig.SYSCONFDIR(), "warewulf/nodes.conf")
	}
}

func New() (nodeYaml, error) {
	var ret nodeYaml

	wwlog.Printf(wwlog.VERBOSE, "Opening node configuration file: %s\n", ConfigFile)
	data, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
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

/*
Get all the nodes of a configuration. This function also merges
the nodes with the given profiles and set the default values
for every node
*/
func (config *nodeYaml) FindAllNodes() ([]NodeInfo, error) {
	var ret []NodeInfo
	wwconfig, err := warewulfconf.New()
	if err != nil {
		return ret, err
	}
	wwlog.Printf(wwlog.DEBUG, "Finding all nodes...\n")
	for nodename, node := range config.Nodes {
		var n NodeInfo

		wwlog.Printf(wwlog.DEBUG, "In node loop: %s\n", nodename)
		n.NetDevs = make(map[string]*NetDevEntry)
		n.Tags = make(map[string]*Entry)
		n.Kernel = new(KernelEntry)
		n.Ipmi = new(IpmiEntry)
		n.SystemOverlay.SetDefault("wwinit")
		n.RuntimeOverlay.SetDefault("generic")
		n.Ipxe.SetDefault("default")
		n.Init.SetDefault("/sbin/init")
		n.Root.SetDefault("initramfs")

		if n.Kernel == nil {
			n.Kernel = &KernelEntry{}
		}
		n.Kernel.Args.SetDefault("quiet crashkernel=no vga=791")

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
		if node.Kernel != nil {
			n.Kernel.Args.Set(node.Kernel.Args)
		}
		n.ClusterName.Set(node.ClusterName)
		n.Ipxe.Set(node.Ipxe)
		n.Init.Set(node.Init)
		if node.Ipmi != nil {
			n.Ipmi.Ipaddr.Set(node.Ipmi.Ipaddr)
			n.Ipmi.Netmask.Set(node.Ipmi.Netmask)
			n.Ipmi.Port.Set(node.Ipmi.Port)
			n.Ipmi.Gateway.Set(node.Ipmi.Gateway)
			n.Ipmi.UserName.Set(node.Ipmi.UserName)
			n.Ipmi.Password.Set(node.Ipmi.Password)
			n.Ipmi.Interface.Set(node.Ipmi.Interface)
			n.Ipmi.Write.SetB(node.Ipmi.Write)
		}
		n.SystemOverlay.SetSlice(node.SystemOverlay)
		n.RuntimeOverlay.SetSlice(node.RuntimeOverlay)
		n.Root.Set(node.Root)
		n.AssetKey.Set(node.AssetKey)
		n.Discoverable.Set(node.Discoverable)

		if node.Kernel != nil {
			if node.Kernel.Override != "" {
				n.Kernel.Override.Set(node.Kernel.Override)
			} else if node.Kernel.Version != "" {
				n.Kernel.Override.Set(node.Kernel.Version)
			}
		}

		for devname, netdev := range node.NetDevs {
			if _, ok := n.NetDevs[devname]; !ok {
				var netdev NetDevEntry
				n.NetDevs[devname] = &netdev
			}
			n.NetDevs[devname].Device.Set(netdev.Device)
			n.NetDevs[devname].Ipaddr.Set(netdev.Ipaddr)
			n.NetDevs[devname].Ipaddr6.Set(netdev.Ipaddr6)

			// Derive value of ipv6 address from ipv4 if not explicitly set
			if wwconfig.Ipaddr6 != "" && netdev.Ipaddr != "" {
				ipv4Arr := strings.Split(netdev.Ipaddr, ".")
				// error can be ignored as check was done at init
				_, ipv6Net, _ := net.ParseCIDR(wwconfig.Ipaddr6)
				mSize, _ := ipv6Net.Mask.Size()
				ipv6str := fmt.Sprintf("%s%s:%s:%s:%s/%v",
					ipv6Net.IP.String(), ipv4Arr[0], ipv4Arr[1], ipv4Arr[2], ipv4Arr[3], mSize)
				if strings.Count(ipv6Net.IP.String(), ":") == 5 {
					ipv6str = strings.Replace(ipv6str, "::", ":", -1)
				}
				n.NetDevs[devname].Ipaddr6.SetDefault(ipv6str)
			}
			n.NetDevs[devname].Netmask.Set(netdev.Netmask)
			n.NetDevs[devname].Netmask.SetDefault("255.255.255.0")
			n.NetDevs[devname].Hwaddr.Set(netdev.Hwaddr)
			n.NetDevs[devname].Gateway.Set(netdev.Gateway)
			n.NetDevs[devname].Type.Set(netdev.Type)
			n.NetDevs[devname].OnBoot.Set(netdev.OnBoot)
			n.NetDevs[devname].Default.Set(netdev.Default)
			// for just one netdev, it is always the default
			if len(node.NetDevs) == 1 {
				n.NetDevs[devname].Default.Set("true")
			}
			n.NetDevs[devname].Tags = make(map[string]*Entry)
			for keyname, key := range netdev.Tags {
				if _, ok := n.Tags[keyname]; !ok {
					var keyVar Entry
					n.NetDevs[devname].Tags[keyname] = &keyVar
				}
				n.NetDevs[devname].Tags[keyname].Set(key)
			}

		}

		// Merge Keys into Tags for backwards compatibility
		if len(node.Tags) == 0 {
			node.Tags = make(map[string]string)
		}
		for keyname, key := range node.Keys {
			node.Tags[keyname] = key
			delete(node.Keys, keyname)
		}

		for keyname, key := range node.Tags {
			if _, ok := n.Tags[keyname]; !ok {
				var key Entry
				n.Tags[keyname] = &key
			}
			n.Tags[keyname].Set(key)
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
			if config.NodeProfiles[p].Kernel != nil {
				n.Kernel.Args.SetAlt(config.NodeProfiles[p].Kernel.Args, p)
			}
			n.Ipxe.SetAlt(config.NodeProfiles[p].Ipxe, p)
			n.Init.SetAlt(config.NodeProfiles[p].Init, p)
			if config.NodeProfiles[p].Ipmi != nil {
				n.Ipmi.Ipaddr.SetAlt(config.NodeProfiles[p].Ipmi.Ipaddr, p)
				n.Ipmi.Netmask.SetAlt(config.NodeProfiles[p].Ipmi.Netmask, p)
				n.Ipmi.Port.SetAlt(config.NodeProfiles[p].Ipmi.Port, p)
				n.Ipmi.Gateway.SetAlt(config.NodeProfiles[p].Ipmi.Gateway, p)
				n.Ipmi.UserName.SetAlt(config.NodeProfiles[p].Ipmi.UserName, p)
				n.Ipmi.Password.SetAlt(config.NodeProfiles[p].Ipmi.Password, p)
				n.Ipmi.Interface.SetAlt(config.NodeProfiles[p].Ipmi.Interface, p)
				n.Ipmi.Write.SetB(config.NodeProfiles[p].Ipmi.Write)
			}
			n.SystemOverlay.SetAltSlice(config.NodeProfiles[p].SystemOverlay, p)
			n.RuntimeOverlay.SetAltSlice(config.NodeProfiles[p].RuntimeOverlay, p)
			n.Root.SetAlt(config.NodeProfiles[p].Root, p)
			n.AssetKey.SetAlt(config.NodeProfiles[p].AssetKey, p)
			n.Discoverable.SetAlt(config.NodeProfiles[p].Discoverable, p)

			if config.NodeProfiles[p].Kernel != nil {
				if config.NodeProfiles[p].Kernel.Override != "" {
					n.Kernel.Override.SetAlt(config.NodeProfiles[p].Kernel.Override, p)
				} else if config.NodeProfiles[p].Kernel.Version != "" {
					n.Kernel.Override.SetAlt(config.NodeProfiles[p].Kernel.Version, p)
				}
			}

			for devname, netdev := range config.NodeProfiles[p].NetDevs {
				if _, ok := n.NetDevs[devname]; !ok {
					var netdev NetDevEntry
					n.NetDevs[devname] = &netdev
				}
				wwlog.Printf(wwlog.DEBUG, "Updating profile (%s) netdev: %s\n", p, devname)

				n.NetDevs[devname].Device.SetAlt(netdev.Device, p)
				n.NetDevs[devname].Ipaddr.SetAlt(netdev.Ipaddr, p) //FIXME? <- Ipaddr must be uniq
				n.NetDevs[devname].Netmask.SetAlt(netdev.Netmask, p)
				n.NetDevs[devname].Hwaddr.SetAlt(netdev.Hwaddr, p)
				n.NetDevs[devname].Gateway.SetAlt(netdev.Gateway, p)
				n.NetDevs[devname].Type.SetAlt(netdev.Type, p)
				n.NetDevs[devname].OnBoot.SetAlt(netdev.OnBoot, p)
				n.NetDevs[devname].Default.SetAlt(netdev.Default, p)
				if len(netdev.Tags) != 0 {
					for keyname, key := range netdev.Tags {
						if _, ok := n.Tags[keyname]; !ok {
							var keyVar Entry
							n.NetDevs[devname].Tags[keyname] = &keyVar
						}
						n.NetDevs[devname].Tags[keyname].SetAlt(key, p)
					}
				}
			}

			// Merge Keys into Tags for backwards compatibility
			if len(config.NodeProfiles[p].Tags) == 0 {
				config.NodeProfiles[p].Tags = make(map[string]string)
			}
			for keyname, key := range config.NodeProfiles[p].Keys {
				config.NodeProfiles[p].Tags[keyname] = key
				delete(config.NodeProfiles[p].Keys, keyname)
			}

			for keyname, key := range config.NodeProfiles[p].Tags {
				if _, ok := n.Tags[keyname]; !ok {
					var key Entry
					n.Tags[keyname] = &key
				}
				n.Tags[keyname].SetAlt(key, p)
			}
		}
		// set default name of kernel to the kernelname of the container
		if n.ContainerName.Get() != "" {
			listKernel, err := kernel.ListKernels()
			if err != nil {
				for _, kern := range listKernel {
					if kern == n.ContainerName.Get() {
						n.Kernel.Override.SetDefault(n.ContainerName.Get())
					}
				}
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
		p.Tags = make(map[string]*Entry)
		p.Kernel = new(KernelEntry)
		p.Ipmi = new(IpmiEntry)
		p.Id.Set(name)
		p.Comment.Set(profile.Comment)
		p.ClusterName.Set(profile.ClusterName)
		p.ContainerName.Set(profile.ContainerName)
		p.Ipxe.Set(profile.Ipxe)
		p.Init.Set(profile.Init)
		if profile.Kernel != nil {
			p.Kernel.Args.Set(profile.Kernel.Args)
			if profile.Kernel.Override != "" {
				p.Kernel.Override.Set(profile.Kernel.Override)
			} else if profile.Kernel.Version != "" {
				p.Kernel.Override.Set(profile.Kernel.Version)
			}
		}
		if profile.Ipmi != nil {
			p.Ipmi.Netmask.Set(profile.Ipmi.Netmask)
			p.Ipmi.Port.Set(profile.Ipmi.Port)
			p.Ipmi.Gateway.Set(profile.Ipmi.Gateway)
			p.Ipmi.UserName.Set(profile.Ipmi.UserName)
			p.Ipmi.Password.Set(profile.Ipmi.Password)
			p.Ipmi.Interface.Set(profile.Ipmi.Interface)
			p.Ipmi.Write.SetB(profile.Ipmi.Write)
		}
		p.RuntimeOverlay.SetSlice(profile.RuntimeOverlay)
		p.SystemOverlay.SetSlice(profile.SystemOverlay)
		p.Root.Set(profile.Root)
		p.AssetKey.Set(profile.AssetKey)
		p.Discoverable.Set(profile.Discoverable)

		for devname, netdev := range profile.NetDevs {
			if _, ok := p.NetDevs[devname]; !ok {
				var netdev NetDevEntry
				p.NetDevs[devname] = &netdev
			}

			wwlog.Printf(wwlog.DEBUG, "Updating profile netdev: %s\n", devname)

			p.NetDevs[devname].Device.Set(netdev.Device)
			p.NetDevs[devname].Netmask.Set(netdev.Netmask)
			p.NetDevs[devname].Gateway.Set(netdev.Gateway)
			p.NetDevs[devname].Type.Set(netdev.Type)
			p.NetDevs[devname].OnBoot.Set(netdev.OnBoot)
			p.NetDevs[devname].Default.Set(netdev.Default)

			// The following should not be set in a profile.
			if netdev.Ipaddr != "" {
				wwlog.Printf(wwlog.WARN, "Ignoring ip address %v in profile %v\n", netdev.Ipaddr, name)
			}
			if netdev.Hwaddr != "" {
				wwlog.Printf(wwlog.WARN, "Ignoring hardware address %v in profile %v\n", netdev.Hwaddr, name)
			}
			p.NetDevs[devname].Tags = make(map[string]*Entry)
			for keyname, key := range netdev.Tags {
				if _, ok := p.Tags[keyname]; !ok {
					var keyVar Entry
					p.NetDevs[devname].Tags[keyname] = &keyVar
				}
				p.NetDevs[devname].Tags[keyname].Set(key)
			}

		}

		// Merge Keys into Tags for backwards compatibility
		if len(profile.Tags) == 0 {
			profile.Tags = make(map[string]string)
		}
		for keyname, key := range profile.Keys {
			profile.Tags[keyname] = key
			delete(profile.Keys, keyname)
		}

		for keyname, key := range profile.Tags {
			if _, ok := p.Tags[keyname]; !ok {
				var key Entry
				p.Tags[keyname] = &key
			}
			p.Tags[keyname].Set(key)
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

	return ret, "", errors.New("no unconfigured nodes found")
}
