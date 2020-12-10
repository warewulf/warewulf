package node

import (
	"github.com/hpcng/warewulf/internal/pkg/errors"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v2"
	"os"
)

/****
 *
 * NODE MODIFIERS
 *
****/

func (self *nodeYaml) AddNode(nodeID string) (NodeInfo, error) {
	var node NodeConf
	var n NodeInfo

	wwlog.Printf(wwlog.VERBOSE, "Adding new node: %s\n", nodeID)

	if _, ok := self.Nodes[nodeID]; ok {
		return n, errors.New("Nodename already exists: " + nodeID)
	}

	self.Nodes[nodeID] = &node
	self.Nodes[nodeID].Profiles = []string{"default"}
	self.Nodes[nodeID].NetDevs = make(map[string]*NetDevs)

	n.Id.Set(nodeID)
	n.Profiles = []string{"default"}
	n.NetDevs = make(map[string]*NetDevEntry)

	return n, nil
}

func (self *nodeYaml) DelNode(nodeID string) error {

	if _, ok := self.Nodes[nodeID]; !ok {
		return errors.New("Nodename does not exist: " + nodeID)
	}

	wwlog.Printf(wwlog.VERBOSE, "Deleting node: %s\n", nodeID)
	delete(self.Nodes, nodeID)

	return nil
}

func (self *nodeYaml) NodeUpdate(node NodeInfo) error {
	nodeID := node.Id.Get()

	if _, ok := self.Nodes[nodeID]; !ok {
		return errors.New("Nodename does not exist: " + nodeID)
	}

	self.Nodes[nodeID].Comment = node.Comment.GetReal()
	self.Nodes[nodeID].ContainerName = node.ContainerName.GetReal()
	self.Nodes[nodeID].ClusterName = node.ClusterName.GetReal()
	self.Nodes[nodeID].Ipxe = node.Ipxe.GetReal()
	self.Nodes[nodeID].Init = node.Init.GetReal()
	self.Nodes[nodeID].KernelVersion = node.KernelVersion.GetReal()
	self.Nodes[nodeID].KernelArgs = node.KernelArgs.GetReal()
	self.Nodes[nodeID].IpmiIpaddr = node.IpmiIpaddr.GetReal()
	self.Nodes[nodeID].IpmiNetmask = node.IpmiNetmask.GetReal()
	self.Nodes[nodeID].IpmiGateway = node.IpmiGateway.GetReal()
	self.Nodes[nodeID].IpmiUserName = node.IpmiUserName.GetReal()
	self.Nodes[nodeID].IpmiPassword = node.IpmiPassword.GetReal()
	self.Nodes[nodeID].RuntimeOverlay = node.RuntimeOverlay.GetReal()
	self.Nodes[nodeID].SystemOverlay = node.SystemOverlay.GetReal()
	self.Nodes[nodeID].Profiles = node.Profiles
	self.Nodes[nodeID].NetDevs = make(map[string]*NetDevs)

	for devname, netdev := range node.NetDevs {
		var newdev NetDevs
		self.Nodes[nodeID].NetDevs[devname] = &newdev

		self.Nodes[nodeID].NetDevs[devname].Ipaddr = netdev.Ipaddr.GetReal()
		self.Nodes[nodeID].NetDevs[devname].Netmask = netdev.Netmask.GetReal()
		self.Nodes[nodeID].NetDevs[devname].Hwaddr = netdev.Hwaddr.GetReal()
		self.Nodes[nodeID].NetDevs[devname].Gateway = netdev.Gateway.GetReal()
		self.Nodes[nodeID].NetDevs[devname].Type = netdev.Type.GetReal()
		self.Nodes[nodeID].NetDevs[devname].Default = netdev.Default.GetRealB()
	}

	return nil
}

/****
 *
 * PROFILE MODIFIERS
 *
****/

func (self *nodeYaml) AddProfile(profileID string) (NodeInfo, error) {
	var node NodeConf
	var n NodeInfo

	wwlog.Printf(wwlog.VERBOSE, "Adding new profile: %s\n", profileID)

	if _, ok := self.NodeProfiles[profileID]; ok {
		return n, errors.New("Profile name already exists: " + profileID)
	}

	self.NodeProfiles[profileID] = &node

	n.Id.Set(profileID)

	return n, nil
}

func (self *nodeYaml) DelProfile(profileID string) error {

	if _, ok := self.NodeProfiles[profileID]; !ok {
		return errors.New("Profile does not exist: " + profileID)
	}

	wwlog.Printf(wwlog.VERBOSE, "Deleting profile: %s\n", profileID)
	delete(self.NodeProfiles, profileID)

	return nil
}

func (self *nodeYaml) ProfileUpdate(profile NodeInfo) error {
	profileID := profile.Id.Get()

	if _, ok := self.NodeProfiles[profileID]; !ok {
		return errors.New("Profile name does not exist: " + profileID)
	}
	self.NodeProfiles[profileID].Comment = profile.Comment.GetReal()
	self.NodeProfiles[profileID].ContainerName = profile.ContainerName.GetReal()
	self.NodeProfiles[profileID].Ipxe = profile.Ipxe.GetReal()
	self.NodeProfiles[profileID].Init = profile.Init.GetReal()
	self.NodeProfiles[profileID].ClusterName = profile.ClusterName.GetReal()
	self.NodeProfiles[profileID].KernelVersion = profile.KernelVersion.GetReal()
	self.NodeProfiles[profileID].KernelArgs = profile.KernelArgs.GetReal()
	self.NodeProfiles[profileID].IpmiIpaddr = profile.IpmiIpaddr.GetReal()
	self.NodeProfiles[profileID].IpmiNetmask = profile.IpmiNetmask.GetReal()
	self.NodeProfiles[profileID].IpmiGateway = profile.IpmiGateway.GetReal()
	self.NodeProfiles[profileID].IpmiUserName = profile.IpmiUserName.GetReal()
	self.NodeProfiles[profileID].IpmiPassword = profile.IpmiPassword.GetReal()
	self.NodeProfiles[profileID].RuntimeOverlay = profile.RuntimeOverlay.GetReal()
	self.NodeProfiles[profileID].SystemOverlay = profile.SystemOverlay.GetReal()
	self.NodeProfiles[profileID].Profiles = profile.Profiles
	self.NodeProfiles[profileID].NetDevs = make(map[string]*NetDevs)

	for devname, netdev := range profile.NetDevs {
		var newdev NetDevs
		self.NodeProfiles[profileID].NetDevs[devname] = &newdev

		self.NodeProfiles[profileID].NetDevs[devname].Ipaddr = netdev.Ipaddr.GetReal()
		self.NodeProfiles[profileID].NetDevs[devname].Netmask = netdev.Netmask.GetReal()
		self.NodeProfiles[profileID].NetDevs[devname].Hwaddr = netdev.Hwaddr.GetReal()
		self.NodeProfiles[profileID].NetDevs[devname].Gateway = netdev.Gateway.GetReal()
		self.NodeProfiles[profileID].NetDevs[devname].Type = netdev.Type.GetReal()
		self.NodeProfiles[profileID].NetDevs[devname].Default = netdev.Default.GetRealB()
	}

	return nil
}

/****
 *
 * PERSISTENCE
 *
****/

func (self *nodeYaml) Persist() error {

	out, err := yaml.Marshal(self)
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
