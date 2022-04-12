package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	"github.com/hpcng/warewulf/cmd/update_configuration/vers42"
	"github.com/hpcng/warewulf/cmd/update_configuration/vers43"
	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"gopkg.in/yaml.v2"
)

var nowrite bool
var quiet bool
var confFile string

const actvers int = 43

type nodeVersionOnly struct {
	WWInternal int `yaml:"WW_INTERNAL"`
}

func saveConf(conf interface{}) {
	out, err := yaml.Marshal(conf)
	if err != nil {
		myprintf("Error in marshal of conf: %s\n", err)
		os.Exit(1)
	}
	if nowrite {
		fmt.Print(string(out))
	} else {
		err = util.CopyFile(confFile, confFile+".bak")
		if err != nil {
			myprintf("Could write file: %s\n", err)
			os.Exit(1)
		}
		info, err := os.Stat(confFile)
		if err != nil {
			myprintf("Could not get file mode: %s\n", err)
			os.Exit(1)
		}
		myprintf("writing configuration file %s as type %s\n", confFile, reflect.TypeOf(conf))
		err = ioutil.WriteFile(confFile, out, info.Mode())
		if err != nil {
			myprintf("Could not write file: %s\n", err)
			os.Exit(1)
		}
	}
}

func printB(x bool) string {
	if x {
		return "true"
	}
	return "false"
}

func update42to43(conf42 vers42.NodeConf) vers43.NodeConf {
	ret := vers43.NodeConf{
		Comment:       conf42.Comment,
		ClusterName:   conf42.ClusterName,
		ContainerName: conf42.ContainerName,
		Init:          conf42.Init,
		Root:          conf42.Root,
		Discoverable:  printB(conf42.Discoverable),
		Profiles:      conf42.Profiles,
		Ipxe:          conf42.Ipxe}
	if conf42.RuntimeOverlay != "" {
		ret.RuntimeOverlay = []string{conf42.RuntimeOverlay}
	}
	if conf42.SystemOverlay != "" {
		ret.SystemOverlay = []string{conf42.SystemOverlay}
	}
	if conf42.KernelArgs != "" || conf42.KernelVersion != "" {
		ret.Kernel = new(vers43.KernelConf)
		ret.Kernel.Override = conf42.KernelVersion
		ret.Kernel.Args = conf42.KernelArgs

	}
	if conf42.IpmiUserName != "" || conf42.IpmiPassword != "" || conf42.IpmiIpaddr != "" ||
		conf42.IpmiNetmask != "" || conf42.IpmiPort != "" || conf42.IpmiGateway != "" ||
		conf42.IpmiInterface != "" {
		ret.Ipmi = new(vers43.IpmiConf)
		ret.Ipmi.UserName = conf42.IpmiUserName
		ret.Ipmi.Password = conf42.IpmiPassword
		ret.Ipmi.Ipaddr = conf42.IpmiIpaddr
		ret.Ipmi.Netmask = conf42.IpmiNetmask
		ret.Ipmi.Port = conf42.IpmiPort
		ret.Ipmi.Gateway = conf42.IpmiGateway
		ret.Ipmi.Interface = conf42.IpmiInterface
	}
	if len(conf42.Keys) != 0 {
		ret.Keys = map[string]string{}
		for k, v := range conf42.Keys {
			ret.Keys[k] = v
		}
	}
	ret.NetDevs = make(map[string]*vers43.NetDevs)
	for devn, netdev := range conf42.NetDevs {
		var device vers43.NetDevs = vers43.NetDevs{
			Type:    netdev.Type,
			Device:  devn,
			Default: printB(netdev.Default),
			Hwaddr:  netdev.Hwaddr,
			Ipaddr:  netdev.Ipaddr,
			IpCIDR:  netdev.IpCIDR,
			Prefix:  netdev.Prefix,
			Netmask: netdev.Netmask,
			Gateway: netdev.Gateway}
		ret.NetDevs[devn] = &device
	}
	return ret
}

func myprintf(format string, a ...interface{}) {
	if !quiet {
		fmt.Printf(format, a...)
	}
}

func main() {
	var endVers int
	var startVers int
	flag.StringVar(&confFile, "f", "", "Config file for update")
	flag.IntVar(&endVers, "e", buildconfig.WWVer, "Final version of configuration file")
	flag.IntVar(&startVers, "s", 0, "Start version  of configuration file, 0  is for autodetection")
	flag.BoolVar(&nowrite, "n", false, "Do not write, just print new conf to terminal")
	flag.BoolVar(&quiet, "q", false, "Do not print what the program is doing")
	flag.Parse()
	if confFile == "" {
		myprintf("No config file given\n!")
		os.Exit(1)
	}
	myprintf("Opening node configuration file: %s\n", confFile)
	data, err := ioutil.ReadFile(confFile)
	if err != nil {
		myprintf("Could open file %v\n", err)
		os.Exit(1)
	}
	var getConf nodeVersionOnly
	myprintf("Unmarshaling the node configuration\n")
	err = yaml.Unmarshal(data, &getConf)
	if err != nil {
		myprintf("Could not unmarshall: %v\n", err)
	}
	myprintf("Got version %v in %s\n", getConf.WWInternal, confFile)
	if getConf.WWInternal == actvers {
		myprintf("On actual version, bailing out\n")
		os.Exit(0)
	}
	if startVers == 0 && getConf.WWInternal == 0 {
		startVers = 42
	}
	var conf42 vers42.NodeYaml
	conf42.NodeProfiles = make(map[string]*vers42.NodeConf)
	conf42.Nodes = make(map[string]*vers42.NodeConf)

	var conf43 vers43.NodeYaml
	conf43.NodeProfiles = make(map[string]*vers43.NodeConf)
	conf43.Nodes = make(map[string]*vers43.NodeConf)
	conf43.WWInternal = 43

	if startVers == 42 {
		myprintf("Unmarshaling the node configuration vers 42\n")
		err = yaml.Unmarshal(data, &conf42)
		if err != nil {
			myprintf("Could not unmarshall version 42: %v\n", err)
		}
		for pname, profile := range conf42.NodeProfiles {
			p43 := update42to43(*profile)
			conf43.NodeProfiles[pname] = &p43
		}
		for nname, node := range conf42.Nodes {
			n43 := update42to43(*node)
			conf43.Nodes[nname] = &n43
		}
		saveConf(conf43)

	}
}
