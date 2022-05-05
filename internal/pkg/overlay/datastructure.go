package overlay

import (
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
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
