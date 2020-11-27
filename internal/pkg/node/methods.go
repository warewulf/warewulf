package node

func (self *NodeInfoEntry) String() string {
	if self.value != "" {
		return "node=" + self.value
	}
	if self.group != "" {
		return "group=" + self.group
	}
	if self.profile != "" {
		return "profile=" + self.profile
	}
	if self.def != "" {
		return "default=" + self.def
	}
	return "--"
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

func (self *NodeInfoEntry) Set(value string) {
	if value == "" {
		return
	}
	self.value = value
}

func (self *NodeInfoEntry) Unset() {
	self.value = ""
}

func (self *NodeInfoEntry) GetReal() string {
	return self.value
}