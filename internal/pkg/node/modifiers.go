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

func (self *nodeYaml) AddNode(groupID string, nodeID string) error {
	var node NodeConf

	wwlog.Printf(wwlog.VERBOSE, "Adding new node: %s/%s\n", groupID, nodeID)

	if _, ok := self.NodeGroups[groupID]; !ok {
		return errors.New("Group does not exist: " + groupID)
	}

	if _, ok := self.NodeGroups[groupID].Nodes[groupID]; ok {
		return errors.New("Nodename already exists in group: " + nodeID)
	}

	self.NodeGroups[groupID].Nodes[nodeID] = &node
	self.NodeGroups[groupID].Nodes[nodeID].Hostname = nodeID

	return nil
}

func (self *nodeYaml) DelNode(groupID string, nodeID string) error {

	if _, ok := self.NodeGroups[groupID]; !ok {
		return errors.New("Group '" + groupID + "' was not found")
	}
	if _, ok := self.NodeGroups[groupID].Nodes[nodeID]; !ok {
		return errors.New("Node '" + nodeID + "' was not found in group '" + groupID + "'")
	}

	wwlog.Printf(wwlog.VERBOSE, "Deleting node network device: %s/%s\n", groupID, nodeID)
	delete(self.NodeGroups[groupID].Nodes, nodeID)

	return nil
}

func(self *nodeYaml) NodeUpdate(node NodeInfo) error {
	groupID := node.Gid.Get()
	nodeID := node.Id.Get()

	if _, ok := self.NodeGroups[groupID]; !ok {
		return errors.New("Group '" + groupID + "' was not found")
	}
	if _, ok := self.NodeGroups[groupID].Nodes[nodeID]; !ok {
		return errors.New("Node '" + nodeID + "' was not found in group '" + groupID + "'")
	}

	self.NodeGroups[groupID].Nodes[nodeID].Hostname = node.HostName.GetReal()
	self.NodeGroups[groupID].Nodes[nodeID].Vnfs = node.Vnfs.GetReal()
	self.NodeGroups[groupID].Nodes[nodeID].DomainName = node.DomainName.GetReal()
	self.NodeGroups[groupID].Nodes[nodeID].Ipxe = node.Ipxe.GetReal()
	self.NodeGroups[groupID].Nodes[nodeID].KernelVersion = node.KernelVersion.GetReal()
	self.NodeGroups[groupID].Nodes[nodeID].KernelArgs = node.KernelArgs.GetReal()
	self.NodeGroups[groupID].Nodes[nodeID].IpmiIpaddr = node.IpmiIpaddr.GetReal()
	self.NodeGroups[groupID].Nodes[nodeID].IpmiUserName = node.IpmiUserName.GetReal()
	self.NodeGroups[groupID].Nodes[nodeID].IpmiPassword = node.IpmiPassword.GetReal()
	self.NodeGroups[groupID].Nodes[nodeID].RuntimeOverlay = node.RuntimeOverlay.GetReal()
	self.NodeGroups[groupID].Nodes[nodeID].SystemOverlay = node.SystemOverlay.GetReal()
	self.NodeGroups[groupID].Nodes[nodeID].Profiles = node.Profiles
	self.NodeGroups[groupID].Nodes[nodeID].NetDevs = node.NetDevs

	return nil
}



/****
 *
 * GROUP MODIFIERS
 *
****/

func (self *nodeYaml) AddGroup(groupID string) error {
	var group GroupConf

	wwlog.Printf(wwlog.VERBOSE, "Adding new group: %s/%s\n", groupID)

	if _, ok := self.NodeGroups[groupID]; ok {
		return errors.New("Group name already exists: " + groupID)
	}

	self.NodeGroups[groupID] = &group
	self.NodeGroups[groupID].DomainName = groupID
	self.NodeGroups[groupID].Profiles = []string{"default"}

	return nil
}

func (self *nodeYaml) DelGroup(groupID string) error {
	if _, ok := self.NodeGroups[groupID]; !ok {

		return errors.New("Group '" + groupID + "' was not found")
	}

	wwlog.Printf(wwlog.VERBOSE, "Deleting group: %s\n", groupID)
	delete(self.NodeGroups, groupID)

	return nil
}

func(self *nodeYaml) GroupUpdate(group GroupInfo) error {
	groupID := group.Id

	if _, ok := self.NodeGroups[groupID]; !ok {
		return errors.New("Group '" + groupID + "' was not found")
	}

	self.NodeGroups[groupID].DomainName = group.DomainName
	self.NodeGroups[groupID].Vnfs = group.Vnfs
	self.NodeGroups[groupID].KernelVersion = group.KernelVersion
	self.NodeGroups[groupID].KernelArgs = group.KernelArgs
	self.NodeGroups[groupID].Ipxe = group.Ipxe
	self.NodeGroups[groupID].IpmiUserName = group.IpmiUserName
	self.NodeGroups[groupID].IpmiPassword = group.IpmiPassword
	self.NodeGroups[groupID].RuntimeOverlay = group.RuntimeOverlay
	self.NodeGroups[groupID].SystemOverlay = group.SystemOverlay
	self.NodeGroups[groupID].Profiles = group.Profiles

	return nil
}




/****
 *
 * PROFILE MODIFIERS
 *
****/

func (self *nodeYaml) AddProfile(profileID string) error {
	var profile ProfileConf

	wwlog.Printf(wwlog.VERBOSE, "Adding new profile: %s/%s\n", profileID)

	if _, ok := self.NodeProfiles[profileID]; ok {
		return errors.New("Profile name already exists: " + profileID)
	}

	self.NodeProfiles[profileID] = &profile

	return nil
}

func (self *nodeYaml) DelProfile(profileID string) error {
	if _, ok := self.NodeProfiles[profileID]; ! ok {
		return errors.New("Group '" + profileID + "' was not found")
	}

	wwlog.Printf(wwlog.VERBOSE, "Deleting profile: %s\n", profileID)
	delete(self.NodeProfiles, profileID)

	return nil
}

func(self *nodeYaml) ProfileUpdate(profile ProfileInfo) error {
	profileID := profile.Id

	if _, ok := self.NodeProfiles[profileID]; ! ok {
		return errors.New("Group '" + profileID + "' was not found")
	}

	self.NodeProfiles[profileID].DomainName = profile.DomainName
	self.NodeProfiles[profileID].Vnfs = profile.Vnfs
	self.NodeProfiles[profileID].Ipxe = profile.Ipxe
	self.NodeProfiles[profileID].KernelVersion = profile.KernelVersion
	self.NodeProfiles[profileID].IpmiUserName = profile.IpmiUserName
	self.NodeProfiles[profileID].IpmiPassword = profile.IpmiPassword
	self.NodeProfiles[profileID].RuntimeOverlay = profile.RuntimeOverlay
	self.NodeProfiles[profileID].SystemOverlay = profile.SystemOverlay

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
