package node

import (
	"errors"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"

	"gopkg.in/yaml.v2"
)

var (
	ConfigFile    string
	DefaultConfig string
)

// used as fallback if DefaultConfig can't be read
var FallBackConf = `---
defaultnode:
  runtime overlay:
  - generic
  system overlay:
  - wwinit
  kernel:
    args: quiet crashkernel=no vga=791 net.naming-scheme=v238
  init: /sbin/init
  root: initramfs
  ipxe template: default
  profiles:
  - default
  network devices:
    dummy:
      device: eth0
      type: ethernet
      netmask: 255.255.255.0
      onboot: true`

func init() {
	conf := warewulfconf.Get()
	if ConfigFile == "" {
		ConfigFile = path.Join(conf.Paths.Sysconfdir, "warewulf/nodes.conf")
	}
	if DefaultConfig == "" {
		DefaultConfig = path.Join(conf.Paths.Datadir, "warewulf/defaults.conf")
	}
}

/*
Creates a new nodeDb object from the on-disk configuration
*/
func New() (NodeYaml, error) {
	wwlog.Verbose("Opening node configuration file: %s", ConfigFile)
	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		return NodeYaml{}, err
	}
	return Parse(data)
}

// Parse constructs a new nodeDb object from an input YAML
// document. Passes any errors return from yaml.Unmarshal. Returns an
// error if any parsed value is not of a valid type for the given
// parameter.
func Parse(data []byte) (NodeYaml, error) {
	var ret NodeYaml
	var err error
	wwlog.Debug("Unmarshaling the node configuration")
	err = yaml.Unmarshal(data, &ret)
	if err != nil {
		return ret, err
	}
	wwlog.Debug("Checking nodes for types")
	if ret.Nodes == nil {
		ret.Nodes = map[string]*NodeConf{}
	}
	for nodeName, node := range ret.Nodes {
		err = node.Check()
		if err != nil {
			wwlog.Warn("node: %s parsing error: %s", nodeName, err)
			return ret, err
		}
	}
	if ret.NodeProfiles == nil {
		ret.NodeProfiles = map[string]*NodeConf{}
	}
	for profileName, profile := range ret.NodeProfiles {
		err = profile.Check()
		if err != nil {
			wwlog.Warn("node: %s parsing error: %s", profileName, err)
			return ret, err
		}
	}

	wwlog.Debug("Returning node object")
	return ret, nil
}

/*
Get all the nodes of a configuration. This function also merges
the nodes with the given profiles and set the default values
for every node
*/
func (config *NodeYaml) FindAllNodes() ([]NodeInfo, error) {
	var ret []NodeInfo
	var defConf map[string]*NodeConf
	wwlog.Verbose("Opening defaults from file failed %s\n", DefaultConfig)
	defData, err := os.ReadFile(DefaultConfig)
	if err != nil {
		wwlog.Verbose("Couldn't read DefaultConfig :%s\n", err)
		wwlog.Verbose("Using building defaults")
		defData = []byte(FallBackConf)
	}
	wwlog.Debug("Unmarshalling default config\n")
	err = yaml.Unmarshal(defData, &defConf)
	if err != nil {
		wwlog.Verbose("Couldn't unmarshall defaults from file :%s\n", err)
		err = yaml.Unmarshal([]byte(FallBackConf), &defConf)
		if err != nil {
			wwlog.Warn("Could not get any defaults")
		}
	}
	var defConfNet *NetDevs
	if _, ok := defConf["defaultnode"]; ok {
		if _, ok := defConf["defaultnode"].NetDevs["dummy"]; ok {
			defConfNet = defConf["defaultnode"].NetDevs["dummy"]
		}
		defConf["defaultnode"].NetDevs = nil
	}

	wwlog.Debug("Finding all nodes...")
	for nodename, node := range config.Nodes {
		var n NodeInfo

		wwlog.Debug("In node loop: %s", nodename)
		n.NetDevs = make(map[string]*NetDevEntry)
		n.Tags = make(map[string]*Entry)
		n.Kernel = new(KernelEntry)
		n.Ipmi = new(IpmiEntry)
		n.SetDefFrom(defConf["defaultnode"])
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
		// node explicitly nodename field in NodeConf
		n.Id.Set(nodename)
		// backward compatibility
		for keyname, key := range node.Keys {
			node.Tags[keyname] = key
			delete(node.Keys, keyname)
		}
		err = node.Check()
		if err != nil {
			return nil, fmt.Errorf("node: %s check error: %s", nodename, err)
		}
		n.SetFrom(node)
		// only now the netdevs start to exist so that default values can be set
		for _, netdev := range n.NetDevs {
			SetDefFrom(defConfNet, netdev)
		}
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
				wwlog.Warn("Profile not found for node '%s': %s", nodename, profileName)
				continue
			}
			// can't call setFrom() as we have to use SetAlt instead of Set for an Entry
			wwlog.Verbose("Merging profile into node: %s <- %s", nodename, profileName)
			n.SetAltFrom(config.NodeProfiles[profileName], profileName)
		}
		// set default/primary network is just one network exist
		if len(n.NetDevs) >= 1 {
			tmpNets := make([]string, 0, len(n.NetDevs))
			for key := range node.NetDevs {
				tmpNets = append(tmpNets, key)
			}
			sort.Strings(tmpNets)
			// if a value is present in profile or node, default is not visible
			wwlog.Debug("%s setting primary network device: %s", n.Id.Get(), tmpNets[0])
			n.PrimaryNetDev.SetDefault(tmpNets[0])
		}
		if dev, ok := n.NetDevs[n.PrimaryNetDev.Get()]; ok {
			dev.Primary.SetDefaultB(true)
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
		p.NetDevs = make(map[string]*NetDevEntry)
		p.Tags = make(map[string]*Entry)
		p.Kernel = new(KernelEntry)
		p.Ipmi = new(IpmiEntry)
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
		// Merge Keys into Tags for backwards compatibility
		if len(profile.Tags) == 0 {
			profile.Tags = make(map[string]string)
		}

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
