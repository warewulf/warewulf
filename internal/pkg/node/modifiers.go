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
	n.Ipmi = new(IpmiEntry)
	n.Kernel = new(KernelEntry)

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

	if node.Kernel != nil && (node.Kernel.Override.GotReal() || node.Kernel.Args.GotReal()) {
		config.Nodes[nodeID].Kernel = new(KernelConf)
		config.Nodes[nodeID].Kernel.Override = node.Kernel.Override.GetReal()
		config.Nodes[nodeID].Kernel.Args = node.Kernel.Args.GetReal()
	}

	if node.Ipmi != nil && (node.Ipmi.Ipaddr.GotReal() || node.Ipmi.Netmask.GotReal() ||
		node.Ipmi.Port.GotReal() || node.Ipmi.Gateway.GotReal() || node.Ipmi.UserName.GotReal() ||
		node.Ipmi.Password.GotReal() || node.Ipmi.Interface.GotReal() || node.Ipmi.Write.GotReal()) {
		config.Nodes[nodeID].Ipmi = new(IpmiConf)
		config.Nodes[nodeID].Ipmi.Ipaddr = node.Ipmi.Ipaddr.GetReal()
		config.Nodes[nodeID].Ipmi.Netmask = node.Ipmi.Netmask.GetReal()
		config.Nodes[nodeID].Ipmi.Port = node.Ipmi.Port.GetReal()
		config.Nodes[nodeID].Ipmi.Gateway = node.Ipmi.Gateway.GetReal()
		config.Nodes[nodeID].Ipmi.UserName = node.Ipmi.UserName.GetReal()
		config.Nodes[nodeID].Ipmi.Password = node.Ipmi.Password.GetReal()
		config.Nodes[nodeID].Ipmi.Interface = node.Ipmi.Interface.GetReal()
		config.Nodes[nodeID].Ipmi.Write = node.Ipmi.Write.GetB()
	}
	config.Nodes[nodeID].RuntimeOverlay = node.RuntimeOverlay.GetRealSlice()
	config.Nodes[nodeID].SystemOverlay = node.SystemOverlay.GetRealSlice()
	config.Nodes[nodeID].Root = node.Root.GetReal()
	config.Nodes[nodeID].AssetKey = node.AssetKey.GetReal()
	config.Nodes[nodeID].Discoverable = node.Discoverable.GetReal()

	config.Nodes[nodeID].Profiles = node.Profiles

	config.Nodes[nodeID].NetDevs = make(map[string]*NetDevs)
	for devname, netdev := range node.NetDevs {
		var newdev NetDevs
		config.Nodes[nodeID].NetDevs[devname] = &newdev

		config.Nodes[nodeID].NetDevs[devname].Device = netdev.Device.GetReal()
		config.Nodes[nodeID].NetDevs[devname].Ipaddr = netdev.Ipaddr.GetReal()
		config.Nodes[nodeID].NetDevs[devname].Netmask = netdev.Netmask.GetReal()
		config.Nodes[nodeID].NetDevs[devname].Hwaddr = netdev.Hwaddr.GetReal()
		config.Nodes[nodeID].NetDevs[devname].Gateway = netdev.Gateway.GetReal()
		config.Nodes[nodeID].NetDevs[devname].Type = netdev.Type.GetReal()
		config.Nodes[nodeID].NetDevs[devname].OnBoot = netdev.OnBoot.GetReal()
		config.Nodes[nodeID].NetDevs[devname].Default = netdev.Default.GetReal()
		config.Nodes[nodeID].NetDevs[devname].Tags = make(map[string]string)
		for keyname, key := range netdev.Tags {
			if key.GetReal() != "" {
				config.Nodes[nodeID].NetDevs[devname].Tags[keyname] = key.GetReal()
			}
		}
	}

	config.Nodes[nodeID].Tags = make(map[string]string)
	for keyname, key := range node.Tags {
		if key.GetReal() != "" {
			config.Nodes[nodeID].Tags[keyname] = key.GetReal()
		}
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

/*
Update the the config for the given profile so that it can unmarshalled.
*/
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
	if profile.Kernel.Override.GotReal() || profile.Kernel.Args.GotReal() {
		config.NodeProfiles[profileID].Kernel = new(KernelConf)
		config.NodeProfiles[profileID].Kernel.Override = profile.Kernel.Override.GetReal()
		config.NodeProfiles[profileID].Kernel.Args = profile.Kernel.Args.GetReal()
	}
	if profile.Ipmi.Ipaddr.GotReal() || profile.Ipmi.Netmask.GotReal() ||
		profile.Ipmi.Port.GotReal() || profile.Ipmi.Gateway.GotReal() || profile.Ipmi.UserName.GotReal() ||
		profile.Ipmi.Password.GotReal() || profile.Ipmi.Interface.GotReal() || profile.Ipmi.Write.GotReal() {
		config.NodeProfiles[profileID].Ipmi = new(IpmiConf)
		config.NodeProfiles[profileID].Ipmi.Ipaddr = profile.Ipmi.Ipaddr.GetReal()
		config.NodeProfiles[profileID].Ipmi.Netmask = profile.Ipmi.Netmask.GetReal()
		config.NodeProfiles[profileID].Ipmi.Port = profile.Ipmi.Port.GetReal()
		config.NodeProfiles[profileID].Ipmi.Gateway = profile.Ipmi.Gateway.GetReal()
		config.NodeProfiles[profileID].Ipmi.UserName = profile.Ipmi.UserName.GetReal()
		config.NodeProfiles[profileID].Ipmi.Password = profile.Ipmi.Password.GetReal()
		config.NodeProfiles[profileID].Ipmi.Interface = profile.Ipmi.Interface.GetReal()
		config.NodeProfiles[profileID].Ipmi.Write = profile.Ipmi.Interface.GetB()
	}
	config.NodeProfiles[profileID].RuntimeOverlay = profile.RuntimeOverlay.GetRealSlice()
	config.NodeProfiles[profileID].SystemOverlay = profile.SystemOverlay.GetRealSlice()
	config.NodeProfiles[profileID].Root = profile.Root.GetReal()
	config.NodeProfiles[profileID].AssetKey = profile.AssetKey.GetReal()
	config.NodeProfiles[profileID].Discoverable = profile.Discoverable.GetReal()

	config.NodeProfiles[profileID].Profiles = profile.Profiles

	config.NodeProfiles[profileID].NetDevs = make(map[string]*NetDevs)
	for devname, netdev := range profile.NetDevs {
		var newdev NetDevs
		config.NodeProfiles[profileID].NetDevs[devname] = &newdev

		config.NodeProfiles[profileID].NetDevs[devname].Device = netdev.Device.GetReal()
		config.NodeProfiles[profileID].NetDevs[devname].Ipaddr = netdev.Ipaddr.GetReal()
		config.NodeProfiles[profileID].NetDevs[devname].Netmask = netdev.Netmask.GetReal()
		config.NodeProfiles[profileID].NetDevs[devname].Hwaddr = netdev.Hwaddr.GetReal()
		config.NodeProfiles[profileID].NetDevs[devname].Gateway = netdev.Gateway.GetReal()
		config.NodeProfiles[profileID].NetDevs[devname].Type = netdev.Type.GetReal()
		config.NodeProfiles[profileID].NetDevs[devname].OnBoot = netdev.OnBoot.GetReal()
		config.NodeProfiles[profileID].NetDevs[devname].Default = netdev.Default.GetReal()
	}

	config.NodeProfiles[profileID].Tags = make(map[string]string)
	for keyname, key := range profile.Tags {
		config.NodeProfiles[profileID].Tags[keyname] = key.GetReal()
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
