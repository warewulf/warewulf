package node

import (
	"errors"
	// "net"
	"os"
	"sort"

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
	wwlog.Verbose("Opening node configuration file: %s", configFile)
	fileData, err := os.ReadFile(configFile)
	if err != nil {
		return NodeYaml{}, err
	}
	return ParseNodeYaml(fileData)
}


// ParseNodeYaml returns a NodeYaml unmarshalled from the input yaml
// document.
func ParseNodeYaml(fileData []byte) (NodeYaml, error) {
	var nodeYaml NodeYaml
	wwlog.Debug("Unmarshaling the node configuration")
	err := yaml.Unmarshal(fileData, &nodeYaml)
	if err != nil {
		return nodeYaml, err
	}

	wwlog.Debug("Returning NodeYaml")
	return nodeYaml, nil
}


// GetAllNodeInfo returns a slice of NodeInfo, including effective
// settings inherited from profiles and default values.
func (config *NodeYaml) GetAllNodeInfo() ([]NodeInfo, error) {
	var allNodeInfo []NodeInfo

	var defaultYaml []byte
	var defaultConf map[string]*NodeConf
	var err error
	defaultYaml, err = os.ReadFile(DefaultConfig)
	if err != nil {
		wwlog.Warn("Error reading %s: %s", DefaultConfig, err)
		wwlog.Verbose("Using built-in defaults")
		defaultYaml = []byte(FallBackConf)
	}
	wwlog.Debug("Unmarshalling default config")
	err = yaml.Unmarshal(defaultYaml, &defaultConf)
	if err != nil {
		wwlog.Warn("Error unmarshalling defaults: %s", err)
		wwlog.Verbose("Using built-in defaults")
		err = yaml.Unmarshal([]byte(FallBackConf), &defaultConf)
		if err != nil {
			wwlog.Warn("Error unmarshalling built-in defaults")
		}
	}

	wwlog.Debug("Getting all nodes...")
	for nodename, nodeConf := range config.Nodes {
		wwlog.Debug("Getting node: %s", nodename)
		nodeConf.CompatibilityUpdate()
		nodeInfo := NewNodeInfo(nodename, nodeConf, defaultConf["defaultnode"], config.NodeProfiles)
		allNodeInfo = append(allNodeInfo, nodeInfo)
	}

	sort.Slice(allNodeInfo, func(i, j int) bool {
		if allNodeInfo[i].ClusterName.Get() < allNodeInfo[j].ClusterName.Get() {
			return true
		} else if allNodeInfo[i].ClusterName.Get() == allNodeInfo[j].ClusterName.Get() {
			if allNodeInfo[i].Id.Get() < allNodeInfo[j].Id.Get() {
				return true
			}
		}
		return false
	})

	return allNodeInfo, nil
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

	nodes, _ := config.GetAllNodeInfo()

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

// 	n, _ := config.GetAllNodeInfo()

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

// 	n, _ := config.GetAllNodeInfo()

// 	for _, node := range n {
// 		for _, dev := range node.NetDevs {
// 			if dev.Ipaddr.Get() == ipaddr {
// 				return node, nil
// 			}
// 		}
// 	}

// 	return ret, errors.New("No nodes found with IP Addr: " + ipaddr)
// }
