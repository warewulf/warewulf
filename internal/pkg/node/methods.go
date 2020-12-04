package node

import (
	"strings"
)

/**********
 *
 * Sets
 *
 *********/

func (self *Entry) Set(val string) {
	if val == "" {
		return
	}

	if strings.ToUpper(val) == "DELETE" {
		self.value = ""
	} else {
		self.value = val
	}

	return
}

func (self *Entry) SetB(val bool) {
	self.bool = val
	return
}

func (self *Entry) SetAlt(val string, from string) {
	if val == "" {
		return
	}

	self.altvalue = val
	self.from = from

	return
}

func (self *Entry) SetAltB(val bool, from string) {
	self.altbool = val
	self.from = from
	return
}

func (self *Entry) SetDefault(val string) {
	if val == "" {
		return
	}

	self.def = val

	return
}

/**********
 *
 * Gets
 *
 *********/

func (self *Entry) Get() string {
	if self.value != "" {
		return self.value
	}
	if self.altvalue != "" {
		return self.altvalue
	}
	if self.def != "" {
		return self.def
	}
	return ""
}

func (self *Entry) GetB() bool {
	return self.bool
}

func (self *Entry) GetReal() string {
	return self.value
}

func (self *Entry) GetRealB() bool {
	return self.bool
}

/**********
 *
 * Misc
 *
 *********/

func (self *Entry) Print() string {
	if self.value != "" {
		return self.value
	}
	if self.altvalue != "" {
		return self.altvalue
	}
	if self.def != "" {
		return self.def
	}
	return "--"
}

func (self *Entry) Source() string {
	if self.value != "" && self.altvalue != "" {
		return "SUPERSEDED"
		//return fmt.Sprintf("[%s]", self.from)
	} else if self.from == "" {
		return "--"
	}
	return self.from
}

func (self *Entry) Defined() bool {
	if self.value != "" {
		return true
	}
	if self.altvalue != "" {
		return true
	}
	if self.def != "" {
		return true
	}
	return false
}
