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
				for key, val := range entryMap {
					confField.Interface().(map[string]string)[key] = getter(val)
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
						for key, val := range entryMap {
							nestedVal.Interface().(map[string]string)[key] = getter(val)
						}
					}
					//}
				}

			} else if nodeInfoVal.Field(i).Type() == reflect.TypeOf(map[string]*NetDevEntry{}) {
				if confField.IsNil() {
					netMapPtr := confField.Addr().Interface().(*map[string](*NetDevs))
					*netMapPtr = make(map[string](*NetDevs))
				}
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
								*newVal = getter((netVal.Addr().Interface()).(*Entry))
							} else if netVal.Type() == reflect.TypeOf(map[string]string{}) {
								for key, val := range (netVal.Interface()).(map[string]*string) {
									*val = getter(netConfVal.Elem().Field(j).Interface().((map[string](*Entry)))[key])
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
								setter(netInfoVal.Elem().Field(j).Addr().Interface().((*Entry)), netVal.String(), nameArg)
							} else if netVal.Type() == reflect.TypeOf(map[string]string{}) {
								// normaly the map should be created here, but did not manage it
								for key, val := range (netVal.Interface()).(map[string]string) {
									entr := new(Entry)
									setter(entr, val, nameArg)
									netInfoVal.Elem().Field(j).Interface().((map[string](*Entry)))[key] = entr
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
