package node

import (
	"github.com/hpcng/warewulf/internal/pkg/errors"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"

	"gopkg.in/yaml.v2"
	"os"
	"strings"
)

func (self *nodeYaml) AddGroup(groupID string) error {
	var group nodeGroup

	if _, ok := self.NodeGroups[groupID]; ok {
		return errors.New("Group name already exists: " + groupID)
	}

	self.NodeGroups[groupID] = &group
	self.NodeGroups[groupID].DomainSuffix = groupID

	return nil
}

func (self *nodeYaml) AddNode(groupID string, nodeID string) error {
	var node nodeEntry

	wwlog.Printf(wwlog.VERBOSE, "Adding new node: %s/%s\n", groupID, nodeID)

	if _, ok := self.NodeGroups[groupID]; ok {
		if _, ok := self.NodeGroups[groupID].Nodes[nodeID]; ok {
			return errors.New("Nodename already exists in group: " + nodeID)
		}
	} else {
		return errors.New("Group does not exist: "+groupID)
	}

	self.NodeGroups[groupID].Nodes[nodeID] = &node
	self.NodeGroups[groupID].Nodes[nodeID].Hostname = nodeID

	return nil
}

func (self *nodeYaml) DelNode(groupID string, nodeID string) error {

	if _, ok := self.NodeGroups[groupID]; ok {
		if _, ok := self.NodeGroups[groupID].Nodes[nodeID]; ok {
			delete(self.NodeGroups[groupID].Nodes, nodeID)
			wwlog.Printf(wwlog.VERBOSE, "Deleting node: %s/%s\n", groupID, nodeID)
		} else {
			return errors.New("Node '"+nodeID+"' was not found in group '"+groupID+"'")
		}
	} else {
		return errors.New("Group '"+groupID+"' was not found")
	}

	return nil
}

func (self *nodeYaml) DelNodeNet(groupID string, nodeID string, netDev string) error {

	if _, ok := self.NodeGroups[groupID]; ok {
		if _, ok := self.NodeGroups[groupID].Nodes[nodeID]; ok {
			if _, ok := self.NodeGroups[groupID].Nodes[nodeID].NetDevs[netDev]; ok {
				wwlog.Printf(wwlog.VERBOSE, "Deleting node network device: %s/%s:%s\n", groupID, nodeID, netDev)
				delete(self.NodeGroups[groupID].Nodes[nodeID].NetDevs, netDev)
			} else {
				return errors.New("Network device '"+netDev+"' was not found in node '"+groupID+"/"+nodeID+"'")
			}
		} else {
			return errors.New("Node '"+nodeID+"' was not found in group '"+groupID+"'")
		}
	} else {
		return errors.New("Group '"+groupID+"' was not found")
	}

	return nil
}


func (self *nodeYaml) DelGroup(groupname string) error {
	if _, ok := self.NodeGroups[groupname]; ok {
		delete(self.NodeGroups, groupname)
	} else {
		return errors.New("Node group undefined: " + groupname)
	}

	return nil
}

func (self *nodeYaml) SetGroupVal(groupID string, entry string, value string) error {
	if strings.ToUpper(value) == "UNDEF" || strings.ToUpper(value) == "NIL" || strings.ToUpper(value) == "DEL" {
		value = ""
	}

	if _, ok := self.NodeGroups[groupID]; ok {
		wwlog.Printf(wwlog.VERBOSE, "Setting group %s to: %s = '%s'\n", groupID, entry, value )

		switch strings.ToUpper(entry) {
		case "DOMAINSUFFIX":
			util.ValidateOrDie("Domain", entry, "^[a-zA-Z0-9-._]*$")
			self.NodeGroups[groupID].DomainSuffix = value
		}

	} else {
		return errors.New("Group does not exist: " +groupID)
	}

	return nil
}

func (self *nodeYaml) SetNodeVal(groupID string, nodeID string, entry string, value string) error {
	if strings.ToUpper(value) == "UNDEF" || strings.ToUpper(value) == "NIL" || strings.ToUpper(value) == "DEL" {
		value = ""
	}
	if _, ok := self.NodeGroups[groupID]; ok {
		if _, ok := self.NodeGroups[groupID].Nodes[nodeID]; ok {
			wwlog.Printf(wwlog.VERBOSE, "Setting node %s/%s to: %s = '%s'\n", groupID, nodeID, entry, value )

			switch strings.ToUpper(entry) {
			case "VNFS":
				util.ValidateOrDie("VNFS", entry, "^[a-zA-Z0-9-._]*$")
				self.NodeGroups[groupID].Nodes[nodeID].Vnfs = value
			case "KERNEL":
				util.ValidateOrDie("Kernel Version", entry, "^[a-zA-Z0-9-._]*$")
				self.NodeGroups[groupID].Nodes[nodeID].KernelVersion = value
			case "DOMAINSUFFIX":
				util.ValidateOrDie("Domain", entry, "^[a-zA-Z0-9-._]*$")
				self.NodeGroups[groupID].Nodes[nodeID].DomainSuffix = value
			case "IPXE":
				util.ValidateOrDie("iPXE Template", entry, "^[a-zA-Z0-9-._]*$")
				self.NodeGroups[groupID].Nodes[nodeID].Ipxe = value
			case "HOSTNAME":
				util.ValidateOrDie("Hostname", entry, "^[a-zA-Z0-9-._]*$")
				self.NodeGroups[groupID].Nodes[nodeID].Hostname = value
			case "IPMIIPADDR":
				util.ValidateOrDie("IPMI IP Address", entry, "^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}$")
				self.NodeGroups[groupID].Nodes[nodeID].IpmiIpaddr = value
			case "IPMIUSERNAME":
				util.ValidateOrDie("IPMI Username", entry, "^[a-zA-Z0-9-._]*$")
				self.NodeGroups[groupID].Nodes[nodeID].IpmiUserName = value
			case "IPMIPASSWORD":
				util.ValidateOrDie("IPMI Password", entry, "^[a-zA-Z0-9-._]*$")
				self.NodeGroups[groupID].Nodes[nodeID].IpmiPassword = value

			}
		} else {
			return errors.New("Node does not exist: " +groupID+ "/" +nodeID)
		}
	} else {
		return errors.New("Group does not exist: " +groupID)
	}

	return nil
}

func (self *nodeYaml) SetNodeNet(groupID string, nodeID string, netDev string, entry string, value string) error {
	if strings.ToUpper(value) == "UNDEF" || strings.ToUpper(value) == "NIL" || strings.ToUpper(value) == "DEL" {
		value = ""
	}

	if _, ok := self.NodeGroups[groupID]; ok {
		if _, ok := self.NodeGroups[groupID].Nodes[nodeID]; ok {
			if _, ok := self.NodeGroups[groupID].Nodes[nodeID].NetDevs[netDev]; ok {
				wwlog.Printf(wwlog.VERBOSE, "Editing existing node NetDev entry for node: %s/%s\n", groupID, nodeID)
			} else {
				var nd NetDevs
				self.NodeGroups[groupID].Nodes[nodeID].NetDevs[netDev] = &nd
			}
		} else {
			return errors.New("Node does not exist: "+groupID+"/"+nodeID)
		}
	} else {
		return errors.New("Group does not exist: "+groupID)
	}

	switch strings.ToUpper(entry) {
	case "IPADDR":
		util.ValidateOrDie("IP address", value, "^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}$")
		self.NodeGroups[groupID].Nodes[nodeID].NetDevs[netDev].Ipaddr = value
	case "NETMASK":
		util.ValidateOrDie("Netmask", value, "^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}$")
		self.NodeGroups[groupID].Nodes[nodeID].NetDevs[netDev].Netmask = value
	case "GATEWAY":
		util.ValidateOrDie("Gateway", value, "^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}$")
		self.NodeGroups[groupID].Nodes[nodeID].NetDevs[netDev].Gateway = value
	case "TYPE":
		util.ValidateOrDie("Network device type", value, "^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}$")
		self.NodeGroups[groupID].Nodes[nodeID].NetDevs[netDev].Type = value
	case "HWADDR":
		util.ValidateOrDie("HW address", value, "^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$")
		self.NodeGroups[groupID].Nodes[nodeID].NetDevs[netDev].Hwaddr = value

	}

	return nil
}

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

