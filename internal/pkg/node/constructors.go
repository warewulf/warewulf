package node

import (
	"bytes"
	"encoding/gob"
	"os"
	"path"
	"sort"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"

	"gopkg.in/yaml.v2"
)

var (
	ConfigFile string
)

/*
Creates a new nodeDb object from the on-disk configuration
*/
func New() (NodeYaml, error) {
	conf := warewulfconf.Get()
	if ConfigFile == "" {
		ConfigFile = path.Join(conf.Paths.Sysconfdir, "warewulf/nodes.conf")
	}
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
	var nodeList NodeYaml
	var err error
	wwlog.Debug("Unmarshaling the node configuration")
	err = yaml.Unmarshal(data, &nodeList)
	if err != nil {
		return nodeList, err
	}
	wwlog.Debug("Checking profiles for types")
	if nodeList.Nodes == nil {
		nodeList.Nodes = map[string]*NodeConf{}
	}
	/*
		for nodeName, node := range nodeList.Nodes {
			err = node.Check()
			if err != nil {
				wwlog.Warn("node: %s parsing error: %s", nodeName, err)
				return nodeList, err
			}
		}
	*/
	if nodeList.NodeProfiles == nil {
		nodeList.NodeProfiles = map[string]*ProfileConf{}
	}
	/*
		for profileName, profile := range nodeList.NodeProfiles {
			err = profile.Check()
			if err != nil {
				wwlog.Warn("node: %s parsing error: %s", profileName, err)
				return nodeList, err
			}
		}
	*/
	wwlog.Debug("Returning node object")
	return nodeList, nil
}

/*
Get a node with its merged in profiles
*/
func (config *NodeYaml) GetNode(id string) (node NodeConf, err error) {
	if _, ok := config.Nodes[id]; !ok {
		return node, ErrNotFound
	}
	wwlog.Debug("constructing node: %s", id)
	for _, p := range config.Nodes[id].Profiles {
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		dec := gob.NewDecoder(&buf)
		includedProfile, err := config.GetProfile(p)
		if err != nil {
			return node, err
		}
		err = enc.Encode(includedProfile)
		if err != nil {
			return node, err
		}
		err = dec.Decode(&node)
		if err != nil {
			return node, err
		}
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	err = enc.Encode(config.Nodes[id])
	if err != nil {
		return node, err
	}
	err = dec.Decode(&node)
	if err != nil {
		return node, err
	}
	// finally set no exported values
	node.id = id
	node.valid = true
	if netdev, ok := node.NetDevs[node.PrimaryNetDev]; ok {
		netdev.primary = true
	} else {
		keys := make([]string, 0)
		for k := range node.NetDevs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		if len(keys) > 0 {
			wwlog.Debug("%s: no primary defined, sanitizing to: %s", id, keys[0])
			node.NetDevs[keys[0]].primary = true
			node.PrimaryNetDev = keys[0]
		}
	}
	return
}

/*
Get the profile with all the profiles merged in
*/
func (config *NodeYaml) GetProfile(id string) (profile NodeConf, err error) {
	if _, ok := config.NodeProfiles[id]; !ok {
		return profile, ErrNotFound
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	// finally merge in the real profile
	err = enc.Encode(config.NodeProfiles[id])
	if err != nil {
		return profile, err
	}
	err = dec.Decode(&profile)
	if err != nil {
		return profile, err
	}
	// finally set no exported values
	profile.id = id
	return
}

/*
Get the profiles from the loaded configuration. This function also merges
the profiles with the given profiles.
*/
func (config *NodeYaml) FindAllNodes(profiles ...string) (nodeList []NodeConf, err error) {
	if len(profiles) == 0 {
		for n := range config.Nodes {
			profiles = append(profiles, n)
		}
	}
	wwlog.Debug("Finding profiles: %s", profiles)
	for _, profileId := range profiles {
		node, err := config.GetNode(profileId)
		if err != nil {
			return nodeList, err
		}
		nodeList = append(nodeList, node)
	}
	sort.Slice(nodeList, func(i, j int) bool {
		if nodeList[i].ClusterName < nodeList[j].ClusterName {
			return true
		} else if nodeList[i].ClusterName == nodeList[j].ClusterName {
			if nodeList[i].id < nodeList[j].id {
				return true
			}
		}
		return false
	})
	return nodeList, nil
}

/*
Return all profiles as NodeInfo
*/
func (config *NodeYaml) FindAllProfiles(profiles ...string) (profileList []NodeConf, err error) {
	if len(profiles) == 0 {
		for n := range config.NodeProfiles {
			profiles = append(profiles, n)
		}
	}
	wwlog.Debug("Finding profiles: %s", profiles)
	for _, profileId := range profiles {
		node, err := config.GetProfile(profileId)
		if err != nil {
			return profileList, err
		}
		profileList = append(profileList, node)
	}
	sort.Slice(profileList, func(i, j int) bool {
		if profileList[i].ClusterName < profileList[j].ClusterName {
			return true
		} else if profileList[i].ClusterName == profileList[j].ClusterName {
			if profileList[i].id < profileList[j].id {
				return true
			}
		}
		return false
	})

	return profileList, nil
}

/*
Return the names of all available nodes
*/
func (config *NodeYaml) ListAllNodes() []string {
	nodeList := make([]string, len(config.Nodes))
	for name := range config.Nodes {
		nodeList = append(nodeList, name)
	}
	return nodeList
}

/*
Return the names of all available profiles
*/
func (config *NodeYaml) ListAllProfiles() []string {
	var nodeList []string
	for name := range config.NodeProfiles {
		nodeList = append(nodeList, name)
	}
	return nodeList
}

/*
FindDiscoverableNode returns the first discoverable node and an
interface to associate with the discovered interface. If the node has
a primary interface, it is returned; otherwise, the first interface
without a hardware address is returned.

If no unconfigured node is found, an error is returned.
*/
func (config *NodeYaml) FindDiscoverableNode() (string, string, error) {

	nodes, _ := config.FindAllNodes()

	for _, node := range nodes {
		if !node.Discoverable {
			continue
		}
		if _, ok := node.NetDevs[node.PrimaryNetDev]; ok {
			return node.Id(), node.PrimaryNetDev, nil
		}
		for netdev, dev := range node.NetDevs {
			if dev.Hwaddr != "" {
				return node.Id(), netdev, nil
			}
		}
	}

	return "", "", ErrNoUnconfigured
}
