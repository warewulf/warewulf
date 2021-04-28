package node

import "regexp"

/**********
 *
 * Filters
 *
 *********/

func FilterByName(set []NodeInfo, searchList []string) []NodeInfo {
	var ret []NodeInfo

	if len(searchList) > 0 {
		for _, search := range searchList {
			for _, entry := range set {
				b, _ := regexp.MatchString(search, entry.Id.Get())
				if b {
					ret = append(ret, entry)
				}
			}
		}
	} else {
		ret = set
	}

	return ret
}

/**********
 *
 * Sets
 *
 *********/

func (ent *Entry) Set(val string) {
	if val == "" {
		return
	}

	if val == "UNDEF" || val == "DELETE" || val == "UNSET" {
		ent.value = ""
	} else {
		ent.value = val
	}

}

func (ent *Entry) SetB(val bool) {
	ent.bool = val
}

func (ent *Entry) SetAlt(val string, from string) {
	if val == "" {
		return
	}

	ent.altvalue = val
	ent.from = from

}

func (ent *Entry) SetAltB(val bool, from string) {
	ent.altbool = val
	ent.from = from

}

func (ent *Entry) SetDefault(val string) {
	if val == "" {
		return
	}

	ent.def = val

}

/**********
 *
 * Gets
 *
 *********/

func (ent *Entry) Get() string {
	if ent.value != "" {
		return ent.value
	}
	if ent.altvalue != "" {
		return ent.altvalue
	}
	if ent.def != "" {
		return ent.def
	}
	return ""
}

func (ent *Entry) GetB() bool {
	return ent.bool
}

func (ent *Entry) GetReal() string {
	return ent.value
}

func (ent *Entry) GetRealB() bool {
	return ent.bool
}

/**********
 *
 * Misc
 *
 *********/

func (ent *Entry) Print() string {
	if ent.value != "" {
		return ent.value
	}
	if ent.altvalue != "" {
		return ent.altvalue
	}
	if ent.def != "" {
		return "(" + ent.def + ")"
	}
	return "--"
}

func (ent *Entry) PrintB() bool {
	if ent.from == "" {
		return ent.bool
	}
	return ent.altbool
}

func (ent *Entry) Source() string {
	if ent.value != "" && ent.altvalue != "" {
		return "SUPERSEDED"
		//return fmt.Sprintf("[%s]", ent.from)
	} else if ent.from == "" {
		return "--"
	}
	return ent.from
}

func (ent *Entry) Defined() bool {
	if ent.value != "" {
		return true
	}
	if ent.altvalue != "" {
		return true
	}
	if ent.def != "" {
		return true
	}
	return false
}
