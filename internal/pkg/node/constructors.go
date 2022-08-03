package node

import (
	"errors"
	"io/ioutil"
	"path"
	"reflect"
	"sort"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"

	"gopkg.in/yaml.v2"
)

var ConfigFile string

func init() {
	if ConfigFile == "" {
		ConfigFile = path.Join(buildconfig.SYSCONFDIR(), "warewulf/nodes.conf")
	}
}

func New() (NodeYaml, error) {
	var ret NodeYaml

	wwlog.Printf(wwlog.VERBOSE, "Opening node configuration file: %s\n", ConfigFile)
	data, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		return ret, err
	}

	wwlog.Printf(wwlog.DEBUG, "Unmarshaling the node configuration\n")
	err = yaml.Unmarshal(data, &ret)
	if err != nil {
		return ret, err
	}

	wwlog.Printf(wwlog.DEBUG, "Returning node object\n")

	return ret, nil
}

/*
Get all the nodes of a configuration. This function also merges
the nodes with the given profiles and set the default values
for every node
*/
func (config *NodeYaml) FindAllNodes() ([]NodeInfo, error) {
	var ret []NodeInfo
	/*
		wwconfig, err := warewulfconf.New()
		if err != nil {
			return ret, err
		}
	*/
	wwlog.Printf(wwlog.DEBUG, "Finding all nodes...\n")
	for nodename, node := range config.Nodes {
		var n NodeInfo

		wwlog.Printf(wwlog.DEBUG, "In node loop: %s\n", nodename)
		n.NetDevs = make(map[string]*NetDevEntry)
		n.Tags = make(map[string]*Entry)
		n.Kernel = new(KernelEntry)
		n.Ipmi = new(IpmiEntry)
		n.SystemOverlay.SetDefault("wwinit")
		n.RuntimeOverlay.SetDefault("generic")
		n.Ipxe.SetDefault("default")
		n.Init.SetDefault("/sbin/init")
		n.Root.SetDefault("initramfs")
		n.Kernel.Args.SetDefault("quiet crashkernel=no vga=791")

		fullname := strings.SplitN(nodename, ".", 2)
		if len(fullname) > 1 {
			n.ClusterName.SetDefault(fullname[1])
		}
		// special handling for profile to get the default one
		if len(node.Profiles) == 0 {
			n.Profiles = []string{"default"}
		} else {
			n.Profiles = node.Profiles
		}
		// node explciti nodename field in NodeConf
		n.Id.Set(nodename)
		// backward compatibilty
		for keyname, key := range node.Keys {
			node.Tags[keyname] = key
			delete(node.Keys, keyname)
		}
		nodeInfoVal := reflect.ValueOf(&n)
		nodeInfoType := reflect.TypeOf(&n)
		nodeConfVal := reflect.ValueOf(node)
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
						// This should be done a bit down, but didn'tknow how to do it
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
		// delete deprecated structures so that they do not get unmarshalled
		node.IpmiIpaddr = ""
		node.IpmiNetmask = ""
		node.IpmiGateway = ""
		node.IpmiUserName = ""
		node.IpmiPassword = ""
		node.IpmiInterface = ""
		node.IpmiWrite = ""
		// backward compatibility
		n.Kernel.Args.Set(node.KernelArgs)
		n.Kernel.Override.Set(node.KernelOverride)
		n.Kernel.Override.Set(node.KernelVersion)
		node.KernelArgs = ""
		node.KernelOverride = ""
		node.KernelVersion = ""
		// Merge Keys into Tags for backwards compatibility
		if len(node.Tags) == 0 {
			node.Tags = make(map[string]string)
		}

		for _, profileName := range n.Profiles {
			if _, ok := config.NodeProfiles[profileName]; !ok {
				wwlog.Printf(wwlog.WARN, "Profile not found for node '%s': %s\n", nodename, profileName)
				continue
			}

			wwlog.Printf(wwlog.VERBOSE, "Merging profile into node: %s <- %s\n", nodename, profileName)
			nodeInfoVal := reflect.ValueOf(&n)
			nodeInfoType := reflect.TypeOf(&n)
			profileConfVal := reflect.ValueOf(config.NodeProfiles[profileName])
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
										// normaly the map should be created here, but did not manage it
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
		ret = append(ret, n)
	}

	sort.Slice(ret, func(i, j int) bool {
		if ret[i].ClusterName.Get() < ret[j].ClusterName.Get() {
			return true
		} else if ret[i].ClusterName.Get() == ret[j].ClusterName.Get() {
			if ret[i].Id.Get() < ret[j].Id.Get() {
				return true
			}
		}
		return false
	})

	return ret, nil
}

/*
Return all profiles as NodeInfo
*/
func (config *NodeYaml) FindAllProfiles() ([]NodeInfo, error) {
	var ret []NodeInfo

	for name, profile := range config.NodeProfiles {
		var p NodeInfo
		p.NetDevs = make(map[string]*NetDevEntry)
		p.Tags = make(map[string]*Entry)
		p.Kernel = new(KernelEntry)
		p.Ipmi = new(IpmiEntry)
		p.Id.Set(name)
		p.Comment.Set(profile.Comment)
		p.ClusterName.Set(profile.ClusterName)
		p.ContainerName.Set(profile.ContainerName)
		p.Ipxe.Set(profile.Ipxe)
		p.Init.Set(profile.Init)
		// backward compatibility
		p.Kernel.Args.Set(profile.KernelArgs)
		p.Kernel.Override.Set(profile.KernelOverride)
		p.Kernel.Override.Set(profile.KernelVersion)
		profile.KernelArgs = ""
		profile.KernelOverride = ""
		profile.KernelVersion = ""
		if profile.Kernel != nil {
			p.Kernel.Args.Set(profile.Kernel.Args)
			if profile.Kernel.Override != "" {
				p.Kernel.Override.Set(profile.Kernel.Override)
			} else if profile.Kernel.Version != "" {
				p.Kernel.Override.Set(profile.Kernel.Version)
			}
		}
		// backward compatibility for old Ipmi config
		p.Ipmi.Ipaddr.Set(profile.IpmiIpaddr)
		p.Ipmi.Netmask.Set(profile.IpmiNetmask)
		p.Ipmi.Port.Set(profile.IpmiPort)
		p.Ipmi.Gateway.Set(profile.IpmiGateway)
		p.Ipmi.UserName.Set(profile.IpmiUserName)
		p.Ipmi.Password.Set(profile.IpmiPassword)
		p.Ipmi.Interface.Set(profile.IpmiInterface)
		p.Ipmi.Write.Set(profile.IpmiWrite)
		// delete deprectated structures so that they do not get unmarshalled
		profile.IpmiIpaddr = ""
		profile.IpmiNetmask = ""
		profile.IpmiGateway = ""
		profile.IpmiUserName = ""
		profile.IpmiPassword = ""
		profile.IpmiInterface = ""
		profile.IpmiWrite = ""
		if profile.Ipmi != nil {
			p.Ipmi.Netmask.Set(profile.Ipmi.Netmask)
			p.Ipmi.Port.Set(profile.Ipmi.Port)
			p.Ipmi.Gateway.Set(profile.Ipmi.Gateway)
			p.Ipmi.UserName.Set(profile.Ipmi.UserName)
			p.Ipmi.Password.Set(profile.Ipmi.Password)
			p.Ipmi.Interface.Set(profile.Ipmi.Interface)
			p.Ipmi.Write.Set(profile.Ipmi.Write)
		}
		p.RuntimeOverlay.SetSlice(profile.RuntimeOverlay)
		p.SystemOverlay.SetSlice(profile.SystemOverlay)
		p.Root.Set(profile.Root)
		p.AssetKey.Set(profile.AssetKey)
		p.Discoverable.Set(profile.Discoverable)

		for devname, netdev := range profile.NetDevs {
			if _, ok := p.NetDevs[devname]; !ok {
				var netdev NetDevEntry
				p.NetDevs[devname] = &netdev
			}

			wwlog.Printf(wwlog.DEBUG, "Updating profile netdev: %s\n", devname)

			p.NetDevs[devname].Device.Set(netdev.Device)
			p.NetDevs[devname].Netmask.Set(netdev.Netmask)
			p.NetDevs[devname].Gateway.Set(netdev.Gateway)
			p.NetDevs[devname].Type.Set(netdev.Type)
			p.NetDevs[devname].OnBoot.Set(netdev.OnBoot)
			p.NetDevs[devname].Primary.Set(netdev.Primary)
			p.NetDevs[devname].Primary.Set(netdev.Default) // backwards compatibility

			// The following should not be set in a profile.
			if netdev.Ipaddr != "" {
				wwlog.Printf(wwlog.WARN, "Ignoring ip address %v in profile %v\n", netdev.Ipaddr, name)
			}
			if netdev.Hwaddr != "" {
				wwlog.Printf(wwlog.WARN, "Ignoring hardware address %v in profile %v\n", netdev.Hwaddr, name)
			}
			p.NetDevs[devname].Tags = make(map[string]*Entry)
			for keyname, key := range netdev.Tags {
				if _, ok := p.Tags[keyname]; !ok {
					var keyVar Entry
					p.NetDevs[devname].Tags[keyname] = &keyVar
				}
				p.NetDevs[devname].Tags[keyname].Set(key)
			}

		}

		// Merge Keys into Tags for backwards compatibility
		if len(profile.Tags) == 0 {
			profile.Tags = make(map[string]string)
		}
		for keyname, key := range profile.Keys {
			profile.Tags[keyname] = key
			delete(profile.Keys, keyname)
		}

		for keyname, key := range profile.Tags {
			if _, ok := p.Tags[keyname]; !ok {
				var key Entry
				p.Tags[keyname] = &key
			}
			p.Tags[keyname].Set(key)
		}

		// TODO: Validate or die on all inputs

		ret = append(ret, p)
	}
	sort.Slice(ret, func(i, j int) bool {
		if ret[i].ClusterName.Get() < ret[j].ClusterName.Get() {
			return true
		} else if ret[i].ClusterName.Get() == ret[j].ClusterName.Get() {
			if ret[i].Id.Get() < ret[j].Id.Get() {
				return true
			}
		}
		return false
	})

	return ret, nil
}

/*
Return the names of all available profiles
*/
func (config *NodeYaml) ListAllProfiles() []string {
	var ret []string
	for name := range config.NodeProfiles {
		ret = append(ret, name)
	}
	return ret
}

func (config *NodeYaml) FindDiscoverableNode() (NodeInfo, string, error) {
	var ret NodeInfo

	nodes, _ := config.FindAllNodes()

	for _, node := range nodes {
		if !node.Discoverable.GetB() {
			continue
		}
		for netdev, dev := range node.NetDevs {
			if !dev.Hwaddr.Defined() {
				return node, netdev, nil
			}
		}
	}

	return ret, "", errors.New("no unconfigured nodes found")
}
