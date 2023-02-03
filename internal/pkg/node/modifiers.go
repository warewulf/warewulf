package node

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

/****
 *
 * NODE MODIFIERS
 *
****/

func (config *NodeYaml) AddNode(nodeID string) (NodeInfo, error) {
	var node NodeConf
	var n NodeInfo

	wwlog.Verbose("Adding new node: %s", nodeID)

	if _, ok := config.Nodes[nodeID]; ok {
		return n, errors.New("Nodename already exists: " + nodeID)
	}

	config.Nodes[nodeID] = &node
	config.Nodes[nodeID].Profiles = []string{"default"}
	config.Nodes[nodeID].NetDevs = make(map[string]*NetDevs)
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

/****
 *
 * PROFILE MODIFIERS
 *
****/

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

/****
 *
 * PERSISTENCE
 *
****/
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
