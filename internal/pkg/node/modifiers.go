package node

import (
	"github.com/hpcng/warewulf/internal/pkg/errors"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"strings"

	"gopkg.in/yaml.v2"
	"os"
)

func get2Set(input string) string {
	if strings.ToUpper(input) == "UNDEF" {
		return ""
	}
	return input
}

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

	if _, ok := self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID]; !ok {
		return errors.New("Nodename does not exist in group: " + nodeID)
	}

	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].Hostname = get2Set(node.HostName.GetNode())
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].Vnfs = get2Set(node.Vnfs.GetNode())
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].DomainName = get2Set(node.DomainName.GetNode())
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].Ipxe = get2Set(node.Ipxe.GetNode())
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].KernelVersion = get2Set(node.KernelVersion.GetNode())
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].KernelArgs = get2Set(node.KernelArgs.GetNode())
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].IpmiIpaddr = get2Set(node.IpmiIpaddr.GetNode())
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].IpmiNetmask = get2Set(node.IpmiNetmask.GetNode())
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].IpmiUserName = get2Set(node.IpmiUserName.GetNode())
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].IpmiPassword = get2Set(node.IpmiPassword.GetNode())
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].RuntimeOverlay = get2Set(node.RuntimeOverlay.GetNode())
	self.Controllers[controllerID].NodeGroups[groupID].Nodes[nodeID].SystemOverlay = get2Set(node.SystemOverlay.GetNode())
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
	controllerID := group.Cid.Get()
	groupID := group.Id.Get()

	if _, ok := self.Controllers[controllerID]; !ok {
		return errors.New("Controller does not exist: " + controllerID)
	}

	if _, ok := self.Controllers[controllerID].NodeGroups[groupID]; !ok {
		return errors.New("Group does not exist: " + groupID)
	}

	self.Controllers[controllerID].NodeGroups[groupID].DomainName = group.DomainName.Get()
	self.Controllers[controllerID].NodeGroups[groupID].Vnfs = group.Vnfs.Get()
	self.Controllers[controllerID].NodeGroups[groupID].KernelVersion = group.KernelVersion.Get()
	self.Controllers[controllerID].NodeGroups[groupID].KernelArgs = group.KernelArgs.Get()
	self.Controllers[controllerID].NodeGroups[groupID].Ipxe = group.Ipxe.Get()
	self.Controllers[controllerID].NodeGroups[groupID].IpmiNetmask = group.IpmiNetmask.Get()
	self.Controllers[controllerID].NodeGroups[groupID].IpmiUserName = group.IpmiUserName.Get()
	self.Controllers[controllerID].NodeGroups[groupID].IpmiPassword = group.IpmiPassword.Get()
	self.Controllers[controllerID].NodeGroups[groupID].RuntimeOverlay = group.RuntimeOverlay.Get()
	self.Controllers[controllerID].NodeGroups[groupID].SystemOverlay = group.SystemOverlay.Get()
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

	self.Controllers[controllerID].Services.Warewulfd.Port = get2Set(controller.Services.Warewulfd.Port)
	self.Controllers[controllerID].Services.Warewulfd.Secure = controller.Services.Warewulfd.Secure
	self.Controllers[controllerID].Services.Warewulfd.StartCmd = get2Set(controller.Services.Warewulfd.StartCmd)
	self.Controllers[controllerID].Services.Warewulfd.RestartCmd = get2Set(controller.Services.Warewulfd.RestartCmd)
	self.Controllers[controllerID].Services.Warewulfd.EnableCmd = get2Set(controller.Services.Warewulfd.EnableCmd)

	self.Controllers[controllerID].Services.Dhcp.Enabled = controller.Services.Dhcp.Enabled
	self.Controllers[controllerID].Services.Dhcp.ConfigFile = get2Set(controller.Services.Dhcp.ConfigFile)
	self.Controllers[controllerID].Services.Dhcp.RangeStart = get2Set(controller.Services.Dhcp.RangeStart)
	self.Controllers[controllerID].Services.Dhcp.RangeEnd = get2Set(controller.Services.Dhcp.RangeEnd)
	self.Controllers[controllerID].Services.Dhcp.StartCmd = get2Set(controller.Services.Dhcp.StartCmd)
	self.Controllers[controllerID].Services.Dhcp.RestartCmd = get2Set(controller.Services.Dhcp.RestartCmd)
	self.Controllers[controllerID].Services.Dhcp.EnableCmd = get2Set(controller.Services.Dhcp.EnableCmd)

	self.Controllers[controllerID].Services.Nfs.Enabled = controller.Services.Nfs.Enabled
	self.Controllers[controllerID].Services.Nfs.Exports = controller.Services.Nfs.Exports
	self.Controllers[controllerID].Services.Nfs.StartCmd = get2Set(controller.Services.Nfs.StartCmd)
	self.Controllers[controllerID].Services.Nfs.RestartCmd = get2Set(controller.Services.Nfs.RestartCmd)
	self.Controllers[controllerID].Services.Nfs.EnableCmd = get2Set(controller.Services.Nfs.EnableCmd)

	self.Controllers[controllerID].Services.Tftp.Enabled = controller.Services.Tftp.Enabled
	self.Controllers[controllerID].Services.Tftp.TftpRoot = get2Set(controller.Services.Tftp.TftpRoot)
	self.Controllers[controllerID].Services.Tftp.StartCmd = get2Set(controller.Services.Tftp.StartCmd)
	self.Controllers[controllerID].Services.Tftp.RestartCmd = get2Set(controller.Services.Tftp.RestartCmd)
	self.Controllers[controllerID].Services.Tftp.EnableCmd = get2Set(controller.Services.Tftp.EnableCmd)

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

	self.NodeProfiles[profileID].DomainName = get2Set(profile.DomainName)
	self.NodeProfiles[profileID].Vnfs = get2Set(profile.Vnfs)
	self.NodeProfiles[profileID].Ipxe = get2Set(profile.Ipxe)
	self.NodeProfiles[profileID].KernelVersion = get2Set(profile.KernelVersion)
	self.NodeProfiles[profileID].IpmiNetmask = get2Set(profile.IpmiNetmask)
	self.NodeProfiles[profileID].IpmiUserName = get2Set(profile.IpmiUserName)
	self.NodeProfiles[profileID].IpmiPassword = get2Set(profile.IpmiPassword)
	self.NodeProfiles[profileID].RuntimeOverlay = get2Set(profile.RuntimeOverlay)
	self.NodeProfiles[profileID].SystemOverlay = get2Set(profile.SystemOverlay)

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
