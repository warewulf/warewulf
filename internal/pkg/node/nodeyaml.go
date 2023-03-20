package node

import (
	"errors"
	// "net"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

// A NodeYaml represents the top-level node configuration as read from
// disk. Contains a map of nodes and profiles, each indexed by name.
type NodeYaml struct {
	WWInternal   int `yaml:"WW_INTERNAL"`
	NodeProfiles map[string]*NodeConf
	Nodes        map[string]*NodeConf
}


// ReadNodeYaml returns a NodeYaml representing the current state of
// nodes.conf as read from its default location.
//
// The default location is read from the package variable ConfigFile.
func ReadNodeYaml() (NodeYaml, error) {
	return ReadNodeYamlFromFile(ConfigFile)
}


// ReadNodeYamlFromFile returns a NodeYaml representing the current
// state of nodes.conf as read from the specified configFile path.
func ReadNodeYamlFromFile(configFile string) (NodeYaml, error) {
	var nodeYaml NodeYaml
	wwlog.Verbose("Opening node configuration file: %s", configFile)
	fileData, err := os.ReadFile(configFile)
	if err != nil {
		return nodeYaml, err
	}

	wwlog.Debug("Unmarshaling the node configuration")
	err = yaml.Unmarshal(fileData, &nodeYaml)
	if err != nil {
		return nodeYaml, err
	}

	wwlog.Debug("Returning NodeYaml")
	return nodeYaml, nil
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
	var defConf map[string]*NodeConf
	wwlog.Verbose("Opening defaults from file failed %s\n", DefaultConfig)
	defData, err := os.ReadFile(DefaultConfig)
	if err != nil {
		wwlog.Verbose("Couldn't read DefaultConfig :%s\n", err)
	}
	wwlog.Debug("Unmarshalling default config\n")
	err = yaml.Unmarshal(defData, &defConf)
	if err != nil {
		wwlog.Verbose("Couldn't unmarshall defaults from file :%s\n", err)
		wwlog.Verbose("Using building defaults")
		err = yaml.Unmarshal([]byte(FallBackConf), &defConf)
		if err != nil {
			wwlog.Warn("Could not get any defaults")
		}
	}
	var defConfNet *NetDevConf
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
		n.SetFrom(node)
		// only now the netdevs start to exist so that default values can be set
		for _, netdev := range n.NetDevs {
			netdev.SetDefFrom(defConfNet)
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


func (config *NodeYaml) AddNode(nodeID string) (NodeInfo, error) {
	var node NodeConf
	var n NodeInfo

	wwlog.Verbose("Adding new node: %s", nodeID)

	if _, ok := config.Nodes[nodeID]; ok {
		return n, errors.New("Nodename already exists: " + nodeID)
	}

	config.Nodes[nodeID] = &node
	config.Nodes[nodeID].Profiles = []string{"default"}
	config.Nodes[nodeID].NetDevs = make(map[string]*NetDevConf)
	n.Id.Set(nodeID)
	n.Profiles.SetSlice([]string{"default"})
	n.NetDevs = make(map[string]*NetDevEntry)
	n.Ipmi = new(IpmiEntry)
	n.Kernel = new(KernelEntry)

	return n, nil
}

func (config *NodeYaml) DelNode(nodeID string) error {

	if _, ok := config.Nodes[nodeID]; !ok {
		return errors.New("Nodename does not exist: " + nodeID)
	}

	wwlog.Verbose("Deleting node: %s", nodeID)
	delete(config.Nodes, nodeID)

	return nil
}

func (config *NodeYaml) NodeUpdate(node NodeInfo) error {
	nodeID := node.Id.Get()

	if _, ok := config.Nodes[nodeID]; !ok {
		return errors.New("Nodename does not exist: " + nodeID)
	}
	config.Nodes[nodeID].GetRealFrom(node)
	return nil
}


func (config *NodeYaml) AddProfile(profileID string) (NodeInfo, error) {
	var node NodeConf
	var n NodeInfo

	wwlog.Verbose("Adding new profile: %s", profileID)

	if _, ok := config.NodeProfiles[profileID]; ok {
		return n, errors.New("Profile name already exists: " + profileID)
	}

	config.NodeProfiles[profileID] = &node

	n.Id.Set(profileID)

	return n, nil
}

func (config *NodeYaml) DelProfile(profileID string) error {

	if _, ok := config.NodeProfiles[profileID]; !ok {
		return errors.New("Profile does not exist: " + profileID)
	}

	wwlog.Verbose("Deleting profile: %s", profileID)
	delete(config.NodeProfiles, profileID)

	return nil
}

/*
Update the the config for the given profile so that it can unmarshalled.
*/
func (config *NodeYaml) ProfileUpdate(profile NodeInfo) error {
	profileID := profile.Id.Get()

	if _, ok := config.NodeProfiles[profileID]; !ok {
		return errors.New("Profile name does not exist: " + profileID)
	}
	config.NodeProfiles[profileID].GetRealFrom(profile)
	return nil
}


/*
Write the the NodeYaml to disk.
*/
func (config *NodeYaml) Persist() error {
	// flatten out profiles and nodes
	for _, val := range config.NodeProfiles {
		val.Flatten()
	}
	for _, val := range config.Nodes {
		val.Flatten()
	}
	out, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(ConfigFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		wwlog.Error("%s", err)
		os.Exit(1)
	}

	defer file.Close()

	_, err = file.WriteString(string(out))
	if err != nil {
		return err
	}

	return nil
}


// func (config *NodeYaml) FindByHwaddr(hwa string) (NodeInfo, error) {
// 	if _, err := net.ParseMAC(hwa); err != nil {
// 		return NodeInfo{}, errors.New("invalid hardware address: " + hwa)
// 	}

// 	var ret NodeInfo

// 	n, _ := config.FindAllNodes()

// 	for _, node := range n {
// 		for _, dev := range node.NetDevs {
// 			if strings.EqualFold(dev.Hwaddr.Get(), hwa) {
// 				return node, nil
// 			}
// 		}
// 	}

// 	return ret, errors.New("No nodes found with HW Addr: " + hwa)
// }


// func (config *NodeYaml) FindByIpaddr(ipaddr string) (NodeInfo, error) {
// 	if net.ParseIP(ipaddr) == nil {
// 		return NodeInfo{}, errors.New("invalid IP:" + ipaddr)
// 	}

// 	var ret NodeInfo

// 	n, _ := config.FindAllNodes()

// 	for _, node := range n {
// 		for _, dev := range node.NetDevs {
// 			if dev.Ipaddr.Get() == ipaddr {
// 				return node, nil
// 			}
// 		}
// 	}

// 	return ret, errors.New("No nodes found with IP Addr: " + ipaddr)
// }
