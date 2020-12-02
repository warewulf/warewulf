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

func (self *nodeYaml) AddNode(controllerID string, groupID string, nodeID string) error {
	var node NodeConf

	wwlog.Printf(wwlog.VERBOSE, "Adding new node: %s/%s\n", groupID, nodeID)

	if _, ok := self.Controllers[controllerID]; !ok {
		return errors.New("Controller does not exist: " + controllerID)
	}

	if _, ok := self.Controllers[controllerID].NodeGroups[groupID]; !ok {
		return errors.New("Group does not exist: " + groupID)
	}

	if _, ok := self.Controllers[controllerID].NodeGroups[groupID].Nodes[groupID]; ok {
		return errors.New("Nodename already exists in group: " + nodeID)
	}

	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID] = &node
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].Hostname = nodeID

	return nil
}

func (self *nodeYaml) DelNode(controllerID string, groupID string, nodeID string) error {

	if _, ok := self.Controllers[controllerID]; !ok {
		return errors.New("Controller does not exist: " + controllerID)
	}

	if _, ok := self.Controllers[controllerID].NodeGroups[groupID]; !ok {
		return errors.New("Group does not exist: " + groupID)
	}

	if _, ok := self.Controllers[controllerID].NodeGroups[groupID].Nodes[groupID]; ok {
		return errors.New("Nodename does not exist in group: " + nodeID)
	}

	wwlog.Printf(wwlog.VERBOSE, "Deleting node network device: %s/%s\n", groupID, nodeID)
	delete(self.Controllers[controllerID].NodeGroups[groupID].Nodes, nodeID)

	return nil
}

func (self *nodeYaml) NodeUpdate(node NodeInfo) error {
	controllerID := node.Cid.Get()
	groupID := node.Gid.Get()
	nodeID := node.Id.Get()

	if _, ok := self.Controllers[controllerID]; !ok {
		return errors.New("Controller does not exist: " + controllerID)
	}

	if _, ok := self.Controllers[controllerID].NodeGroups[groupID]; !ok {
		return errors.New("Group does not exist: " + groupID)
	}

	if _, ok := self.Controllers[controllerID].NodeGroups[groupID].Nodes[groupID]; !ok {
		return errors.New("Nodename does not exist in group: " + nodeID)
	}

	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].Hostname = node.HostName.GetReal()
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].Vnfs = node.Vnfs.GetReal()
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].DomainName = node.DomainName.GetReal()
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].Ipxe = node.Ipxe.GetReal()
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].KernelVersion = node.KernelVersion.GetReal()
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].KernelArgs = node.KernelArgs.GetReal()
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].IpmiIpaddr = node.IpmiIpaddr.GetReal()
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].IpmiNetmask = node.IpmiNetmask.GetReal()
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].IpmiUserName = node.IpmiUserName.GetReal()
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].IpmiPassword = node.IpmiPassword.GetReal()
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].RuntimeOverlay = node.RuntimeOverlay.GetReal()
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].SystemOverlay = node.SystemOverlay.GetReal()
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].Profiles = node.Profiles
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].NetDevs = node.NetDevs

	return nil
}

/****
 *
 * GROUP MODIFIERS
 *
****/

func (self *nodeYaml) AddGroup(controllerID string, groupID string) error {
	var group GroupConf

	wwlog.Printf(wwlog.VERBOSE, "Adding new group: %s/%s\n", groupID)

	if _, ok := self.Controllers[controllerID]; !ok {
		return errors.New("Controller does not exist: " + controllerID)
	}

	if _, ok := self.Controllers[controllerID].NodeGroups[groupID]; ok {
		return errors.New("Group already exists: " + groupID)
	}

	self.Controllers[controllerID].NodeGroups[groupID] = &group
	self.Controllers[controllerID].NodeGroups[groupID].DomainName = groupID
	self.Controllers[controllerID].NodeGroups[groupID].Profiles = []string{"default"}

	return nil
}

func (self *nodeYaml) DelGroup(controllerID string, groupID string) error {

	if _, ok := self.Controllers[controllerID]; !ok {
		return errors.New("Controller does not exist: " + controllerID)
	}

	if _, ok := self.Controllers[controllerID].NodeGroups[groupID]; !ok {
		return errors.New("Group does not exist: " + groupID)
	}

	wwlog.Printf(wwlog.VERBOSE, "Deleting group: %s\n", groupID)
	delete(self.Controllers[controllerID].NodeGroups, groupID)

	return nil
}

func (self *nodeYaml) GroupUpdate(group GroupInfo) error {
	controllerID := group.Cid
	groupID := group.Id

	if _, ok := self.Controllers[controllerID]; !ok {
		return errors.New("Controller does not exist: " + controllerID)
	}

	if _, ok := self.Controllers[controllerID].NodeGroups[groupID]; !ok {
		return errors.New("Group does not exist: " + groupID)
	}

	self.Controllers[controllerID].NodeGroups[groupID].DomainName = group.DomainName
	self.Controllers[controllerID].NodeGroups[groupID].Vnfs = group.Vnfs
	self.Controllers[controllerID].NodeGroups[groupID].KernelVersion = group.KernelVersion
	self.Controllers[controllerID].NodeGroups[groupID].KernelArgs = group.KernelArgs
	self.Controllers[controllerID].NodeGroups[groupID].Ipxe = group.Ipxe
	self.Controllers[controllerID].NodeGroups[groupID].IpmiNetmask = group.IpmiNetmask
	self.Controllers[controllerID].NodeGroups[groupID].IpmiUserName = group.IpmiUserName
	self.Controllers[controllerID].NodeGroups[groupID].IpmiPassword = group.IpmiPassword
	self.Controllers[controllerID].NodeGroups[groupID].RuntimeOverlay = group.RuntimeOverlay
	self.Controllers[controllerID].NodeGroups[groupID].SystemOverlay = group.SystemOverlay
	self.Controllers[controllerID].NodeGroups[groupID].Profiles = group.Profiles

	return nil
}

/****
 *
 * CONTROLLER MODIFIERS
 *
****/

func (self *nodeYaml) AddController(controllerID string) error {
	var controller ControllerConf
	var group GroupConf

	wwlog.Printf(wwlog.VERBOSE, "Adding new controller: %s/%s\n", controllerID)

	if _, ok := self.Controllers[controllerID]; ok {
		return errors.New("Controller already exists: " + controllerID)
	}

	self.Controllers[controllerID] = &controller
	self.Controllers[controllerID].NodeGroups = make(map[string]*GroupConf)
	self.Controllers[controllerID].NodeGroups["default"] = &group

	return nil
}

func (self *nodeYaml) DelController(controllerID string) error {

	if _, ok := self.Controllers[controllerID]; !ok {
		return errors.New("Controller does not exist: " + controllerID)
	}

	wwlog.Printf(wwlog.VERBOSE, "Deleting controller: %s\n", controllerID)
	delete(self.Controllers, controllerID)

	return nil
}

func (self *nodeYaml) ControllerUpdate(controller ControllerInfo) error {
	controllerID := controller.Id

	if _, ok := self.Controllers[controllerID]; !ok {
		return errors.New("Controller does not exist: " + controllerID)
	}

	self.Controllers[controllerID].Ipaddr = controller.Ipaddr
	self.Controllers[controllerID].Comment = controller.Comment
	self.Controllers[controllerID].Fqdn = controller.Fqdn

	self.Controllers[controllerID].Services.Warewulfd.Port = controller.Services.Warewulfd.Port
	self.Controllers[controllerID].Services.Warewulfd.Secure = controller.Services.Warewulfd.Secure
	self.Controllers[controllerID].Services.Warewulfd.StartCmd = controller.Services.Warewulfd.StartCmd
	self.Controllers[controllerID].Services.Warewulfd.RestartCmd = controller.Services.Warewulfd.RestartCmd
	self.Controllers[controllerID].Services.Warewulfd.EnableCmd = controller.Services.Warewulfd.EnableCmd

	self.Controllers[controllerID].Services.Dhcp.Enabled = controller.Services.Dhcp.Enabled
	self.Controllers[controllerID].Services.Dhcp.ConfigFile = controller.Services.Dhcp.ConfigFile
	self.Controllers[controllerID].Services.Dhcp.RangeStart = controller.Services.Dhcp.RangeStart
	self.Controllers[controllerID].Services.Dhcp.RangeEnd = controller.Services.Dhcp.RangeEnd
	self.Controllers[controllerID].Services.Dhcp.StartCmd = controller.Services.Dhcp.StartCmd
	self.Controllers[controllerID].Services.Dhcp.RestartCmd = controller.Services.Dhcp.RestartCmd
	self.Controllers[controllerID].Services.Dhcp.EnableCmd = controller.Services.Dhcp.EnableCmd

	self.Controllers[controllerID].Services.Nfs.Enabled = controller.Services.Nfs.Enabled
	self.Controllers[controllerID].Services.Nfs.Exports = controller.Services.Nfs.Exports
	self.Controllers[controllerID].Services.Nfs.StartCmd = controller.Services.Nfs.StartCmd
	self.Controllers[controllerID].Services.Nfs.RestartCmd = controller.Services.Nfs.RestartCmd
	self.Controllers[controllerID].Services.Nfs.EnableCmd = controller.Services.Nfs.EnableCmd

	self.Controllers[controllerID].Services.Tftp.Enabled = controller.Services.Tftp.Enabled
	self.Controllers[controllerID].Services.Tftp.TftpRoot = controller.Services.Tftp.TftpRoot
	self.Controllers[controllerID].Services.Tftp.StartCmd = controller.Services.Tftp.StartCmd
	self.Controllers[controllerID].Services.Tftp.RestartCmd = controller.Services.Tftp.RestartCmd
	self.Controllers[controllerID].Services.Tftp.EnableCmd = controller.Services.Tftp.EnableCmd

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
	if _, ok := self.NodeProfiles[profileID]; !ok {
		return errors.New("Group '" + profileID + "' was not found")
	}

	wwlog.Printf(wwlog.VERBOSE, "Deleting profile: %s\n", profileID)
	delete(self.NodeProfiles, profileID)

	return nil
}

func (self *nodeYaml) ProfileUpdate(profile ProfileInfo) error {
	profileID := profile.Id

	if _, ok := self.NodeProfiles[profileID]; !ok {
		return errors.New("Group '" + profileID + "' was not found")
	}

	self.NodeProfiles[profileID].DomainName = profile.DomainName
	self.NodeProfiles[profileID].Vnfs = profile.Vnfs
	self.NodeProfiles[profileID].Ipxe = profile.Ipxe
	self.NodeProfiles[profileID].KernelVersion = profile.KernelVersion
	self.NodeProfiles[profileID].IpmiNetmask = profile.IpmiNetmask
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
