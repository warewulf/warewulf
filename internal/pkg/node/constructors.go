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
		nodeInfoType := reflect.TypeOf(&n)
		nodeInfoVal := reflect.ValueOf(&n)
		// backward compatibilty
		for keyname, key := range node.Keys {
			node.Tags[keyname] = key
			delete(node.Keys, keyname)
		}
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
				} else if nodeInfoType.Elem().Field(i).Type.Kind() == reflect.Ptr {
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

		for _, p := range n.Profiles {
			if _, ok := config.NodeProfiles[p]; !ok {
				wwlog.Printf(wwlog.WARN, "Profile not found for node '%s': %s\n", nodename, p)
				continue
			}

			wwlog.Printf(wwlog.VERBOSE, "Merging profile into node: %s <- %s\n", nodename, p)

			n.Comment.SetAlt(config.NodeProfiles[p].Comment, p)
			n.ClusterName.SetAlt(config.NodeProfiles[p].ClusterName, p)
			n.ContainerName.SetAlt(config.NodeProfiles[p].ContainerName, p)
			if config.NodeProfiles[p].Kernel != nil {
				n.Kernel.Args.SetAlt(config.NodeProfiles[p].Kernel.Args, p)
			}
			n.Ipxe.SetAlt(config.NodeProfiles[p].Ipxe, p)
			n.Init.SetAlt(config.NodeProfiles[p].Init, p)
			if config.NodeProfiles[p].Ipmi != nil {
				n.Ipmi.Ipaddr.SetAlt(config.NodeProfiles[p].Ipmi.Ipaddr, p)
				n.Ipmi.Netmask.SetAlt(config.NodeProfiles[p].Ipmi.Netmask, p)
				n.Ipmi.Port.SetAlt(config.NodeProfiles[p].Ipmi.Port, p)
				n.Ipmi.Gateway.SetAlt(config.NodeProfiles[p].Ipmi.Gateway, p)
				n.Ipmi.UserName.SetAlt(config.NodeProfiles[p].Ipmi.UserName, p)
				n.Ipmi.Password.SetAlt(config.NodeProfiles[p].Ipmi.Password, p)
				n.Ipmi.Interface.SetAlt(config.NodeProfiles[p].Ipmi.Interface, p)
				n.Ipmi.Write.SetAlt(config.NodeProfiles[p].Ipmi.Write, p)
			}
			n.SystemOverlay.SetAltSlice(config.NodeProfiles[p].SystemOverlay, p)
			n.RuntimeOverlay.SetAltSlice(config.NodeProfiles[p].RuntimeOverlay, p)
			n.Root.SetAlt(config.NodeProfiles[p].Root, p)
			n.AssetKey.SetAlt(config.NodeProfiles[p].AssetKey, p)
			n.Discoverable.SetAlt(config.NodeProfiles[p].Discoverable, p)

			if config.NodeProfiles[p].Kernel != nil {
				if config.NodeProfiles[p].Kernel.Override != "" {
					n.Kernel.Override.SetAlt(config.NodeProfiles[p].Kernel.Override, p)
				} else if config.NodeProfiles[p].Kernel.Version != "" {
					n.Kernel.Override.SetAlt(config.NodeProfiles[p].Kernel.Version, p)
				}
			}

			for devname, netdev := range config.NodeProfiles[p].NetDevs {
				if _, ok := n.NetDevs[devname]; !ok {
					var netdev NetDevEntry
					n.NetDevs[devname] = &netdev
				}
				wwlog.Printf(wwlog.DEBUG, "Updating profile (%s) netdev: %s\n", p, devname)

				n.NetDevs[devname].Device.SetAlt(netdev.Device, p)
				n.NetDevs[devname].Ipaddr.SetAlt(netdev.Ipaddr, p) //FIXME? <- Ipaddr must be uniq
				n.NetDevs[devname].Netmask.SetAlt(netdev.Netmask, p)
				n.NetDevs[devname].Hwaddr.SetAlt(strings.ToLower(netdev.Hwaddr), p)
				n.NetDevs[devname].Gateway.SetAlt(netdev.Gateway, p)
				n.NetDevs[devname].Type.SetAlt(netdev.Type, p)
				n.NetDevs[devname].OnBoot.SetAlt(netdev.OnBoot, p)
				n.NetDevs[devname].Primary.SetAlt(netdev.Primary, p)
				if len(netdev.Tags) != 0 {
					if len(n.NetDevs[devname].Tags) == 0 {
						n.NetDevs[devname].Tags = make(map[string]*Entry)
					}
					for keyname, key := range netdev.Tags {
						if _, ok := n.NetDevs[devname].Tags[keyname]; !ok {
							var keyVar Entry
							n.NetDevs[devname].Tags[keyname] = &keyVar
						}
						n.NetDevs[devname].Tags[keyname].SetAlt(key, p)
					}
				}
			}

			// Merge Keys into Tags for backwards compatibility
			if len(config.NodeProfiles[p].Tags) == 0 {
				config.NodeProfiles[p].Tags = make(map[string]string)
			}
			for keyname, key := range config.NodeProfiles[p].Keys {
				config.NodeProfiles[p].Tags[keyname] = key
				delete(config.NodeProfiles[p].Keys, keyname)
			}

			for keyname, key := range config.NodeProfiles[p].Tags {
				if _, ok := n.Tags[keyname]; !ok {
					var key Entry
					n.Tags[keyname] = &key
				}
				n.Tags[keyname].SetAlt(key, p)
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
