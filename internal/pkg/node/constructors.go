package node

import (
	"bytes"
	"encoding/gob"
	"os"
	"path"
	"sort"
	"fmt"

	"dario.cat/mergo"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"

	"gopkg.in/yaml.v3"
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
func Parse(data []byte) (nodeList NodeYaml, err error) {
	wwlog.Debug("Unmarshaling the node configuration")
	err = yaml.Unmarshal(data, &nodeList)
	if err != nil {
		return nodeList, err
	}
	wwlog.Debug("Checking nodes for types")
	if nodeList.nodes == nil {
		nodeList.nodes = map[string]*NodeConf{}
	}
	if nodeList.nodeProfiles == nil {
		nodeList.nodeProfiles = map[string]*ProfileConf{}
	}
	wwlog.Debug("returning node object")
	return nodeList, nil
}

/*
Get a node with its merged in nodes
*/
func (config *NodeYaml) GetNode(id string) (node NodeConf, err error) {
	if _, ok := config.nodes[id]; !ok {
		return node, ErrNotFound
	}
	node = EmptyNode()
	// create a deep copy of the node, as otherwise pointers
	// and not their contents is merged
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	err = enc.Encode(config.nodes[id])
	if err != nil {
		return node, err
	}
	err = dec.Decode(&node)
	if err != nil {
		return node, err
	}

	// Collect all profiles, including nested ones, before building the node config
	visitedProfiles := make(map[string]bool)
	var allProfiles []string
	for _, p := range cleanList(config.nodes[id].Profiles) {
		profiles, err := config.collectProfiles(p, visitedProfiles)
		if err != nil {
			wwlog.Warn("error collecting profiles for %s: %v", p, err)
			continue
		}
		allProfiles = append(allProfiles, profiles...)
	}

	// Remove duplicates and sort the profiles alphanumerically
	uniqueProfiles := removeDuplicates(allProfiles)
	sort.Strings(uniqueProfiles)

	// Merge profiles into the node configuration
	for _, p := range uniqueProfiles {
		includedProfile, err := config.GetProfile(p)
		if err != nil {
			wwlog.Warn("profile not found: %s", p)
			continue
		}
		err = mergo.Merge(&node.ProfileConf, includedProfile, mergo.WithAppendSlice)
		if err != nil {
			return node, err
		}
	}

	// Set unexported node fields
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
	wwlog.Debug("constructed node: %s", id)
	return
}

// collectProfiles recursively collects all profiles for a given profile ID
func (config *NodeYaml) collectProfiles(profileID string, visited map[string]bool) ([]string, error) {
	if visited[profileID] {
		return nil, fmt.Errorf("cycle detected with profile: %s", profileID)
	}
	visited[profileID] = true

	profiles := []string{profileID}
	profile, err := config.GetProfile(profileID)
	if err != nil {
		return nil, err
	}

	for _, p := range cleanList(profile.Profiles) {
		nestedProfiles, err := config.collectProfiles(p, visited)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, nestedProfiles...)
	}
	return profiles, nil
}

// removeDuplicates removes duplicate entries from a slice of strings
func removeDuplicates(slice []string) []string {
	seen := make(map[string]struct{})
	result := []string{}
	for _, item := range slice {
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

/*
Return the node with the id string without the merged in nodes, return ErrNotFound
otherwise
*/
func (config *NodeYaml) GetNodeOnly(id string) (node NodeConf, err error) {
	node = EmptyNode()
	if found, ok := config.nodes[id]; ok {
		return *found, nil
	}
	return node, ErrNotFound
}

/*
Return pointer to the  node with the id string without the merged in nodes, return ErrMotFound
otherwise
*/
func (config *NodeYaml) GetNodeOnlyPtr(id string) (*NodeConf, error) {
	node := EmptyNode()
	if found, ok := config.nodes[id]; ok {
		return found, nil
	}
	return &node, ErrNotFound
}

/*
Get the profile with id, return ErrNotFound otherwise
*/
func (config *NodeYaml) GetProfile(id string) (profile ProfileConf, err error) {
	if found, ok := config.nodeProfiles[id]; ok {
		found.id = id
		return *found, nil
	}
	return profile, ErrNotFound
}

/*
Get the profile with id, return ErrNotFound otherwise
*/
func (config *NodeYaml) GetProfilePtr(id string) (profile *ProfileConf, err error) {
	if found, ok := config.nodeProfiles[id]; ok {
		found.id = id
		return found, nil
	}
	return profile, ErrNotFound
}

/*
Get the nodes from the loaded configuration. This function also merges
the nodes with the given nodes.
*/
func (config *NodeYaml) FindAllNodes(nodes ...string) (nodeList []NodeConf, err error) {
	if len(nodes) == 0 {
		for n := range config.nodes {
			nodes = append(nodes, n)
		}
	}
	wwlog.Debug("Finding nodes: %s", nodes)
	for _, nodeId := range nodes {
		node, err := config.GetNode(nodeId)
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
Return all nodes as ProfileConf
*/
func (config *NodeYaml) FindAllProfiles(nodes ...string) (profileList []ProfileConf, err error) {
	if len(nodes) == 0 {
		for n := range config.nodeProfiles {
			nodes = append(nodes, n)
		}
	}
	wwlog.Debug("Finding nodes: %s", nodes)
	for _, profileId := range nodes {
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
	nodeList := make([]string, len(config.nodes))
	for name := range config.nodes {
		nodeList = append(nodeList, name)
	}
	return nodeList
}

/*
Return the names of all available nodes
*/
func (config *NodeYaml) ListAllProfiles() []string {
	var nodeList []string
	for name := range config.nodeProfiles {
		nodeList = append(nodeList, name)
	}
	return nodeList
}

/*
FindDiscoverableNode returns the first discoverable node and an
interface to associate with the discovered interface. If the nodUNDEFe has
a primary interface, it is returned; otherwise, the first interface
without a hardware address is returned.

If no unconfigured node is found, an error is returned.
*/
func (config *NodeYaml) FindDiscoverableNode() (NodeConf, string, error) {

	nodes, _ := config.FindAllNodes()

	for _, node := range nodes {
		if !(node.Discoverable.Bool()) {
			continue
		}
		if _, ok := node.NetDevs[node.PrimaryNetDev]; ok {
			return node, node.PrimaryNetDev, nil
		}
		for netdev, dev := range node.NetDevs {
			if dev.Hwaddr != "" {
				return node, netdev, nil
			}
		}
	}

	return EmptyNode(), "", ErrNoUnconfigured
}
