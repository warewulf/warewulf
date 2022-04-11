package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hpcng/warewulf/cmd/update_configuration/vers42"
	"github.com/hpcng/warewulf/cmd/update_configuration/vers43"
	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"gopkg.in/yaml.v2"
)

var nowrite bool
var confFile string

type nodeVersionOnly struct {
	WWInternal int `yaml:"WW_INTERNAL"`
}

func saveConf(conf interface{}) {
	out, err := yaml.Marshal(conf)
	if err != nil {
		fmt.Printf("Error in marshal of conf: %s\n", err)
		os.Exit(1)
	}
	if nowrite {
		fmt.Println(string(out))
	} else {
		err = util.CopyFile(confFile, confFile+".bak")
		if err != nil {
			fmt.Printf("Could write file: %s\n", err)
			os.Exit(1)
		}
		info, err := os.Stat(confFile)
		if err != nil {
			fmt.Printf("Could not get file mode: %s\n", err)
			os.Exit(1)
		}
		err = ioutil.WriteFile(confFile, out, info.Mode())
	}
}

func printB(x bool) string {
	if x {
		return "true"
	}
	return "false"
}

func main() {
	var endVers int
	var startVers int
	flag.StringVar(&confFile, "f", "", "Config file for update")
	flag.IntVar(&endVers, "e", buildconfig.WWVer, "Final version of configuration file")
	flag.IntVar(&startVers, "s", 0, "Start version  of configuration file, 0  is for autodetection")
	flag.BoolVar(&nowrite, "n", false, "Do not write, just print new conf to terminal")
	flag.Parse()
	if confFile == "" {
		fmt.Printf("No config file given\n!")
		os.Exit(1)
	}
	fmt.Printf("Opening node configuration file: %s\n", confFile)
	data, err := ioutil.ReadFile(confFile)
	if err != nil {
		fmt.Printf("Could open file %v\n", err)
		os.Exit(1)
	}
	if startVers == 0 {
		var getConf nodeVersionOnly
		fmt.Printf("Unmarshaling the node configuration\n")
		err = yaml.Unmarshal(data, &getConf)
		if err != nil {
			fmt.Printf("Could not unmarshall: %v\n", err)
		}
		fmt.Printf("Got version %v in %s\n", getConf.WWInternal, confFile)
		if getConf.WWInternal == 0 {
			startVers = 42
		}
	}
	var conf42 vers42.NodeYaml
	conf42.NodeProfiles = make(map[string]*vers42.NodeConf)
	conf42.Nodes = make(map[string]*vers42.NodeConf)

	var conf43 vers43.NodeYaml
	conf43.NodeProfiles = make(map[string]*vers43.NodeConf)
	conf43.Nodes = make(map[string]*vers43.NodeConf)
	conf43.WWInternal = 43

	if startVers == 42 {
		fmt.Printf("Unmarshaling the node configuration vers 42\n")
		err = yaml.Unmarshal(data, &conf42)
		if err != nil {
			fmt.Printf("Could not unmarshall version 42: %v\n", err)
		}
		for pname, profile := range conf42.NodeProfiles {
			profileConf := vers43.NodeConf{
				Comment:       profile.Comment,
				ClusterName:   profile.ClusterName,
				ContainerName: profile.ContainerName,
				Init:          profile.Init,
				Root:          profile.Root,
				Discoverable:  printB(profile.Discoverable),
				Profiles:      profile.Profiles,
				Ipxe:          profile.Ipxe}
			conf43.NodeProfiles[pname] = &profileConf
			if profile.RuntimeOverlay != "" {
				conf43.NodeProfiles[pname].RuntimeOverlay = []string{profile.RuntimeOverlay}
			}
			if profile.SystemOverlay != "" {
				conf43.NodeProfiles[pname].SystemOverlay = []string{profile.SystemOverlay}
			}
			if profile.KernelArgs != "" || profile.KernelVersion != "" {
				conf43.NodeProfiles[pname].Kernel = new(vers43.KernelConf)
				conf43.NodeProfiles[pname].Kernel.Override = profile.KernelVersion
				conf43.NodeProfiles[pname].Kernel.Args = profile.KernelArgs

			}
			if profile.IpmiUserName != "" || profile.IpmiPassword != "" || profile.IpmiIpaddr != "" ||
				profile.IpmiNetmask != "" || profile.IpmiPort != "" || profile.IpmiGateway != "" ||
				profile.IpmiInterface != "" {
				conf43.NodeProfiles[pname].Ipmi = new(vers43.IpmiConf)
				conf43.NodeProfiles[pname].Ipmi.UserName = profile.IpmiUserName
				conf43.NodeProfiles[pname].Ipmi.Password = profile.IpmiPassword
				conf43.NodeProfiles[pname].Ipmi.Ipaddr = profile.IpmiIpaddr
				conf43.NodeProfiles[pname].Ipmi.Netmask = profile.IpmiNetmask
				conf43.NodeProfiles[pname].Ipmi.Port = profile.IpmiPort
				conf43.NodeProfiles[pname].Ipmi.Gateway = profile.IpmiGateway
				conf43.NodeProfiles[pname].Ipmi.Interface = profile.IpmiInterface
			}
			if len(profile.Keys) != 0 {
				conf43.NodeProfiles[pname].Keys = map[string]string{}
				for k, v := range profile.Keys {
					conf43.NodeProfiles[pname].Keys[k] = v
				}
			}
			conf43.NodeProfiles[pname].NetDevs = make(map[string]*vers43.NetDevs)
			for devn, netdev := range profile.NetDevs {
				var device vers43.NetDevs = vers43.NetDevs{
					Type:    netdev.Type,
					Default: printB(netdev.Default),
					Hwaddr:  netdev.Hwaddr,
					Ipaddr:  netdev.Ipaddr,
					IpCIDR:  netdev.IpCIDR,
					Prefix:  netdev.Prefix,
					Netmask: netdev.Netmask,
					Gateway: netdev.Gateway}
				conf43.NodeProfiles[pname].NetDevs[devn] = &device
			}

		}
		for nname, node := range conf42.Nodes {
			nodeConf := vers43.NodeConf{
				Comment:       node.Comment,
				ClusterName:   node.ClusterName,
				ContainerName: node.ContainerName,
				Init:          node.Init,
				Root:          node.Root,
				Discoverable:  printB(node.Discoverable),
				Profiles:      node.Profiles,
				Ipxe:          node.Ipxe}
			conf43.Nodes[nname] = &nodeConf
			if node.RuntimeOverlay != "" {
				conf43.Nodes[nname].RuntimeOverlay = []string{node.RuntimeOverlay}
			}
			if node.SystemOverlay != "" {
				conf43.Nodes[nname].SystemOverlay = []string{node.SystemOverlay}
			}
			if node.KernelArgs != "" || node.KernelVersion != "" {
				conf43.Nodes[nname].Kernel = new(vers43.KernelConf)
				conf43.Nodes[nname].Kernel.Override = node.KernelVersion
				conf43.Nodes[nname].Kernel.Args = node.KernelArgs

			}
			if node.IpmiUserName != "" || node.IpmiPassword != "" || node.IpmiIpaddr != "" ||
				node.IpmiNetmask != "" || node.IpmiPort != "" || node.IpmiGateway != "" ||
				node.IpmiInterface != "" {
				conf43.Nodes[nname].Ipmi = new(vers43.IpmiConf)
				conf43.Nodes[nname].Ipmi.UserName = node.IpmiUserName
				conf43.Nodes[nname].Ipmi.Password = node.IpmiPassword
				conf43.Nodes[nname].Ipmi.Ipaddr = node.IpmiIpaddr
				conf43.Nodes[nname].Ipmi.Netmask = node.IpmiNetmask
				conf43.Nodes[nname].Ipmi.Port = node.IpmiPort
				conf43.Nodes[nname].Ipmi.Gateway = node.IpmiGateway
				conf43.Nodes[nname].Ipmi.Interface = node.IpmiInterface
			}
			if len(node.Keys) != 0 {
				conf43.Nodes[nname].Keys = map[string]string{}
				for k, v := range node.Keys {
					conf43.Nodes[nname].Keys[k] = v
				}
			}
			conf43.Nodes[nname].NetDevs = make(map[string]*vers43.NetDevs)
			for devn, netdev := range node.NetDevs {
				var device vers43.NetDevs = vers43.NetDevs{
					Type:    netdev.Type,
					Default: printB(netdev.Default),
					Hwaddr:  netdev.Hwaddr,
					Ipaddr:  netdev.Ipaddr,
					IpCIDR:  netdev.IpCIDR,
					Prefix:  netdev.Prefix,
					Netmask: netdev.Netmask,
					Gateway: netdev.Gateway}
				conf43.Nodes[nname].NetDevs[devn] = &device
			}
		}
		saveConf(conf43)

	}
}
