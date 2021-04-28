package node

import (
	"os"

	"github.com/hpcng/warewulf/internal/pkg/errors"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v2"
)

/****
 *
 * NODE MODIFIERS
 *
****/

func (config *nodeYaml) AddNode(nodeID string) (NodeInfo, error) {
	var node NodeConf
	var n NodeInfo

	wwlog.Printf(wwlog.VERBOSE, "Adding new node: %s\n", nodeID)

	if _, ok := config.Nodes[nodeID]; ok {
		return n, errors.New("Nodename already exists: " + nodeID)
	}

	config.Nodes[nodeID] = &node
	config.Nodes[nodeID].Profiles = []string{"default"}
	config.Nodes[nodeID].NetDevs = make(map[string]*NetDevs)

	n.Id.Set(nodeID)
	n.Profiles = []string{"default"}
	n.NetDevs = make(map[string]*NetDevEntry)

	return n, nil
}

func (config *nodeYaml) DelNode(nodeID string) error {

	if _, ok := config.Nodes[nodeID]; !ok {
		return errors.New("Nodename does not exist: " + nodeID)
	}

	wwlog.Printf(wwlog.VERBOSE, "Deleting node: %s\n", nodeID)
	delete(config.Nodes, nodeID)

	return nil
}

func (config *nodeYaml) NodeUpdate(node NodeInfo) error {
	nodeID := node.Id.Get()

	if _, ok := config.Nodes[nodeID]; !ok {
		return errors.New("Nodename does not exist: " + nodeID)
	}

	config.Nodes[nodeID].Comment = node.Comment.GetReal()
	config.Nodes[nodeID].ContainerName = node.ContainerName.GetReal()
	config.Nodes[nodeID].ClusterName = node.ClusterName.GetReal()
	config.Nodes[nodeID].Ipxe = node.Ipxe.GetReal()
	config.Nodes[nodeID].Init = node.Init.GetReal()
	config.Nodes[nodeID].KernelVersion = node.KernelVersion.GetReal()
	config.Nodes[nodeID].KernelArgs = node.KernelArgs.GetReal()
	config.Nodes[nodeID].IpmiIpaddr = node.IpmiIpaddr.GetReal()
	config.Nodes[nodeID].IpmiNetmask = node.IpmiNetmask.GetReal()
	config.Nodes[nodeID].IpmiGateway = node.IpmiGateway.GetReal()
	config.Nodes[nodeID].IpmiUserName = node.IpmiUserName.GetReal()
	config.Nodes[nodeID].IpmiPassword = node.IpmiPassword.GetReal()
	config.Nodes[nodeID].RuntimeOverlay = node.RuntimeOverlay.GetReal()
	config.Nodes[nodeID].SystemOverlay = node.SystemOverlay.GetReal()
	config.Nodes[nodeID].Root = node.Root.GetReal()

	config.Nodes[nodeID].Discoverable = node.Discoverable.GetRealB()

	config.Nodes[nodeID].Profiles = node.Profiles
	config.Nodes[nodeID].NetDevs = make(map[string]*NetDevs)

	for devname, netdev := range node.NetDevs {
		var newdev NetDevs
		config.Nodes[nodeID].NetDevs[devname] = &newdev

		config.Nodes[nodeID].NetDevs[devname].Ipaddr = netdev.Ipaddr.GetReal()
		config.Nodes[nodeID].NetDevs[devname].Netmask = netdev.Netmask.GetReal()
		config.Nodes[nodeID].NetDevs[devname].Hwaddr = netdev.Hwaddr.GetReal()
		config.Nodes[nodeID].NetDevs[devname].Gateway = netdev.Gateway.GetReal()
		config.Nodes[nodeID].NetDevs[devname].Type = netdev.Type.GetReal()
		config.Nodes[nodeID].NetDevs[devname].Default = netdev.Default.GetRealB()
	}

	return nil
}

/****
 *
 * PROFILE MODIFIERS
 *
****/

func (config *nodeYaml) AddProfile(profileID string) (NodeInfo, error) {
	var node NodeConf
	var n NodeInfo

	wwlog.Printf(wwlog.VERBOSE, "Adding new profile: %s\n", profileID)

	if _, ok := config.NodeProfiles[profileID]; ok {
		return n, errors.New("Profile name already exists: " + profileID)
	}

	config.NodeProfiles[profileID] = &node

	n.Id.Set(profileID)

	return n, nil
}

func (config *nodeYaml) DelProfile(profileID string) error {

	if _, ok := config.NodeProfiles[profileID]; !ok {
		return errors.New("Profile does not exist: " + profileID)
	}

	wwlog.Printf(wwlog.VERBOSE, "Deleting profile: %s\n", profileID)
	delete(config.NodeProfiles, profileID)

	return nil
}

func (config *nodeYaml) ProfileUpdate(profile NodeInfo) error {
	profileID := profile.Id.Get()

	if _, ok := config.NodeProfiles[profileID]; !ok {
		return errors.New("Profile name does not exist: " + profileID)
	}
	config.NodeProfiles[profileID].Comment = profile.Comment.GetReal()
	config.NodeProfiles[profileID].ContainerName = profile.ContainerName.GetReal()
	config.NodeProfiles[profileID].Ipxe = profile.Ipxe.GetReal()
	config.NodeProfiles[profileID].Init = profile.Init.GetReal()
	config.NodeProfiles[profileID].ClusterName = profile.ClusterName.GetReal()
	config.NodeProfiles[profileID].KernelVersion = profile.KernelVersion.GetReal()
	config.NodeProfiles[profileID].KernelArgs = profile.KernelArgs.GetReal()
	config.NodeProfiles[profileID].IpmiIpaddr = profile.IpmiIpaddr.GetReal()
	config.NodeProfiles[profileID].IpmiNetmask = profile.IpmiNetmask.GetReal()
	config.NodeProfiles[profileID].IpmiGateway = profile.IpmiGateway.GetReal()
	config.NodeProfiles[profileID].IpmiUserName = profile.IpmiUserName.GetReal()
	config.NodeProfiles[profileID].IpmiPassword = profile.IpmiPassword.GetReal()
	config.NodeProfiles[profileID].RuntimeOverlay = profile.RuntimeOverlay.GetReal()
	config.NodeProfiles[profileID].SystemOverlay = profile.SystemOverlay.GetReal()
	config.NodeProfiles[profileID].Root = profile.Root.GetReal()

	config.NodeProfiles[profileID].Discoverable = profile.Discoverable.GetRealB()

	config.NodeProfiles[profileID].Profiles = profile.Profiles
	config.NodeProfiles[profileID].NetDevs = make(map[string]*NetDevs)

	for devname, netdev := range profile.NetDevs {
		var newdev NetDevs
		config.NodeProfiles[profileID].NetDevs[devname] = &newdev

		config.NodeProfiles[profileID].NetDevs[devname].Ipaddr = netdev.Ipaddr.GetReal()
		config.NodeProfiles[profileID].NetDevs[devname].Netmask = netdev.Netmask.GetReal()
		config.NodeProfiles[profileID].NetDevs[devname].Hwaddr = netdev.Hwaddr.GetReal()
		config.NodeProfiles[profileID].NetDevs[devname].Gateway = netdev.Gateway.GetReal()
		config.NodeProfiles[profileID].NetDevs[devname].Type = netdev.Type.GetReal()
		config.NodeProfiles[profileID].NetDevs[devname].Default = netdev.Default.GetRealB()
	}

	return nil
}

/****
 *
 * PERSISTENCE
 *
****/

func (config *nodeYaml) Persist() error {

	out, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(ConfigFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	defer file.Close()

	_, err = file.WriteString(string(out))
	if err != nil {
		return err
	}

	return nil
}
