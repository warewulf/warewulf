package node

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

/*
Populates a NodeConf struct (the one which goes to disk) from a
NodeInfo (which just lives in memory), with the values from all
the underlying entries using GetReal, so just the explicit values
go do disk.
*/
func (nodeConf *NodeConf) GetRealFrom(nodeInfo NodeInfo) {
	recursiveGetter(&nodeInfo, nodeConf, (*Entry).GetReal, (*Entry).GetRealSlice)
}

/*
Populates a NodeConf struct from a NodeInfo, with the combined
values from the underlying entries using Get.
*/
func (nodeConf *NodeConf) GetFrom(nodeInfo NodeInfo) {
	recursiveGetter(&nodeInfo, nodeConf, (*Entry).Get, (*Entry).GetSlice)
}

/*
Abstract function which populates a NodeConf from the given NodeInfo
via getter functions. Calls recursive itself for nested structures.
Panics if the NodeConf has fields which are not type of string,[]string,map[string]*ptr
*/
func recursiveGetter(
	source, target interface{},
	getter func(*Entry) string,
	getterSlice func(*Entry) []string) {
	sourceValue := reflect.ValueOf(source)
	targetType := reflect.TypeOf(target)
	targetValue := reflect.ValueOf(target)
	if targetValue.Elem().Kind() == reflect.Struct && sourceValue.Elem().Kind() == reflect.Struct {
		for i := 0; i < targetType.Elem().NumField(); i++ {
			sourceValueMatched := sourceValue.Elem().FieldByName(targetType.Elem().Field(i).Name)
			if sourceValueMatched.IsValid() {
				if sourceValueMatched.Type() == reflect.TypeOf(Entry{}) {
					// get the fields which are part of the struct
					switch targetValue.Elem().Field(i).Type() {
					case reflect.TypeOf(""):
						newValue := (targetValue.Elem().Field(i).Addr().Interface()).(*string)
						source := sourceValueMatched.Interface().(Entry)
						*newValue = getter(&source)
					case reflect.TypeOf([]string{}):
						newValue := (targetValue.Elem().Field(i).Addr().Interface()).(*[]string)
						source := sourceValueMatched.Interface().(Entry)
						*newValue = getterSlice(&source)
					default:
						panic(fmt.Errorf("can't convert an Entry to %s", targetValue.Elem().Field(i).Type()))
					}
				} else if sourceValueMatched.Kind() == reflect.Ptr {
					// if we get a pointer, initialize if empty and then have a recursive call
					if targetValue.Elem().Field(i).IsZero() {
						targetValue.Elem().Field(i).Set(reflect.New(targetType.Elem().Field(i).Type.Elem()))
					}
					recursiveGetter(sourceValueMatched.Interface(), targetValue.Elem().Field(i).Interface(), getter, getterSlice)
				} else if sourceValueMatched.Type().Kind() == reflect.Map {
					if targetValue.Elem().Field(i).IsZero() {
						targetValue.Elem().Field(i).Set(reflect.MakeMap(targetType.Elem().Field(i).Type))
					}
					sourceIter := sourceValueMatched.MapRange()
					if sourceValueMatched.Type() == reflect.TypeOf(map[string]*Entry{}) {
						// go over a simple map with strings
						for sourceIter.Next() {
							if !targetValue.Elem().Field(i).MapIndex(sourceIter.Key()).IsValid() {
								str := getter((sourceIter.Value().Interface()).(*Entry))
								targetValue.Elem().Field(i).SetMapIndex(sourceIter.Key(), reflect.ValueOf(str))
							}
						}
					} else {
						// now the complicated map which contains pointers to objects
						for sourceIter.Next() {
							if !targetValue.Elem().Field(i).MapIndex(sourceIter.Key()).IsValid() {
								newPtr := reflect.New(targetType.Elem().Field(i).Type.Elem().Elem())
								targetValue.Elem().Field(i).SetMapIndex(sourceIter.Key(), newPtr)
							}
							recursiveGetter(sourceIter.Value().Interface(), targetValue.Elem().Field(i).MapIndex(sourceIter.Key()).Interface(), getter, getterSlice)

						}
					}
				}
			}
		}
	}
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
				nestedInfoVal := reflect.ValueOf(nodeInfoVal.Elem().Field(i).Interface())
				nestedConfVal := reflect.ValueOf(valField.Interface())
				for j := 0; j < nestedInfoType.Elem().NumField(); j++ {
					nestedVal := nestedConfVal.Elem().FieldByName(nestedInfoType.Elem().Field(j).Name)
					if nestedVal.IsValid() {
						if nestedInfoVal.Elem().Field(j).Type() == reflect.TypeOf(Entry{}) {
							setter(nestedInfoVal.Elem().Field(j).Addr().Interface().(*Entry), nestedVal.String(), nameArg)
						} else if nestedInfoVal.Elem().Field(j).Type() == reflect.TypeOf(map[string](*Entry){}) {
							confMap := nestedVal.Interface().(map[string]string)
							if nestedInfoVal.Elem().Field(j).IsNil() {
								ptr := nestedInfoVal.Elem().Field(j).Addr().Interface().(*map[string](*Entry))
								*ptr = make(map[string]*Entry)
							}
							tagMap := nestedInfoVal.Elem().Field(j).Interface().(map[string](*Entry))
							for key, val := range confMap {
								if entr, ok := tagMap[key]; ok {
									setter(entr, val, nameArg)
								} else {
									entr := new(Entry)
									tagMap[key] = entr
									setter(entr, val, nameArg)
								}
							}
						}
					}
				}
			} else if nodeInfoType.Elem().Field(i).Type == reflect.TypeOf(map[string](*Entry)(nil)) {
				confMap := valField.Interface().(map[string]string)
				for key, val := range confMap {
					tagMap := nodeInfoVal.Elem().Field(i).Interface().(map[string](*Entry))
					if nodeInfoVal.Elem().Field(i).IsNil() {
						tagMap = make(map[string]*Entry)
					}
					if entr, ok := tagMap[key]; ok {
						setter(entr, val, nameArg)
					} else {
						entr := new(Entry)
						tagMap[key] = entr
						setter(entr, val, nameArg)
					}
				}
			} else if nodeInfoType.Elem().Field(i).Type == reflect.TypeOf(map[string](*NetDevEntry)(nil)) {
				netValMap := valField.Interface().(map[string](*NetDevs))
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

/*
Flattens out a NodeConf, which means if there are no explicit values in *IpmiConf
or *KernelConf, these pointer will set to nil. This will remove something like
ipmi: {} from nodes.conf
*/
func (info *NodeConf) Flatten() {
	recursiveFlatten(info)
}
func recursiveFlatten(strct interface{}) {
	confType := reflect.TypeOf(strct)
	confVal := reflect.ValueOf(strct)
	for j := 0; j < confType.Elem().NumField(); j++ {
		if confVal.Elem().Field(j).Type().Kind() == reflect.Ptr && !confVal.Elem().Field(j).IsNil() {
			// iterate now over the ptr fields
			setToNil := true
			nestedType := reflect.TypeOf(confVal.Elem().Field(j).Interface())
			nestedVal := reflect.ValueOf(confVal.Elem().Field(j).Interface())
			for i := 0; i < nestedType.Elem().NumField(); i++ {
				if nestedType.Elem().Field(i).Type.Kind() == reflect.String &&
					nestedVal.Elem().Field(i).Interface().(string) != "" {
					setToNil = false
				} else if nestedType.Elem().Field(i).Type == reflect.TypeOf([]string{}) &&
					len(nestedVal.Elem().Field(i).Interface().([]string)) != 0 {
					setToNil = false
				} else if nestedType.Elem().Field(i).Type == reflect.TypeOf(map[string]string{}) &&
					len(nestedVal.Elem().Field(i).Interface().(map[string]string)) != 0 {
					setToNil = false
				}
			}
			if setToNil {
				confVal.Elem().Field(j).Set(reflect.Zero(confVal.Elem().Field(j).Type()))
			}
		}
	}
}

/*
Populates all fields of NetDevEntry with Set from the
values of NetDevs.
Actually not used, just for completeness.
*/
func (netDev *NetDevEntry) SetFrom(netYaml *NetDevs) {
	setWrap := func(entr *Entry, val string, nameArg string) {
		entr.Set(val)
	}
	setSliceWrap := func(entr *Entry, val []string, nameArg string) {
		entr.SetSlice(val)
	}
	netDev.setterFrom(netYaml, "", setWrap, setSliceWrap)
}

/*
Populates all fields of NetDevEntry with SetAlt from the
values of NetDevs. The string profileName is used to
destermine from which source/NodeInfo the entry came
from.
Actually not used, just for completeness.
*/
func (netDev *NetDevEntry) SetAltFrom(netYaml *NetDevs, profileName string) {
	netDev.setterFrom(netYaml, profileName, (*Entry).SetAlt, (*Entry).SetAltSlice)
}

/*
Populates all fields of NodeInfo with SetDefault from the
values of NodeConf.
*/
func (netDev *NetDevEntry) SetDefFrom(netYaml *NetDevs) {
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
func (netDev *NetDevEntry) setterFrom(netYaml *NetDevs, nameArg string,
	setter func(*Entry, string, string),
	setterSlice func(*Entry, []string, string)) {
	// check if netYaml is empty
	if netYaml == nil {
		netYaml = new(NetDevs)
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

/*
Create a string slice, where every element represents a yaml entry
*/
func (nodeConf *NodeConf) UnmarshalConf(excludeList []string) (lines []string) {
	nodeInfoType := reflect.TypeOf(nodeConf)
	nodeInfoVal := reflect.ValueOf(nodeConf)
	// now iterate of every field
	for i := 0; i < nodeInfoVal.Elem().NumField(); i++ {
		if nodeInfoType.Elem().Field(i).Tag.Get("lopt") != "" {
			if ymlStr, ok := getYamlString(nodeInfoType.Elem().Field(i), excludeList); ok {
				lines = append(lines, ymlStr...)
			}
		} else if nodeInfoType.Elem().Field(i).Type.Kind() == reflect.Ptr {
			nestType := reflect.TypeOf(nodeInfoVal.Elem().Field(i).Interface())
			if ymlStr, ok := getYamlString(nodeInfoType.Elem().Field(i), excludeList); ok {
				lines = append(lines, ymlStr...)
			}
			for j := 0; j < nestType.Elem().NumField(); j++ {
				if nestType.Elem().Field(j).Tag.Get("lopt") != "" &&
					!util.InSlice(excludeList, nestType.Elem().Field(j).Tag.Get("lopt")) {
					if ymlStr, ok := getYamlString(nestType.Elem().Field(j), excludeList); ok {
						for _, str := range ymlStr {
							lines = append(lines, "  "+str)
						}
					}
				}
			}
		} else if nodeInfoType.Elem().Field(i).Type == reflect.TypeOf(map[string]*NetDevs(nil)) {
			netMap := nodeInfoVal.Elem().Field(i).Interface().(map[string]*NetDevs)
			// add a default network so that it can hold values
			key := "default"
			if len(netMap) == 0 {
				netMap[key] = new(NetDevs)
			} else {
				for keyIt := range netMap {
					key = keyIt
					break
				}
			}
			if ymlStr, ok := getYamlString(nodeInfoType.Elem().Field(i), excludeList); ok {
				lines = append(lines, ymlStr[0]+":", "  "+key+":")
				netType := reflect.TypeOf(netMap[key])
				for j := 0; j < netType.Elem().NumField(); j++ {
					if ymlStr, ok := getYamlString(netType.Elem().Field(j), excludeList); ok {
						for _, str := range ymlStr {
							lines = append(lines, "  "+str)
						}
					}
				} // lines
			} // this
		} //not
	} //do
	return lines
}

/*
Get the string of the yaml tag
*/
func getYamlString(myType reflect.StructField, excludeList []string) ([]string, bool) {
	ymlStr := myType.Tag.Get("yaml")
	if len(strings.Split(ymlStr, ",")) > 1 {
		ymlStr = strings.Split(ymlStr, ",")[0]
	}
	if util.InSlice(excludeList, ymlStr) {
		return []string{""}, false
	} else if myType.Tag.Get("lopt") == "" && myType.Type.Kind() == reflect.String {
		return []string{""}, false
	}
	if myType.Type.Kind() == reflect.String {
		ymlStr += ": string"
		return []string{ymlStr}, true
	} else if myType.Type == reflect.TypeOf([]string{}) {
		return []string{ymlStr + ":", "  - string"}, true
	} else if myType.Type == reflect.TypeOf(map[string]string{}) {
		return []string{ymlStr + ":", "  key: value"}, true
	} else if myType.Type.Kind() == reflect.Ptr {
		return []string{ymlStr + ":"}, true
	}
	return []string{ymlStr}, true
}

/*
Set the field of the NodeConf with the given lopt name, returns true if the
field was found. String slices must be comma separated. Network must have the form
net.$NETNAME.lopt or netname.$NETNAME.lopt
*/
func (nodeConf *NodeConf) SetLopt(lopt string, value string) (found bool) {
	found = false
	nodeInfoType := reflect.TypeOf(nodeConf)
	nodeInfoVal := reflect.ValueOf(nodeConf)
	// try to find the normal fields, networks come later
	for i := 0; i < nodeInfoVal.Elem().NumField(); i++ {
		//fmt.Println(nodeInfoType.Elem().Field(i).Tag.Get("lopt"), lopt)
		if nodeInfoType.Elem().Field(i).Tag.Get("lopt") == lopt {
			if nodeInfoType.Elem().Field(i).Type.Kind() == reflect.String {
				wwlog.Verbose("Found lopt %s mapping to %s, setting to %s\n",
					lopt, nodeInfoType.Elem().Field(i).Name, value)
				confVal := nodeInfoVal.Elem().Field(i).Addr().Interface().(*string)
				*confVal = value
				found = true
			} else if nodeInfoType.Elem().Field(i).Type == reflect.TypeOf([]string{}) {
				wwlog.Verbose("Found lopt %s mapping to %s, setting to %s\n",
					lopt, nodeInfoType.Elem().Field(i).Name, value)
				confVal := nodeInfoVal.Elem().Field(i).Addr().Interface().(*[]string)
				*confVal = strings.Split(value, ",")
				found = true
			}
		}
	}
	// check network
	loptSlice := strings.Split(lopt, ".")
	wwlog.Debug("Trying to get network out of %s\n", loptSlice)
	if !found && len(loptSlice) == 3 && (loptSlice[0] == "net" || loptSlice[0] == "network" || loptSlice[0] == "netname") {
		if nodeConf.NetDevs == nil {
			nodeConf.NetDevs = make(map[string]*NetDevs)
		}
		if nodeConf.NetDevs[loptSlice[1]] == nil {
			nodeConf.NetDevs[loptSlice[1]] = new(NetDevs)
		}
		netInfoType := reflect.TypeOf(nodeConf.NetDevs[loptSlice[1]])
		netInfoVal := reflect.ValueOf(nodeConf.NetDevs[loptSlice[1]])
		for i := 0; i < netInfoVal.Elem().NumField(); i++ {
			if netInfoType.Elem().Field(i).Tag.Get("lopt") == loptSlice[2] {
				if netInfoType.Elem().Field(i).Type.Kind() == reflect.String {
					wwlog.Verbose("Found lopt %s for network %s mapping to %s, setting to %s\n",
						lopt, loptSlice[1], netInfoType.Elem().Field(i).Name, value)
					confVal := netInfoVal.Elem().Field(i).Addr().Interface().(*string)
					*confVal = value
					found = true
				} else if netInfoType.Elem().Field(i).Type == reflect.TypeOf([]string{}) {
					wwlog.Verbose("Found lopt %s for network %s mapping to %s, setting to %s\n",
						lopt, loptSlice[1], netInfoType.Elem().Field(i).Name, value)
					confVal := netInfoVal.Elem().Field(i).Addr().Interface().(*[]string)
					*confVal = strings.Split(value, ",")
					found = true
				}
			}
		}
	}
	return found
}
