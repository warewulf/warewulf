package overlay

import (
	"net"
	"os"
	"strconv"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

/*
struct which contains the variables to which are available in
the templates.
*/
type TemplateStruct struct {
	Id             string
	Hostname       string
	ClusterName    string
	Container      string
	Kernel         *node.KernelConf
	Init           string
	Root           string
	Ipmi           *node.IpmiConf
	RuntimeOverlay string
	SystemOverlay  string
	NetDevs        map[string]*node.NetDevs
	Tags           map[string]string
	Keys           map[string]string
	AllNodes       []node.NodeInfo
	BuildHost      string
	BuildTime      string
	BuildTimeUnix  string
	BuildSource    string
	Ipaddr         string
	Ipaddr6        string
	Netmask        string
	Network        string
	NetworkCIDR    string
	Ipv6           bool
	Dhcp           warewulfconf.DhcpConf
	Nfs            warewulfconf.NfsConf
	Warewulf       warewulfconf.WarewulfConf
}

/*
Initialize an TemplateStruct with the given node.NodeInfo
*/
func InitStruct(nodeInfo node.NodeInfo) TemplateStruct {
	var tstruct TemplateStruct
	controller, err := warewulfconf.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}
	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}
	allNodes, err := nodeDB.FindAllNodes()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	tstruct.Kernel = new(node.KernelConf)
	tstruct.Ipmi = new(node.IpmiConf)
	tstruct.Id = nodeInfo.Id.Get()
	tstruct.Hostname = nodeInfo.Id.Get()
	tstruct.Id = nodeInfo.Id.Get()
	tstruct.Hostname = nodeInfo.Id.Get()
	tstruct.ClusterName = nodeInfo.ClusterName.Get()
	tstruct.Container = nodeInfo.ContainerName.Get()
	tstruct.Kernel.Version = nodeInfo.Kernel.Override.Get()
	tstruct.Kernel.Override = nodeInfo.Kernel.Override.Get()
	tstruct.Kernel.Args = nodeInfo.Kernel.Args.Get()
	tstruct.Init = nodeInfo.Init.Get()
	tstruct.Root = nodeInfo.Root.Get()
	tstruct.Ipmi.Ipaddr = nodeInfo.Ipmi.Ipaddr.Get()
	tstruct.Ipmi.Netmask = nodeInfo.Ipmi.Netmask.Get()
	tstruct.Ipmi.Port = nodeInfo.Ipmi.Port.Get()
	tstruct.Ipmi.Gateway = nodeInfo.Ipmi.Gateway.Get()
	tstruct.Ipmi.UserName = nodeInfo.Ipmi.UserName.Get()
	tstruct.Ipmi.Password = nodeInfo.Ipmi.Password.Get()
	tstruct.Ipmi.Interface = nodeInfo.Ipmi.Interface.Get()
	tstruct.Ipmi.Write = nodeInfo.Ipmi.Write.GetB()
	tstruct.RuntimeOverlay = nodeInfo.RuntimeOverlay.Print()
	tstruct.SystemOverlay = nodeInfo.SystemOverlay.Print()
	tstruct.NetDevs = make(map[string]*node.NetDevs)
	tstruct.Keys = make(map[string]string)
	tstruct.Tags = make(map[string]string)
	for devname, netdev := range nodeInfo.NetDevs {
		var nd node.NetDevs
		tstruct.NetDevs[devname] = &nd
		tstruct.NetDevs[devname].Device = netdev.Device.Get()
		tstruct.NetDevs[devname].Hwaddr = netdev.Hwaddr.Get()
		tstruct.NetDevs[devname].Ipaddr = netdev.Ipaddr.Get()
		tstruct.NetDevs[devname].Netmask = netdev.Netmask.Get()
		tstruct.NetDevs[devname].Gateway = netdev.Gateway.Get()
		tstruct.NetDevs[devname].Type = netdev.Type.Get()
		tstruct.NetDevs[devname].OnBoot = netdev.OnBoot.Get()
		tstruct.NetDevs[devname].Primary = netdev.Primary.Get()
		mask := net.IPMask(net.ParseIP(netdev.Netmask.Get()).To4())
		ipaddr := net.ParseIP(netdev.Ipaddr.Get()).To4()
		netaddr := net.IPNet{IP: ipaddr, Mask: mask}
		netPrefix, _ := net.IPMask(net.ParseIP(netdev.Netmask.Get()).To4()).Size()
		tstruct.NetDevs[devname].Prefix = strconv.Itoa(netPrefix)
		tstruct.NetDevs[devname].IpCIDR = netaddr.String()
		tstruct.NetDevs[devname].Ipaddr6 = netdev.Ipaddr6.Get()
		tstruct.NetDevs[devname].Tags = make(map[string]string)
		for key, value := range netdev.Tags {
			tstruct.NetDevs[devname].Tags[key] = value.Get()
		}
	}
	// Backwards compatibility for templates using "Keys"
	for keyname, key := range nodeInfo.Tags {
		tstruct.Keys[keyname] = key.Get()
	}
	for keyname, key := range nodeInfo.Tags {
		tstruct.Tags[keyname] = key.Get()
	}
	tstruct.AllNodes = allNodes
	tstruct.Nfs = *controller.Nfs
	tstruct.Dhcp = *controller.Dhcp
	tstruct.Warewulf = *controller.Warewulf
	tstruct.Ipaddr = controller.Ipaddr
	tstruct.Ipaddr6 = controller.Ipaddr6
	tstruct.Netmask = controller.Netmask
	tstruct.Network = controller.Network
	netaddrStruct := net.IPNet{IP: net.ParseIP(controller.Network), Mask: net.IPMask(net.ParseIP(controller.Netmask))}
	tstruct.NetworkCIDR = netaddrStruct.String()
	if controller.Ipaddr6 != "" {
		tstruct.Ipv6 = true
	} else {
		tstruct.Ipv6 = false
	}
	hostname, _ := os.Hostname()
	tstruct.BuildHost = hostname
	dt := time.Now()
	tstruct.BuildTime = dt.Format("01-02-2006 15:04:05 MST")
	tstruct.BuildTimeUnix = strconv.FormatInt(dt.Unix(), 10)

	return tstruct

}
