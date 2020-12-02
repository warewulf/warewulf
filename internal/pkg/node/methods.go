package node

import (
	"strings"
)

func (self *Entry) Print() string {
	if self.Node != "" {
		return self.Node
	}
	if self.Group != "" {
		return self.Group
	}
	if self.Profile != "" {
		return self.Profile
	}
	if self.Controller != "" {
		return self.Controller
	}
	if self.Default != "" {
		return self.Default
	}
	return "--"
}

func (self *Entry) Source() string {
	if self.Node != "" {
		return "node"
	}
	if self.Group != "" {
		return "group"
	}
	if self.Profile != "" {
		return "profile"
	}
	if self.Controller != "" {
		return "controller"
	}
	if self.Default != "" {
		return "default"
	}
	return ""
}

func (self *Entry) Get() string {
	if self.Node != "" {
		return self.Node
	}
	if self.Group != "" {
		return self.Group
	}
	if self.Profile != "" {
		return self.Profile
	}
	if self.Controller != "" {
		return self.Controller
	}
	if self.Default != "" {
		return self.Default
	}
	return ""
}

func (self *Entry) Defined() bool {
	if self.Get() == "" {
		return false
	}

	return true
}

func (self *Entry) SetDefault(value string) {
	if value == "" {
		return
	} else if strings.ToUpper(value) == "UNDEF" {
		value = ""
	}
	self.Default = value
}

func (self *Entry) SetGroup(value string) {
	if value == "" {
		return
	} else if strings.ToUpper(value) == "UNDEF" {
		value = ""
	}
	self.Group = value
}

func (self *Entry) SetProfile(value string) {
	if value == "" {
		return
	} else if strings.ToUpper(value) == "UNDEF" {
		value = ""
	}
	self.Profile = value
}

func (self *Entry) SetController(value string) {
	if value == "" {
		return
	} else if strings.ToUpper(value) == "UNDEF" {
		value = ""
	}
	self.Controller = value
}

func (self *Entry) SetNode(value string) {
	if value == "" {
		return
	} else if strings.ToUpper(value) == "UNDEF" {
		value = ""
	}
	self.Node = value
}

func (self *Entry) GetNode() string {
	return self.Node
}

func (self *Entry) GetGroup() string {
	return self.Group
}

func (self *Entry) GetController() string {
	return self.Controller
}

func (self *Entry) GetProfile() string {
	return self.Profile
}

func (self *Entry) GetDefault() string {
	return self.Default
}
