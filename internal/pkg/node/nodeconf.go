package node

import (
	// "fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)


// A NodeConf represents a node (or a profile) as directly encoded in
// configuration (e.g., nodes.conf). As such, it does not present the
// final effective configuration data. For example, a NodeConf does
// not include values inherited from a node's profile.
type NodeConf struct {
	Comment       string                  `yaml:"comment,omitempty" lopt:"comment" comment:"Set arbitrary string comment"`
	ClusterName   string                  `yaml:"cluster name,omitempty" lopt:"cluster" sopt:"c" comment:"Set cluster group"`
	ContainerName string                  `yaml:"container name,omitempty" lopt:"container" sopt:"C" comment:"Set container name"`
	Ipxe          string                  `yaml:"ipxe template,omitempty" lopt:"ipxe" comment:"Set the iPXE template name"`
	// Deprecated start
	// Backwards compatibility only; replaced by KernelConf
	KernelVersion  string `yaml:"kernel version,omitempty"`
	KernelOverride string `yaml:"kernel override,omitempty"`
	KernelArgs     string `yaml:"kernel args,omitempty"`
	// Backwards compatibility only; replaced by IpmiConf
	IpmiUserName  string `yaml:"ipmi username,omitempty"`
	IpmiPassword  string `yaml:"ipmi password,omitempty"`
	IpmiIpaddr    string `yaml:"ipmi ipaddr,omitempty"`
	IpmiNetmask   string `yaml:"ipmi netmask,omitempty"`
	IpmiPort      string `yaml:"ipmi port,omitempty"`
	IpmiGateway   string `yaml:"ipmi gateway,omitempty"`
	IpmiInterface string `yaml:"ipmi interface,omitempty"`
	IpmiWrite     string `yaml:"ipmi write,omitempty"`
	// Deprecated end
	RuntimeOverlay []string               `yaml:"runtime overlay,omitempty" lopt:"runtime" sopt:"R" comment:"Set the runtime overlay"`
	SystemOverlay  []string               `yaml:"system overlay,omitempty" lopt:"wwinit" sopt:"O" comment:"Set the system overlay"`
	Kernel         *KernelConf            `yaml:"kernel,omitempty"`
	Ipmi           *IpmiConf              `yaml:"ipmi,omitempty"`
	Init           string                 `yaml:"init,omitempty" lopt:"init" sopt:"i" comment:"Define the init process to boot the container"`
	Root           string                 `yaml:"root,omitempty" lopt:"root" comment:"Define the rootfs" `
	AssetKey       string                 `yaml:"asset key,omitempty" lopt:"asset" comment:"Set the node's Asset tag (key)"`
	Discoverable   string                 `yaml:"discoverable,omitempty" lopt:"discoverable" comment:"Make discoverable in given network (yes/no)"`
	Profiles       []string               `yaml:"profiles,omitempty" lopt:"profile" sopt:"P" comment:"Set the node's profile members (comma separated)"`
	NetDevs        map[string]*NetDevConf `yaml:"network devices,omitempty"`
	Tags           map[string]string      `yaml:"tags,omitempty" lopt:"tagadd" comment:"base key"`
	// Not written to disk
	TagsDel        []string               `yaml:"tagsdel,omitempty" lopt:"tagdel" comment:"remove this tags"`
	// Backwards compatibility only
	Keys           map[string]string      `yaml:"keys,omitempty"`
	PrimaryNetDev  string                 `yaml:"primary network,omitempty" lopt:"primarynet" sopt:"p" comment:"Set the primary network interface"`
}

// An IpmiConf represents the IPMI configuration for a node (or a
// profile) as directly encoded in configuration (E.g.,
// nodes.conf). Referenced as pat of a NodeConf.
type IpmiConf struct {
	UserName  string            `yaml:"username,omitempty" lopt:"ipmiuser" comment:"Set the IPMI username"`
	Password  string            `yaml:"password,omitempty" lopt:"ipmipass" comment:"Set the IPMI password"`
	Ipaddr    string            `yaml:"ipaddr,omitempty" lopt:"ipmiaddr" comment:"Set the IPMI IP address"`
	Netmask   string            `yaml:"netmask,omitempty" lopt:"ipminetmask" comment:"Set the IPMI netmask"`
	Port      string            `yaml:"port,omitempty" lopt:"ipmiport" comment:"Set the IPMI port"`
	Gateway   string            `yaml:"gateway,omitempty" lopt:"ipmigateway" comment:"Set the IPMI gateway"`
	Interface string            `yaml:"interface,omitempty" lopt:"ipmiinterface" comment:"Set the node's IPMI interface (defaults: 'lan')"`
	Write     string            `yaml:"write,omitempty" lopt:"ipmiwrite" comment:"Enable the write of impi configuration (yes/no)"`
	Tags      map[string]string `yaml:"tags,omitempty" lopt:"ipmitagadd" comment:"add ipmitags"`
	// Not written to disk
	TagsDel   []string          `yaml:"tagsdel,omitempty" lopt:"ipmitagdel" comment:"remove ipmitags"`
}


// A KernelConf represents the kernel configuration for  configuration for a node (or a
// profile) as directly encoded in configuration (E.g.,
// nodes.conf). Referenced as pat of a NodeConf.
type KernelConf struct {
	Version  string `yaml:"version,omitempty"`
	Override string `yaml:"override,omitempty" lopt:"kerneloverride" sopt:"K" comment:"Set kernel override version"`
	Args     string `yaml:"args,omitempty" lopt:"kernelargs" sopt:"A" comment:"Set Kernel argument"`
}


type NetDevConf struct {
	Type    string            `yaml:"type,omitempty" lopt:"type" sopt:"T" comment:"Set device type of given network"`
	OnBoot  string            `yaml:"onboot,omitempty" lopt:"onboot" comment:"Enable/disable network device (yes/no)"`
	Device  string            `yaml:"device,omitempty" lopt:"netdev" sopt:"N" comment:"Set the device for given network"`
	Hwaddr  string            `yaml:"hwaddr,omitempty" lopt:"hwaddr" sopt:"H" comment:"Set the device's HW address for given network"`
	Ipaddr  string            `yaml:"ipaddr,omitempty" comment:"IPv4 address in given network" sopt:"I" lopt:"ipaddr"`
	IpCIDR  string            `yaml:"ipcidr,omitempty"`
	Ipaddr6 string            `yaml:"ip6addr,omitempty" lopt:"ipaddr6" comment:"IPv6 address"`
	Prefix  string            `yaml:"prefix,omitempty"`
	Netmask string            `yaml:"netmask,omitempty" lopt:"netmask" sopt:"M" comment:"Set the networks netmask"`
	Gateway string            `yaml:"gateway,omitempty" lopt:"gateway" sopt:"G" comment:"Set the node's network device gateway"`
	MTU     string            `yaml:"mtu,omitempty" lopt:"mtu" comment:"Set the mtu"`
	Tags    map[string]string `yaml:"tags,omitempty" lopt:"nettagadd" comment:"network tags"`
	TagsDel []string          `yaml:"tagsdel,omitempty" lopt:"nettagdel" comment:"delete network tags"` // should not go to disk only to wire
}


/*
Filter a given map of NodeConf against given regular expression.
*/
func FilterMapByName(inputMap map[string]*NodeConf, searchList []string) (retMap map[string]*NodeConf) {
	retMap = map[string]*NodeConf{}
	if len(searchList) > 0 {
		for _, search := range searchList {
			for name, nConf := range inputMap {
				if match, _ := regexp.MatchString("^"+search+"$", name); match {
					retMap[name] = nConf
				}
			}

		}
	}
	return retMap
}


/*
Create an empty node NodeConf
*/
func NewConf() (nodeconf NodeConf) {
	nodeconf.Ipmi = new(IpmiConf)
	nodeconf.Kernel = new(KernelConf)
	nodeconf.NetDevs = make(map[string]*NetDevConf)
	return nodeconf
}


// /*
// Get a entry by its name
// */
// func GetByName(node interface{}, name string) (string, error) {
// 	valEntry := reflect.ValueOf(node)
// 	entryField := valEntry.Elem().FieldByName(name)
// 	if entryField == (reflect.Value{}) {
// 		return "", fmt.Errorf("couldn't find field with name: %s", name)
// 	}
// 	if entryField.Type() != reflect.TypeOf(Entry{}) {
// 		return "", fmt.Errorf("field %s is not of type node.Entry", name)
// 	}
// 	myEntry := entryField.Interface().(Entry)
// 	return myEntry.Get(), nil
// }


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
Abstract function which populates a NodeConf from the given NodeInfo
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
					netMapPtr := confField.Addr().Interface().(*map[string](*NetDevConf))
					*netMapPtr = make(map[string](*NetDevConf))
				}
				nestedMap := nodeInfoVal.Field(i).Interface().(map[string]*NetDevEntry)
				netMap := confField.Interface().(map[string](*NetDevConf))
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
						netMap[netName] = new(NetDevConf)
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
		} else if nodeInfoType.Elem().Field(i).Type == reflect.TypeOf(map[string]*NetDevConf(nil)) {
			netMap := nodeInfoVal.Elem().Field(i).Interface().(map[string]*NetDevConf)
			// add a default network so that it can hold values
			key := "default"
			if len(netMap) == 0 {
				netMap[key] = new(NetDevConf)
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
			nodeConf.NetDevs = make(map[string]*NetDevConf)
		}
		if nodeConf.NetDevs[loptSlice[1]] == nil {
			nodeConf.NetDevs[loptSlice[1]] = new(NetDevConf)
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


// CompatibilityUpdate ports values from old NodeConf attributes to
// new NodeConf attributes for backwards-compatibility.
func (nodeConf *NodeConf) CompatibilityUpdate () {
	// If any "keys" are configured, convert them to tags.
	if len(nodeConf.Tags) == 0 {
		nodeConf.Tags = make(map[string]string)
	}
	for keyname, keyval := range nodeConf.Keys {
		nodeConf.Tags[keyname] = keyval
		delete(nodeConf.Keys, keyname)
	}

	if nodeConf.Ipmi == nil {
		nodeConf.Ipmi = new(IpmiConf)
	}

	if nodeConf.Ipmi.Ipaddr == "" {
		nodeConf.Ipmi.Ipaddr = nodeConf.IpmiIpaddr
	}
	nodeConf.IpmiIpaddr = ""

	if nodeConf.Ipmi.Netmask == "" {
		nodeConf.Ipmi.Netmask = nodeConf.IpmiNetmask
	}
	nodeConf.IpmiNetmask = ""

	if nodeConf.Ipmi.Port == "" {
		nodeConf.Ipmi.Port = nodeConf.IpmiPort
	}
	nodeConf.IpmiPort = ""

	if nodeConf.Ipmi.Gateway == "" {
		nodeConf.Ipmi.Gateway = nodeConf.IpmiGateway
	}
	nodeConf.IpmiGateway = ""

	if nodeConf.Ipmi.UserName == "" {
		nodeConf.Ipmi.UserName = nodeConf.IpmiUserName
	}
	nodeConf.IpmiUserName = ""

	if nodeConf.Ipmi.Password == "" {
		nodeConf.Ipmi.Password = nodeConf.IpmiPassword
	}
	nodeConf.IpmiPassword = ""

	if nodeConf.Ipmi.Interface == "" {
		nodeConf.Ipmi.Interface = nodeConf.IpmiInterface
	}
	nodeConf.IpmiInterface = ""

	if nodeConf.Ipmi.Write == "" {
		nodeConf.Ipmi.Write = nodeConf.IpmiWrite
	}
	nodeConf.IpmiWrite = ""

	if nodeConf.Kernel == nil {
		nodeConf.Kernel = new(KernelConf)
	}

	if nodeConf.Kernel.Args == "" {
		nodeConf.Kernel.Args = nodeConf.KernelArgs
	}
	nodeConf.KernelArgs = ""

	if nodeConf.Kernel.Override == "" {
		nodeConf.Kernel.Override = nodeConf.KernelOverride
	}
	nodeConf.KernelOverride = ""

	if nodeConf.Kernel.Override == "" {
		nodeConf.Kernel.Override = nodeConf.KernelVersion
	}
	nodeConf.KernelVersion = ""
}
