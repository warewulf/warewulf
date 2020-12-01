package node

import (
	"strings"
)

func (self *NodeInfoEntry) String() string {
	if self.value != "" {
		return self.value
	}
	if self.group != "" {
		return self.group
	}
	if self.profile != "" {
		return self.profile
	}
	if self.controller != "" {
		return self.controller
	}
	if self.def != "" {
		return self.def
	}
	return "--"
}

func (self *NodeInfoEntry) Source() string {
	if self.value != "" {
		return "node"
	}
	if self.group != "" {
		return "group"
	}
	if self.profile != "" {
		return "profile"
	}
	if self.controller != "" {
		return "controller"
	}
	if self.def != "" {
		return "default"
	}
	return ""
}

func (self *NodeInfoEntry) Get() string {
	if self.value != "" {
		return self.value
	}
	if self.group != "" {
		return self.group
	}
	if self.profile != "" {
		return self.profile
	}
	if self.controller != "" {
		return self.controller
	}
	if self.def != "" {
		return self.def
	}
	return ""
}

func (self *NodeInfoEntry) Defined() bool {
	if self.Get() == "" {
		return false
	}

	return true
}

func (self *NodeInfoEntry) SetDefault(value string) {
	if value == "" {
		return
	}
	self.def = value
}

func (self *NodeInfoEntry) SetGroup(value string) {
	if value == "" {
		return
	}
	self.group = value
}

func (self *NodeInfoEntry) SetProfile(value string) {
	if value == "" {
		return
	}
	self.profile = value
}

func (self *NodeInfoEntry) SetController(value string) {
	if value == "" {
		return
	}
	self.controller = value
}

func (self *NodeInfoEntry) Set(value string) {
	if value == "" {
		return
	} else if strings.ToUpper(value) == "UNDEF" {
		value = ""
	}
	self.value = value
}

func (self *NodeInfoEntry) Unset() {
	self.value = ""
}

func (self *NodeInfoEntry) GetReal() string {
	return self.value
}