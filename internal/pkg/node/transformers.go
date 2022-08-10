package node

import (
	"reflect"
)

/*
Populates a NodeConf struct (the one which goes to disk) from a
NodeInfo (which just lives in memory), with the values from all
the underlying entries using GetReal, so just the explicit values
go do disk.
*/
func (nodeConf *NodeConf) GetRealFrom(nodeInfo NodeInfo) {
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
					*newValue = entryVal.GetReal()
				} else if confField.Type() == reflect.TypeOf([]string{}) {
					newValue := (confField.Addr().Interface()).(*[]string)
					entryVal := nodeInfoVal.Field(i).Interface().(Entry)
					*newValue = entryVal.GetRealSlice()
				}
			} else if nodeInfoVal.Field(i).Type() == reflect.TypeOf(map[string]*Entry{}) {
				entryMap := nodeInfoVal.Field(i).Interface().(map[string]*Entry)
				for key, val := range entryMap {
					confField.Interface().(map[string]string)[key] = val.GetReal()
				}
			} else if nodeInfoVal.Field(i).Type().Kind() == reflect.Ptr {
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
							*newValue = entryVal.GetReal()
						} else if nestedVal.Type() == reflect.TypeOf([]string{}) {
							newValue := (nestedVal.Addr().Interface()).(*[]string)
							entryVal := nestedInfoVal.Elem().Field(j).Interface().(Entry)
							*newValue = entryVal.GetRealSlice()

						}
					} else if nestedInfoVal.Elem().Field(j).Type() == reflect.TypeOf(map[string]*Entry{}) {
						if nestedVal.IsNil() {
							mapPtr := nestedVal.Addr().Interface().(*map[string]string)
							*mapPtr = make(map[string]string)
						}
						entryMap := nestedInfoVal.Elem().Field(j).Interface().(map[string]*Entry)
						for key, val := range entryMap {
							nestedVal.Interface().(map[string]string)[key] = val.GetReal()
						}
					}
					//}
				}

			} else if nodeInfoVal.Field(i).Type() == reflect.TypeOf(map[string]*NetDevEntry{}) {
				nestedMap := nodeInfoVal.Field(i).Interface().(map[string]*NetDevEntry)
				for netName, netVal := range nestedMap {
					netValsType := reflect.ValueOf(netVal)
					netMap := confField.Interface().(map[string](*NetDevs))
					var newNet NetDevs
					newNet.Tags = make(map[string]string)
					netMap[netName] = &newNet
					netConfType := reflect.TypeOf(newNet)
					netConfVal := reflect.ValueOf(&newNet)
					for j := 0; j < netConfType.NumField(); j++ {
						netVal := netValsType.Elem().FieldByName(netConfType.Field(j).Name)
						if netVal.IsValid() {
							if netVal.Type() == reflect.TypeOf(Entry{}) {
								newVal := netConfVal.Elem().Field(j).Addr().Interface().((*string))
								*newVal = (netVal.Addr().Interface()).(*Entry).GetReal()
							} else if netVal.Type() == reflect.TypeOf(map[string]string{}) {
								// normaly the map should be created here, but did not manage it
								for key, val := range (netVal.Interface()).(map[string]string) {
									var entr Entry
									entr.Set(val)
									netConfVal.Elem().Field(j).Interface().((map[string](*Entry)))[key] = &entr
								}
							}
						}

					}
				}
			}
		}
		/* else {
			// NodeInfo has the Id field, nodeConf not
			fmt.Println("INVALID", nodeInfoType.Field(i).Name)
		}
		*/
	}
}

/*
Populates a NodeConf struct from a NodeInfo, with the combined
values from the underlying entries using Get.
*/
func (nodeConf *NodeConf) GetFrom(nodeInfo NodeInfo) {
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
					*newValue = entryVal.Get()
				} else if confField.Type() == reflect.TypeOf([]string{}) {
					newValue := (confField.Addr().Interface()).(*[]string)
					entryVal := nodeInfoVal.Field(i).Interface().(Entry)
					*newValue = entryVal.GetSlice()
				}
			} else if nodeInfoVal.Field(i).Type() == reflect.TypeOf(map[string]*Entry{}) {
				if confField.IsNil() {
					confFieldPtr := confField.Addr().Interface().(*map[string]string)
					*confFieldPtr = make(map[string]string)
				}
				entryMap := nodeInfoVal.Field(i).Interface().(map[string]*Entry)
				for key, val := range entryMap {
					confField.Interface().(map[string]string)[key] = val.Get()
				}
			} else if nodeInfoVal.Field(i).Type().Kind() == reflect.Ptr {
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
							*newValue = entryVal.Get()
						} else if nestedVal.Type() == reflect.TypeOf([]string{}) {
							newValue := (nestedVal.Addr().Interface()).(*[]string)
							entryVal := nestedInfoVal.Elem().Field(j).Interface().(Entry)
							*newValue = entryVal.GetSlice()

						}
					} else if nestedInfoVal.Elem().Field(j).Type() == reflect.TypeOf(map[string]*Entry{}) {
						if nestedVal.IsNil() {
							mapPtr := nestedVal.Addr().Interface().(*map[string]string)
							*mapPtr = make(map[string]string)
						}
						entryMap := nestedInfoVal.Elem().Field(j).Interface().(map[string]*Entry)
						for key, val := range entryMap {
							nestedVal.Interface().(map[string]string)[key] = val.Get()
						}
					}
				}
			} else if nodeInfoVal.Field(i).Type() == reflect.TypeOf(map[string]*NetDevEntry{}) {
				nestedMap := nodeInfoVal.Field(i).Interface().(map[string]*NetDevEntry)
				for netName, netVal := range nestedMap {
					netValsType := reflect.ValueOf(netVal)
					if confField.IsNil() {
						netMapPtr := confField.Addr().Interface().(*map[string](*NetDevs))
						*netMapPtr = make(map[string](*NetDevs))
					}
					netMap := confField.Interface().(map[string](*NetDevs))
					var newNet NetDevs
					newNet.Tags = make(map[string]string)
					netMap[netName] = &newNet
					netConfType := reflect.TypeOf(newNet)
					netConfVal := reflect.ValueOf(&newNet)
					for j := 0; j < netConfType.NumField(); j++ {
						netVal := netValsType.Elem().FieldByName(netConfType.Field(j).Name)
						if netVal.IsValid() {
							if netVal.Type() == reflect.TypeOf(Entry{}) {
								newVal := netConfVal.Elem().Field(j).Addr().Interface().((*string))
								*newVal = (netVal.Addr().Interface()).(*Entry).Get()
							} else if netVal.Type() == reflect.TypeOf(map[string]string{}) {
								// normaly the map should be created here, but did not manage it
								for key, val := range (netVal.Interface()).(map[string]string) {
									var entr Entry
									entr.Set(val)
									netConfVal.Elem().Field(j).Interface().((map[string](*Entry)))[key] = &entr
								}
							}
						}

					}
				}
			}
		} /*else {
			// NodeInfo has the Id field, nodeConf not
			fmt.Println("INVALID", nodeInfoType.Field(i).Name)
		} */
	}
}

/*
Populates all fields of NodeInfo with Set from the
values of NodeConf.
*/
func (node *NodeInfo) SetFrom(n *NodeConf) {
	// get the full memory, taking the shortcut and init Ipmi and Kernel directly
	if node.Kernel == nil {
		node.Kernel = new(KernelEntry)
	}
	if node.Ipmi == nil {
		node.Ipmi = new(IpmiEntry)
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
					(nodeInfoVal.Elem().Field(i).Addr().Interface()).(*Entry).Set(valField.String())
				} else if valField.Type() == reflect.TypeOf([]string{}) {
					(nodeInfoVal.Elem().Field(i).Addr().Interface()).(*Entry).SetSlice(valField.Interface().([]string))
				}
			} else if nodeInfoType.Elem().Field(i).Type.Kind() == reflect.Ptr && !valField.IsZero() {
				nestedInfoType := reflect.TypeOf(nodeInfoVal.Elem().Field(i).Interface())
				netstedInfoVal := reflect.ValueOf(nodeInfoVal.Elem().Field(i).Interface())
				nestedConfVal := reflect.ValueOf(valField.Interface())
				for j := 0; j < nestedInfoType.Elem().NumField(); j++ {
					nestedVal := nestedConfVal.Elem().FieldByName(nestedInfoType.Elem().Field(j).Name)
					if nestedVal.IsValid() {
						if netstedInfoVal.Elem().Field(j).Type() == reflect.TypeOf(Entry{}) {
							netstedInfoVal.Elem().Field(j).Addr().Interface().(*Entry).Set(nestedVal.String())
						} else {
							confMap := nestedVal.Interface().(map[string]string)
							if netstedInfoVal.Elem().Field(j).IsNil() {
								newMap := make(map[string]*Entry)
								mapPtr := (netstedInfoVal.Elem().Field(j).Addr().Interface()).(*map[string](*Entry))
								*mapPtr = newMap
							}
							for key, val := range confMap {
								var entr Entry
								entr.Set(val)
								(netstedInfoVal.Elem().Field(j).Interface()).(map[string](*Entry))[key] = &entr
							}
						}
					}
				}
			} else if nodeInfoType.Elem().Field(i).Type == reflect.TypeOf(map[string](*Entry)(nil)) {
				confMap := valField.Interface().(map[string]string)
				for key, val := range confMap {
					var entr Entry
					entr.Set(val)
					(nodeInfoVal.Elem().Field(i).Interface()).(map[string](*Entry))[key] = &entr
				}
			} else if nodeInfoType.Elem().Field(i).Type == reflect.TypeOf(map[string](*NetDevEntry)(nil)) {
				nestedMap := valField.Interface().(map[string](*NetDevs))
				for netName, netVals := range nestedMap {
					netValsType := reflect.ValueOf(netVals)
					netMap := nodeInfoVal.Elem().Field(i).Interface().(map[string](*NetDevEntry))
					var newNet NetDevEntry
					newNet.Tags = make(map[string]*Entry)
					// This should be done a bit down, but didn't know how to do it
					netMap[netName] = &newNet
					netInfoType := reflect.TypeOf(newNet)
					netInfoVal := reflect.ValueOf(&newNet)
					for j := 0; j < netInfoType.NumField(); j++ {
						netVal := netValsType.Elem().FieldByName(netInfoType.Field(j).Name)
						if netVal.IsValid() {
							if netVal.Type().Kind() == reflect.String {
								netInfoVal.Elem().Field(j).Addr().Interface().((*Entry)).Set(netVal.String())
								if netInfoType.Field(j).Name == "Netmask" {
									netInfoVal.Elem().Field(j).Addr().Interface().((*Entry)).SetDefault("255.255.255.0")
								}
							} else if netVal.Type() == reflect.TypeOf(map[string]string{}) {
								// normaly the map should be created here, but did not manage it
								for key, val := range (netVal.Interface()).(map[string]string) {
									var entr Entry
									entr.Set(val)
									netInfoVal.Elem().Field(j).Interface().((map[string](*Entry)))[key] = &entr
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
Populates all fields of NodeInfo with SetAlt from the
values of NodeConf. The string profileName is used to
destermine from which source/NodeInfo the entry came
from.
*/
func (node *NodeConf) SetAltFrom(nodeInfo NodeInfo, profileName string) {
	nodeInfoVal := reflect.ValueOf(&nodeInfo)
	nodeInfoType := reflect.TypeOf(&nodeInfo)
	profileConfVal := reflect.ValueOf(node)
	for i := 0; i < nodeInfoType.Elem().NumField(); i++ {
		valField := profileConfVal.Elem().FieldByName(nodeInfoType.Elem().Field(i).Name)
		if valField.IsValid() {
			// found field with same name for Conf and Info
			if nodeInfoType.Elem().Field(i).Type == reflect.TypeOf(Entry{}) {
				if valField.Type().Kind() == reflect.String {
					(nodeInfoVal.Elem().Field(i).Addr().Interface()).(*Entry).SetAlt(valField.String(), profileName)
				} else if valField.Type() == reflect.TypeOf([]string{}) {
					(nodeInfoVal.Elem().Field(i).Addr().Interface()).(*Entry).SetAltSlice(valField.Interface().([]string), profileName)
				}
			} else if nodeInfoType.Elem().Field(i).Type.Kind() == reflect.Ptr && !valField.IsZero() {
				nestedInfoType := reflect.TypeOf(nodeInfoVal.Elem().Field(i).Interface())
				netstedInfoVal := reflect.ValueOf(nodeInfoVal.Elem().Field(i).Interface())
				nestedConfVal := reflect.ValueOf(valField.Interface())
				for j := 0; j < nestedInfoType.Elem().NumField(); j++ {
					nestedVal := nestedConfVal.Elem().FieldByName(nestedInfoType.Elem().Field(j).Name)
					if nestedVal.IsValid() {
						if netstedInfoVal.Elem().Field(j).Type() == reflect.TypeOf(Entry{}) {
							netstedInfoVal.Elem().Field(j).Addr().Interface().(*Entry).SetAlt(nestedVal.String(), profileName)
						}
					}
				}
			} else if nodeInfoType.Elem().Field(i).Type == reflect.TypeOf(map[string](*Entry)(nil)) {
				confMap := valField.Interface().(map[string]string)
				for key, val := range confMap {
					var entr Entry
					entr.SetAlt(val, profileName)
					(nodeInfoVal.Elem().Field(i).Interface()).(map[string](*Entry))[key] = &entr
				}
			} else if nodeInfoType.Elem().Field(i).Type == reflect.TypeOf(map[string](*NetDevEntry)(nil)) {
				nestedMap := valField.Interface().(map[string](*NetDevs))
				for netName, netVals := range nestedMap {
					netValsType := reflect.ValueOf(netVals)
					netMap := nodeInfoVal.Elem().Field(i).Interface().(map[string](*NetDevEntry))
					var newNet NetDevEntry
					newNet.Tags = make(map[string]*Entry)
					// This should be done a bit down, but didn'tknow how to do it
					netMap[netName] = &newNet
					netInfoType := reflect.TypeOf(newNet)
					netInfoVal := reflect.ValueOf(&newNet)
					for j := 0; j < netInfoType.NumField(); j++ {
						netVal := netValsType.Elem().FieldByName(netInfoType.Field(j).Name)
						if netVal.IsValid() {
							if netVal.Type().Kind() == reflect.String {
								netInfoVal.Elem().Field(j).Addr().Interface().((*Entry)).SetAlt(netVal.String(), profileName)
							} else if netVal.Type() == reflect.TypeOf(map[string]string{}) {
								// normally the map should be created here, but did not manage it
								for key, val := range (netVal.Interface()).(map[string]string) {
									var entr Entry
									entr.SetAlt(val, profileName)
									netInfoVal.Elem().Field(j).Interface().((map[string](*Entry)))[key] = &entr
								}
							}
						}
					}
				}
			}
		}
	}
}
