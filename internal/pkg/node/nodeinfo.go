package node

import (
	// "fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

/*
NodeInfo is the in memory datastructure, which can containe
a default value, which is overwritten by the overlay from the
overlay (altvalue) which is overwitten by the value of the
node itself, for all values of type Entry.
*/
type NodeInfo struct {
	Id             Entry
	Comment        Entry
	ClusterName    Entry
	ContainerName  Entry
	Ipxe           Entry
	RuntimeOverlay Entry
	SystemOverlay  Entry
	Root           Entry
	Discoverable   Entry
	Init           Entry // TODO Finish adding this.
	AssetKey       Entry
	Kernel         *KernelEntry
	Ipmi           *IpmiEntry
	Profiles       Entry
	PrimaryNetDev  Entry
	NetDevs        map[string]*NetDevEntry
	Tags           map[string]*Entry
}


/******
 * Internal code data representations
 ******/
/*
Holds string values, when accessed via Get, its value
is returned which is the default or if set the value
from the profile or if set the value of the node itself
*/
type Entry struct {
	value    []string
	altvalue []string
	from     string
	def      []string
}


type IpmiEntry struct {
	Ipaddr    Entry
	Netmask   Entry
	Port      Entry
	Gateway   Entry
	UserName  Entry
	Password  Entry
	Interface Entry
	Write     Entry
	Tags      map[string]*Entry
}


type KernelEntry struct {
	Override Entry
	Args     Entry
}


type NetDevEntry struct {
	Type    Entry
	OnBoot  Entry
	Device  Entry
	Hwaddr  Entry
	Ipaddr  Entry
	Ipaddr6 Entry
	IpCIDR  Entry
	Prefix  Entry
	Netmask Entry
	Gateway Entry
	MTU     Entry
	Primary Entry
	Tags    map[string]*Entry
}


/*
Filter a given slice of NodeInfo against a given
regular expression
*/
func FilterByName(set []NodeInfo, searchList []string) []NodeInfo {
	var ret []NodeInfo
	unique := make(map[string]NodeInfo)

	if len(searchList) > 0 {
		for _, search := range searchList {
			for _, entry := range set {
				if match, _ := regexp.MatchString("^"+search+"$", entry.Id.Get()); match {
					unique[entry.Id.Get()] = entry
				}
			}
		}
		for _, n := range unique {
			ret = append(ret, n)
		}
	} else {
		ret = set
	}

	return ret
}


/*
Set value. If argument is 'UNDEF', 'DELETE',
'UNSET" or '--' the value is removed.
N.B. the '--' might never ever happen as '--'
is parsed out by cobra
*/
func (ent *Entry) Set(val string) {
	if val == "" {
		return
	}

	if val == "UNDEF" || val == "DELETE" || val == "UNSET" || val == "--" || val == "nil" {
		wwlog.Debug("Removing value for %v", *ent)
		ent.value = []string{""}
	} else {
		ent.value = []string{val}
	}
}

/*
Set bool
*/
func (ent *Entry) SetB(val bool) {
	if val {
		ent.value = []string{"true"}
	} else {
		ent.value = []string{"false"}
	}
}

func (ent *Entry) SetSlice(val []string) {
	if len(val) == 0 {
		return
	} else if len(val) == 1 && val[0] == "" { // check also for an "empty" slice
		return
	}
	if val[0] == "UNDEF" || val[0] == "DELETE" || val[0] == "UNSET" || val[0] == "--" {
		ent.value = []string{}
	} else {
		ent.value = val
	}
}

/*
Set alternative value
*/
func (ent *Entry) SetAlt(val string, from string) {
	if val == "" {
		return
	}
	ent.altvalue = []string{val}
	ent.from = from
}

// /*
// Sets alternative bool
// */
// func (ent *Entry) SetAltB(val bool, from string) {
// 	if val {
// 		ent.altvalue = []string{"true"}
// 		ent.from = from
// 	} else {
// 		ent.altvalue = []string{"false"}
// 		ent.from = from
// 	}
// }

/*
Sets alternative slice
*/
func (ent *Entry) SetAltSlice(val []string, from string) {
	if len(val) == 0 {
		return
	}
	ent.altvalue = val
	ent.from = from
}

/*
Sets the default value of an entry.
*/
func (ent *Entry) SetDefault(val string) {
	if val == "" {
		return
	}
	ent.def = []string{val}

}

/*
Set the default entry as slice
*/
func (ent *Entry) SetDefaultSlice(val []string) {
	if len(val) == 0 {
		return
	}
	ent.def = val

}

/*
Set default etry as bool
*/
func (ent *Entry) SetDefaultB(val bool) {
	if val {
		ent.def = []string{"true"}
	} else {
		ent.def = []string{"false"}
	}
}

/*
Remove a element from a slice
*/
func (ent *Entry) SliceRemoveElement(val string) {
	util.SliceRemoveElement(ent.value, val)
}


/*
Gets the the entry of the value in folowing order
* node value if set
* profile value if set
* default value if set
*/
func (ent *Entry) Get() string {
	if len(ent.value) != 0 {
		return ent.value[0]
	}
	if len(ent.altvalue) != 0 {
		return ent.altvalue[0]
	}
	if len(ent.def) != 0 {
		return ent.def[0]
	}
	return ""
}

/*
Get the bool value of an entry.
*/
func (ent *Entry) GetB() bool {
	if len(ent.value) > 0 {
		return !(strings.ToLower(ent.value[0]) == "false" ||
			strings.ToLower(ent.value[0]) == "no" ||
			ent.value[0] == "0")
	} else if len(ent.altvalue) > 0 {
		return !(strings.ToLower(ent.altvalue[0]) == "false" ||
			strings.ToLower(ent.altvalue[0]) == "no" ||
			ent.altvalue[0] == "0")
	} else {
		return !(len(ent.def) == 0 ||
			strings.ToLower(ent.def[0]) == "false" ||
			strings.ToLower(ent.def[0]) == "no" ||
			ent.def[0] == "0")
	}
}

/*
Returns a string slice created from a comma seperated list of the value.
*/
func (ent *Entry) GetSlice() []string {
	var retval []string
	if len(ent.value) != 0 {
		return ent.value
	}
	if len(ent.altvalue) != 0 {
		return ent.altvalue
	}
	if len(ent.def) != 0 {
		return ent.def
	}
	return retval
}

/*
Get the real value, not the alternative of default one.
*/
func (ent *Entry) GetReal() string {
	if len(ent.value) == 0 {
		return ""
	}
	return ent.value[0]
}

/*
Get the real value, not the alternative of default one.
*/
func (ent *Entry) GetRealSlice() []string {
	if len(ent.value) == 0 {
		return []string{}
	}
	return ent.value
}

// /*
// true if the entry has set a real value, else false.
// */
// func (ent *Entry) GotReal() bool {
// 	return len(ent.value) != 0
// }


/*
Returns the value of Entry if it was defined set or
alternative is presend. Default value is in '()'. If
nothing is defined '--' is returned.
*/
func (ent *Entry) Print() string {
	if len(ent.value) != 0 {
		return strings.Join(ent.value, ",")
	}
	if len(ent.altvalue) != 0 {
		return strings.Join(ent.altvalue, ",")
	}
	if len(ent.def) != 0 {
		return "(" + strings.Join(ent.def, ",") + ")"
	}
	return "--"
}

// /*
// same as GetB()
// */
// func (ent *Entry) PrintB() string {
// 	if len(ent.value) != 0 || len(ent.altvalue) != 0 {
// 		return fmt.Sprintf("%t", ent.GetB())
// 	}
// 	return fmt.Sprintf("(%t)", ent.GetB())
// }

/*
Returns SUPERSEDED if value was set per node or
per profile. Else -- is returned.
*/
func (ent *Entry) Source() string {
	if len(ent.value) != 0 && len(ent.altvalue) != 0 {
		return "SUPERSEDED"
		//return fmt.Sprintf("[%s]", ent.from)
	} else if ent.from == "" {
		return "--"
	}
	return ent.from
}

/*
Check if value was defined.
*/
func (ent *Entry) Defined() bool {
	if len(ent.value) != 0 {
		return true
	}
	if len(ent.altvalue) != 0 {
		return true
	}
	if len(ent.def) != 0 {
		return true
	}
	return false
}


/*
Create an empty node NodeInfo
*/
func NewInfo() (nodeInfo NodeInfo) {
	nodeInfo.Ipmi = new(IpmiEntry)
	nodeInfo.Kernel = new(KernelEntry)
	nodeInfo.NetDevs = make(map[string]*NetDevEntry)
	return nodeInfo
}


/*
Populates all fields of NodeInfo with Set from the
values of NodeConf.
*/
func (node *NodeInfo) SetFrom(n *NodeConf) {
	setWrap := func(entr *Entry, val string, nameArg string) {
		entr.Set(val)
	}
	setSliceWrap := func(entr *Entry, val []string, nameArg string) {
		entr.SetSlice(val)
	}
	node.setterFrom(n, "", setWrap, setSliceWrap)
}


/*
Populates all fields of NodeInfo with SetAlt from the
values of NodeConf. The string profileName is used to
destermine from which source/NodeInfo the entry came
from.
*/
func (node *NodeInfo) SetAltFrom(n *NodeConf, profileName string) {
	node.setterFrom(n, profileName, (*Entry).SetAlt, (*Entry).SetAltSlice)
}


/*
Populates all fields of NodeInfo with SetDefault from the
values of NodeConf.
*/
func (node *NodeInfo) SetDefFrom(n *NodeConf) {
	setWrap := func(entr *Entry, val string, nameArg string) {
		entr.SetDefault(val)
	}
	setSliceWrap := func(entr *Entry, val []string, nameArg string) {
		entr.SetDefaultSlice(val)
	}
	node.setterFrom(n, "", setWrap, setSliceWrap)
}


/*
Abstract function which populates a NodeInfo from a NodeConf via
setter functionns.
*/
func (node *NodeInfo) setterFrom(n *NodeConf, nameArg string,
	setter func(*Entry, string, string),
	setterSlice func(*Entry, []string, string)) {
	// get the full memory, taking the shortcut and init Ipmi and Kernel directly
	if node.Kernel == nil {
		node.Kernel = new(KernelEntry)
	}
	if node.Ipmi == nil {
		node.Ipmi = new(IpmiEntry)
	}
	// also n could be nil
	if n == nil {
		myn := NewConf()
		n = &myn
	}
	nodeInfoVal := reflect.ValueOf(node)
	nodeInfoType := reflect.TypeOf(node)
	nodeConfVal := reflect.ValueOf(n)
	// now iterate of every field
	for i := 0; i < nodeInfoType.Elem().NumField(); i++ {
		valField := nodeConfVal.Elem().FieldByName(nodeInfoType.Elem().Field(i).Name)
		if valField.IsValid() {
			// found field with same name for Conf and Info
			if nodeInfoType.Elem().Field(i).Type == reflect.TypeOf(Entry{}) {
				if valField.Type().Kind() == reflect.String {
					setter(nodeInfoVal.Elem().Field(i).Addr().Interface().(*Entry), valField.String(), nameArg)
				} else if valField.Type() == reflect.TypeOf([]string{}) {
					setterSlice(nodeInfoVal.Elem().Field(i).Addr().Interface().(*Entry), valField.Interface().([]string), nameArg)
				}
			} else if nodeInfoType.Elem().Field(i).Type.Kind() == reflect.Ptr && !valField.IsZero() {
				nestedInfoType := reflect.TypeOf(nodeInfoVal.Elem().Field(i).Interface())
				netstedInfoVal := reflect.ValueOf(nodeInfoVal.Elem().Field(i).Interface())
				nestedConfVal := reflect.ValueOf(valField.Interface())
				for j := 0; j < nestedInfoType.Elem().NumField(); j++ {
					nestedVal := nestedConfVal.Elem().FieldByName(nestedInfoType.Elem().Field(j).Name)
					if nestedVal.IsValid() {
						if netstedInfoVal.Elem().Field(j).Type() == reflect.TypeOf(Entry{}) {
							setter(netstedInfoVal.Elem().Field(j).Addr().Interface().(*Entry), nestedVal.String(), nameArg)
						} else {
							confMap := nestedVal.Interface().(map[string]string)
							if netstedInfoVal.Elem().Field(j).IsNil() {
								newMap := make(map[string]*Entry)
								mapPtr := (netstedInfoVal.Elem().Field(j).Addr().Interface()).(*map[string](*Entry))
								*mapPtr = newMap
							}
							for key, val := range confMap {
								entr := new(Entry)
								setter(entr, val, nameArg)
								(netstedInfoVal.Elem().Field(j).Interface()).(map[string](*Entry))[key] = entr
							}
						}
					}
				}
			} else if nodeInfoType.Elem().Field(i).Type == reflect.TypeOf(map[string](*Entry)(nil)) {
				confMap := valField.Interface().(map[string]string)
				for key, val := range confMap {
					entr := new(Entry)
					setter(entr, val, nameArg)
					(nodeInfoVal.Elem().Field(i).Interface()).(map[string](*Entry))[key] = entr
				}
			} else if nodeInfoType.Elem().Field(i).Type == reflect.TypeOf(map[string](*NetDevEntry)(nil)) {
				netValMap := valField.Interface().(map[string](*NetDevConf))
				for netName, netVals := range netValMap {
					netValsType := reflect.ValueOf(netVals)
					netMap := nodeInfoVal.Elem().Field(i).Interface().(map[string](*NetDevEntry))
					if nodeInfoVal.Elem().Field(i).IsNil() {
						netMap = make(map[string]*NetDevEntry)
					}
					if _, ok := netMap[netName]; !ok {
						var newNet NetDevEntry
						newNet.Tags = make(map[string]*Entry)
						netMap[netName] = &newNet
					}
					netInfoType := reflect.TypeOf(*netMap[netName])
					netInfoVal := reflect.ValueOf(netMap[netName])
					for j := 0; j < netInfoType.NumField(); j++ {
						netVal := netValsType.Elem().FieldByName(netInfoType.Field(j).Name)
						if netVal.IsValid() {
							if netVal.Type().Kind() == reflect.String {
								setter(netInfoVal.Elem().Field(j).Addr().Interface().((*Entry)), netVal.String(), nameArg)
							} else if netVal.Type() == reflect.TypeOf(map[string]string{}) {
								for key, val := range (netVal.Interface()).(map[string]string) {
									//netTagMap := netInfoVal.Elem().Field(j).Interface().((map[string](*Entry)))
									if _, ok := netInfoVal.Elem().Field(j).Interface().((map[string](*Entry)))[key]; !ok {
										netInfoVal.Elem().Field(j).Interface().((map[string](*Entry)))[key] = new(Entry)
									}
									setter(netInfoVal.Elem().Field(j).Interface().((map[string](*Entry)))[key], val, nameArg)
								}
							}
						}
					}
				}
			}
		}
	}
}


// /*
// Populates all fields of NetDevEntry with Set from the
// values of NetDevs.
// Actually not used, just for completeness.
// */
// func (netDev *NetDevEntry) SetFrom(netYaml *NetDevConf) {
// 	setWrap := func(entr *Entry, val string, nameArg string) {
// 		entr.Set(val)
// 	}
// 	setSliceWrap := func(entr *Entry, val []string, nameArg string) {
// 		entr.SetSlice(val)
// 	}
// 	netDev.setterFrom(netYaml, "", setWrap, setSliceWrap)
// }

// /*
// Populates all fields of NetDevEntry with SetAlt from the
// values of NetDevs. The string profileName is used to
// destermine from which source/NodeInfo the entry came
// from.
// Actually not used, just for completeness.
// */
// func (netDev *NetDevEntry) SetAltFrom(netYaml *NetDevConf, profileName string) {
// 	netDev.setterFrom(netYaml, profileName, (*Entry).SetAlt, (*Entry).SetAltSlice)
// }

/*
Populates all fields of NodeInfo with SetDefault from the
values of NodeConf.
*/
func (netDev *NetDevEntry) SetDefFrom(netYaml *NetDevConf) {
	setWrap := func(entr *Entry, val string, nameArg string) {
		entr.SetDefault(val)
	}
	setSliceWrap := func(entr *Entry, val []string, nameArg string) {
		entr.SetDefaultSlice(val)
	}
	netDev.setterFrom(netYaml, "", setWrap, setSliceWrap)
}

/*
Abstract function for setting a NetDevEntry from a NetDevs
*/
func (netDev *NetDevEntry) setterFrom(netYaml *NetDevConf, nameArg string,
	setter func(*Entry, string, string),
	setterSlice func(*Entry, []string, string)) {
	// check if netYaml is empty
	if netYaml == nil {
		netYaml = new(NetDevConf)
	}
	netValues := reflect.ValueOf(netDev)
	netInfoType := reflect.TypeOf(*netYaml)
	netInfoVal := reflect.ValueOf(*netYaml)
	for j := 0; j < netInfoType.NumField(); j++ {
		netVal := netValues.Elem().FieldByName(netInfoType.Field(j).Name)
		if netVal.IsValid() {
			if netInfoVal.Field(j).Type().Kind() == reflect.String {
				setter(netVal.Addr().Interface().((*Entry)), netInfoVal.Field(j).String(), nameArg)
			} else if netVal.Type() == reflect.TypeOf(map[string]string{}) {
				// danger zone following code is not tested
				for key, val := range (netVal.Interface()).(map[string]string) {
					//netTagMap := netInfoVal.Elem().Field(j).Interface().((map[string](*Entry)))
					if _, ok := netInfoVal.Elem().Field(j).Interface().((map[string](*Entry)))[key]; !ok {
						netInfoVal.Elem().Field(j).Interface().((map[string](*Entry)))[key] = new(Entry)
					}
					setter(netInfoVal.Elem().Field(j).Interface().((map[string](*Entry)))[key], val, nameArg)
				}
			}
		}
	}
}
