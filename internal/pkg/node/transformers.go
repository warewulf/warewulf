package node

import (
	"reflect"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

/*
Populates a NodeConf struct (the one which goes to disk) from a
NodeInfo (which just lives in memory), with the values from all
the underlying entries using GetReal, so just the explicit values
go do disk.
*/
func (nodeConf *NodeConf) GetRealFrom(nodeInfo NodeInfo) {
	nodeConf.getterFrom(nodeInfo, (*Entry).GetReal, (*Entry).GetRealSlice)
}

/*
Populates a NodeConf struct from a NodeInfo, with the combined
values from the underlying entries using Get.
*/
func (nodeConf *NodeConf) GetFrom(nodeInfo NodeInfo) {
	nodeConf.getterFrom(nodeInfo, (*Entry).Get, (*Entry).GetSlice)
}

/*
Abstract function which populates a NodeConf form the given NodeInfo
via getter functions.
*/
func (nodeConf *NodeConf) getterFrom(nodeInfo NodeInfo,
	getter func(*Entry) string,
	getterSlice func(*Entry) []string) {
	nodeInfoType := reflect.TypeOf(nodeInfo)
	nodeInfoVal := reflect.ValueOf(nodeInfo)
	configVal := reflect.ValueOf(nodeConf)
	// now iterate of every field
	for i := 0; i < nodeInfoType.NumField(); i++ {
		// found field with same name for Conf and Info
		confField := configVal.Elem().FieldByName(nodeInfoType.Field(i).Name)
		if confField.IsValid() {
			if nodeInfoVal.Field(i).Type() == reflect.TypeOf(Entry{}) {
				if confField.Type().Kind() == reflect.String {
					newValue := (confField.Addr().Interface()).(*string)
					entryVal := nodeInfoVal.Field(i).Interface().(Entry)
					*newValue = getter(&entryVal)
				} else if confField.Type() == reflect.TypeOf([]string{}) {
					newValue := (confField.Addr().Interface()).(*[]string)
					entryVal := nodeInfoVal.Field(i).Interface().(Entry)
					*newValue = getterSlice(&entryVal)
				}
			} else if nodeInfoVal.Field(i).Type() == reflect.TypeOf(map[string]*Entry{}) {
				entryMap := nodeInfoVal.Field(i).Interface().(map[string]*Entry)
				confMap := confField.Interface().(map[string]string)

				if len(confMap) > len(entryMap) {
					for confKey := range confMap {
						foundKey := false
						for entrKey := range entryMap {
							if confKey == entrKey {
								foundKey = true
							}
						}
						if !foundKey {
							delete(confMap, confKey)
						}
					}
				}
				for key, val := range entryMap {
					confMap[key] = getter(val)
				}
			} else if nodeInfoVal.Field(i).Type().Kind() == reflect.Ptr && !nodeInfoVal.Field(i).IsNil() {
				// initialize the nested NodeConf structs, but only if these will be set
				if confField.Addr().Elem().IsZero() {
					switch confField.Addr().Elem().Type() {
					case reflect.TypeOf((*KernelConf)(nil)):
						var newConf KernelConf
						newConfPtr := (confField.Addr().Elem().Addr().Interface()).(**KernelConf)
						*newConfPtr = &newConf
					case reflect.TypeOf((*IpmiConf)(nil)):
						var newConf IpmiConf
						newConfPtr := (confField.Addr().Elem().Addr().Interface()).(**IpmiConf)
						*newConfPtr = &newConf
					}
				}
				nestedInfoType := reflect.TypeOf(nodeInfoVal.Field(i).Interface())
				nestedInfoVal := reflect.ValueOf(nodeInfoVal.Field(i).Interface())
				nestedConfVal := reflect.ValueOf(confField.Interface())
				for j := 0; j < nestedInfoType.Elem().NumField(); j++ {
					nestedVal := nestedConfVal.Elem().FieldByName(nestedInfoType.Elem().Field(j).Name)
					if nestedInfoVal.Elem().Field(j).Type() == reflect.TypeOf(Entry{}) {
						if nestedVal.Type().Kind() == reflect.String {
							newValue := (nestedVal.Addr().Interface()).(*string)
							entryVal := nestedInfoVal.Elem().Field(j).Interface().(Entry)
							*newValue = getter(&entryVal)
						} else if nestedVal.Type() == reflect.TypeOf([]string{}) {
							newValue := (nestedVal.Addr().Interface()).(*[]string)
							entryVal := nestedInfoVal.Elem().Field(j).Interface().(Entry)
							*newValue = getterSlice(&entryVal)

						}
					} else if nestedInfoVal.Elem().Field(j).Type() == reflect.TypeOf(map[string]*Entry{}) {
						if nestedVal.IsNil() {
							mapPtr := nestedVal.Addr().Interface().(*map[string]string)
							*mapPtr = make(map[string]string)
						}
						entryMap := nestedInfoVal.Elem().Field(j).Interface().(map[string]*Entry)
						confMap := nestedVal.Interface().(map[string]string)
						if len(confMap) > len(entryMap) {
							for confKey := range confMap {
								foundKey := false
								for entrKey := range entryMap {
									if confKey == entrKey {
										foundKey = true
									}
								}
								if !foundKey {
									delete(confMap, confKey)
								}
							}
						}
						for key, val := range entryMap {
							confMap[key] = getter(val)
						}
					}
				}

			} else if nodeInfoVal.Field(i).Type() == reflect.TypeOf(map[string]*NetDevEntry{}) {
				if confField.IsNil() {
					netMapPtr := confField.Addr().Interface().(*map[string](*NetDevs))
					*netMapPtr = make(map[string](*NetDevs))
				}
				nestedMap := nodeInfoVal.Field(i).Interface().(map[string]*NetDevEntry)
				netMap := confField.Interface().(map[string](*NetDevs))
				// check if a network was deleted
				if len(netMap) > len(nestedMap) {
					for netMapKey := range netMap {
						foundKey := false
						for nestedMapKey := range nestedMap {
							if netMapKey == nestedMapKey {
								foundKey = true
							}
						}
						if !foundKey {
							delete(netMap, netMapKey)
						}
					}
				}
				for netName, netVal := range nestedMap {
					netValsType := reflect.ValueOf(netVal)
					if _, ok := netMap[netName]; !ok {
						netMap[netName] = new(NetDevs)
					}
					netConfType := reflect.TypeOf(*netMap[netName])
					netConfVal := reflect.ValueOf(netMap[netName])
					for j := 0; j < netConfType.NumField(); j++ {
						netVal := netValsType.Elem().FieldByName(netConfType.Field(j).Name)
						if netVal.IsValid() {
							if netVal.Type() == reflect.TypeOf(Entry{}) {
								newVal := netConfVal.Elem().Field(j).Addr().Interface().((*string))
								*newVal = getter((netVal.Addr().Interface()).(*Entry))
							} else if netVal.Type() == reflect.TypeOf(map[string]*Entry{}) {
								entryMap := netVal.Interface().(map[string](*Entry))
								confMap := netConfVal.Elem().Field(j).Interface().(map[string]string)
								if confMap == nil {
									confMapPtr := netConfVal.Elem().Field(j).Addr().Interface().(*map[string]string)
									*confMapPtr = make(map[string]string)
								}
								if len(confMap) > len(entryMap) {
									for confMapKey := range confMap {
										foundKey := false
										for entryMapKey := range entryMap {
											if confMapKey == entryMapKey {
												foundKey = true
											}
										}
										if !foundKey {
											delete(netConfVal.Elem().Field(j).Interface().(map[string]string), confMapKey)
										}
									}
								}
								for key, val := range entryMap {
									netConfVal.Elem().Field(j).Interface().(map[string]string)[key] = getter(val)
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
Create cmd line flags from the NodeConf fields
*/
func (nodeConf *NodeConf) CreateFlags(baseCmd *cobra.Command, excludeList []string) {
	nodeInfoType := reflect.TypeOf(nodeConf)
	nodeInfoVal := reflect.ValueOf(nodeConf)
	// now iterate of every field
	for i := 0; i < nodeInfoVal.Elem().NumField(); i++ {
		if nodeInfoType.Elem().Field(i).Tag.Get("comment") != "" &&
			!util.InSlice(excludeList, nodeInfoType.Elem().Field(i).Tag.Get("lopt")) {
			field := nodeInfoVal.Elem().Field(i)
			createFlags(baseCmd, excludeList, nodeInfoType.Elem().Field(i), &field)
		} else if nodeInfoType.Elem().Field(i).Type.Kind() == reflect.Ptr {
			nestType := reflect.TypeOf(nodeInfoVal.Elem().Field(i).Interface())
			nestVal := reflect.ValueOf(nodeInfoVal.Elem().Field(i).Interface())
			for j := 0; j < nestType.Elem().NumField(); j++ {
				field := nestVal.Elem().Field(j)
				createFlags(baseCmd, excludeList, nestType.Elem().Field(j), &field)
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
			netType := reflect.TypeOf(netMap[key])
			netVal := reflect.ValueOf(netMap[key])
			for j := 0; j < netType.Elem().NumField(); j++ {
				field := netVal.Elem().Field(j)
				createFlags(baseCmd, excludeList, netType.Elem().Field(j), &field)
			}
		}
	}
}

/*
Helper function to create the different PerisitantFlags() for different types.
*/
func createFlags(baseCmd *cobra.Command, excludeList []string,
	myType reflect.StructField, myVal *reflect.Value) {
	if myType.Tag.Get("lopt") != "" {
		if myType.Type.Kind() == reflect.String {
			ptr := myVal.Addr().Interface().(*string)
			if myType.Tag.Get("sopt") != "" {
				baseCmd.PersistentFlags().StringVarP(ptr,
					myType.Tag.Get("lopt"),
					myType.Tag.Get("sopt"),
					myType.Tag.Get("default"),
					myType.Tag.Get("comment"))
			} else if !util.InSlice(excludeList, myType.Tag.Get("lopt")) {
				baseCmd.PersistentFlags().StringVar(ptr,
					myType.Tag.Get("lopt"),
					myType.Tag.Get("default"),
					myType.Tag.Get("comment"))

			}
		} else if myType.Type == reflect.TypeOf([]string{}) {
			ptr := myVal.Addr().Interface().(*[]string)
			if myType.Tag.Get("sopt") != "" {
				baseCmd.PersistentFlags().StringSliceVarP(ptr,
					myType.Tag.Get("lopt"),
					myType.Tag.Get("sopt"),
					[]string{myType.Tag.Get("default")},
					myType.Tag.Get("comment"))
			} else if !util.InSlice(excludeList, myType.Tag.Get("lopt")) {
				baseCmd.PersistentFlags().StringSliceVar(ptr,
					myType.Tag.Get("lopt"),
					[]string{myType.Tag.Get("default")},
					myType.Tag.Get("comment"))

			}
		} else if myType.Type == reflect.TypeOf(map[string]string{}) {
			ptr := myVal.Addr().Interface().(*map[string]string)
			if myType.Tag.Get("sopt") != "" {
				baseCmd.PersistentFlags().StringToStringVarP(ptr,
					myType.Tag.Get("lopt"),
					myType.Tag.Get("sopt"),
					map[string]string{}, // empty default!
					myType.Tag.Get("comment"))
			} else if !util.InSlice(excludeList, myType.Tag.Get("lopt")) {
				baseCmd.PersistentFlags().StringToStringVar(ptr,
					myType.Tag.Get("lopt"),
					map[string]string{}, // empty default!
					myType.Tag.Get("comment"))

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
	confType := reflect.TypeOf(info)
	confVal := reflect.ValueOf(info)
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
				switch confType.Elem().Field(j).Type {
				case reflect.TypeOf((*IpmiConf)(nil)):
					ptr := confVal.Elem().Field(j).Addr().Interface().(**IpmiConf)
					*ptr = (*IpmiConf)(nil)
				case reflect.TypeOf((*KernelConf)(nil)):
					ptr := confVal.Elem().Field(j).Addr().Interface().(**KernelConf)
					*ptr = (*KernelConf)(nil)
				}
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
