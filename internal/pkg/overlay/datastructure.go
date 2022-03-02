package overlay

import (
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
)

type TemplateStruct struct {
	Id             string
	Hostname       string
	ClusterName    string
	Container      string
	KernelVersion  string
	KernelOverride string
	KernelArgs     string
	Init           string
	Root           string
	IpmiIpaddr     string
	IpmiNetmask    string
	IpmiPort       string
	IpmiGateway    string
	IpmiUserName   string
	IpmiPassword   string
	IpmiInterface  string
	IpmiWrite      string
	RuntimeOverlay string
	SystemOverlay  string
	NetDevs        map[string]*node.NetDevs
	Tags           map[string]string
	Keys           map[string]string
	AllNodes       []node.NodeInfo
	BuildHost      string
	BuildTime      string
	BuildSource    string
	Ipaddr         string
	Netmask        string
	Network        string
	Dhcp           warewulfconf.DhcpConf
	Nfs            warewulfconf.NfsConf
	Warewulf       warewulfconf.WarewulfConf
}
