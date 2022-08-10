package node

import (
	"errors"
	"io/ioutil"
	"path"
	"sort"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"

	"gopkg.in/yaml.v2"
)

var ConfigFile string

func init() {
	if ConfigFile == "" {
		ConfigFile = path.Join(buildconfig.SYSCONFDIR(), "warewulf/nodes.conf")
	}
}

func New() (NodeYaml, error) {
	var ret NodeYaml

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
func (config *NodeYaml) FindAllNodes() ([]NodeInfo, error) {
	var ret []NodeInfo
	/*
		wwconfig, err := warewulfconf.New()
		if err != nil {
			return ret, err
		}
	*/
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
		n.Kernel.Args.SetDefault("quiet crashkernel=no vga=791")

		fullname := strings.SplitN(nodename, ".", 2)
		if len(fullname) > 1 {
			n.ClusterName.SetDefault(fullname[1])
		}
		// special handling for profile to get the default one
		if len(node.Profiles) == 0 {
			n.Profiles.SetSlice([]string{"default"})
		} else {
			n.Profiles.SetSlice(node.Profiles)
		}
		// node explciti nodename field in NodeConf
		n.Id.Set(nodename)
		// backward compatibilty
		for keyname, key := range node.Keys {
			node.Tags[keyname] = key
			delete(node.Keys, keyname)
		}
		n.SetFrom(node)
		// backward compatibility
		n.Ipmi.Ipaddr.Set(node.IpmiIpaddr)
		n.Ipmi.Netmask.Set(node.IpmiNetmask)
		n.Ipmi.Port.Set(node.IpmiPort)
		n.Ipmi.Gateway.Set(node.IpmiGateway)
		n.Ipmi.UserName.Set(node.IpmiUserName)
		n.Ipmi.Password.Set(node.IpmiPassword)
		n.Ipmi.Interface.Set(node.IpmiInterface)
		n.Ipmi.Write.Set(node.IpmiWrite)
		n.Kernel.Args.Set(node.KernelArgs)
		n.Kernel.Override.Set(node.KernelOverride)
		n.Kernel.Override.Set(node.KernelVersion)
		// delete deprecated structures so that they do not get unmarshalled
		node.IpmiIpaddr = ""
		node.IpmiNetmask = ""
		node.IpmiGateway = ""
		node.IpmiUserName = ""
		node.IpmiPassword = ""
		node.IpmiInterface = ""
		node.IpmiWrite = ""
		node.KernelArgs = ""
		node.KernelOverride = ""
		node.KernelVersion = ""
		// Merge Keys into Tags for backwards compatibility
		if len(node.Tags) == 0 {
			node.Tags = make(map[string]string)
		}

		for _, profileName := range n.Profiles.GetSlice() {
			if _, ok := config.NodeProfiles[profileName]; !ok {
				wwlog.Printf(wwlog.WARN, "Profile not found for node '%s': %s\n", nodename, profileName)
				continue
			}
			// can't call setFrom() as we have to use SetAlt instead of Set for an Entry
			wwlog.Printf(wwlog.VERBOSE, "Merging profile into node: %s <- %s\n", nodename, profileName)
			config.NodeProfiles[profileName].SetAltFrom(n, profileName)
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

/*
Return all profiles as NodeInfo
*/
func (config *NodeYaml) FindAllProfiles() ([]NodeInfo, error) {
	var ret []NodeInfo

	for name, profile := range config.NodeProfiles {
		var p NodeInfo
		p.Id.Set(name)
		for keyname, key := range profile.Keys {
			profile.Tags[keyname] = key
			delete(profile.Keys, keyname)
		}
		p.SetFrom(profile)
		p.Ipmi.Ipaddr.Set(profile.IpmiIpaddr)
		p.Ipmi.Netmask.Set(profile.IpmiNetmask)
		p.Ipmi.Port.Set(profile.IpmiPort)
		p.Ipmi.Gateway.Set(profile.IpmiGateway)
		p.Ipmi.UserName.Set(profile.IpmiUserName)
		p.Ipmi.Password.Set(profile.IpmiPassword)
		p.Ipmi.Interface.Set(profile.IpmiInterface)
		p.Ipmi.Write.Set(profile.IpmiWrite)
		p.Kernel.Args.Set(profile.KernelArgs)
		p.Kernel.Override.Set(profile.KernelOverride)
		p.Kernel.Override.Set(profile.KernelVersion)
		// delete deprecated stuff
		profile.IpmiIpaddr = ""
		profile.IpmiNetmask = ""
		profile.IpmiGateway = ""
		profile.IpmiUserName = ""
		profile.IpmiPassword = ""
		profile.IpmiInterface = ""
		profile.IpmiWrite = ""
		profile.KernelArgs = ""
		profile.KernelOverride = ""
		profile.KernelVersion = ""
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

/*
Return the names of all available profiles
*/
func (config *NodeYaml) ListAllProfiles() []string {
	var ret []string
	for name := range config.NodeProfiles {
		ret = append(ret, name)
	}
	return ret
}

func (config *NodeYaml) FindDiscoverableNode() (NodeInfo, string, error) {
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
