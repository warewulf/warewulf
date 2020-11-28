package node

import (
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"os"
)

type nodeYaml struct {
	NodeProfiles 	map[string]*ProfileConf
	NodeGroups	 	map[string]*GroupConf
}

type ProfileConf struct {
	Comment 		string `yaml:"comment"`
	Vnfs           	string `yaml:"vnfs,omitempty"`
	Ipxe           	string `yaml:"ipxe template,omitempty"`
	KernelVersion  	string `yaml:"kernel version,omitempty"`
	KernelArgs     	string `yaml:"kernel args,omitempty"`
	IpmiUserName   	string `yaml:"ipmi username,omitempty"`
	IpmiPassword   	string `yaml:"ipmi password,omitempty"`
	DomainName   	string `yaml:"domain name,omitempty"`
	RuntimeOverlay 	string `yaml:"runtime overlay files,omitempty"`
	SystemOverlay  	string `yaml:"system overlay files,omitempty"`
}

type GroupConf struct {
	Comment        	string 	`yaml:"comment"`
	Disabled 		bool 	`yaml:"disabled,omitempty"`
	DomainName   	string 	`yaml:"domain name"`
	Vnfs           	string 	`yaml:"vnfs,omitempty"`
	Ipxe           	string 	`yaml:"ipxe template,omitempty"`
	KernelVersion  	string 	`yaml:"kernel version,omitempty"`
	KernelArgs     	string 	`yaml:"kernel args,omitempty"`
	IpmiUserName   	string 	`yaml:"ipmi username,omitempty"`
	IpmiPassword   	string 	`yaml:"ipmi password,omitempty"`
	RuntimeOverlay 	string 	`yaml:"runtime overlay files,omitempty"`
	SystemOverlay  	string 	`yaml:"system overlay files,omitempty"`
	Profiles   		[]string `yaml:"profiles,omitempty"`
	Nodes          	map[string]*NodeConf
}

type NodeConf struct {
	Comment        	string 	`yaml:"comment,omitempty"`
	Disabled 		bool 	`yaml:"disabled,omitempty"`
	Hostname       	string 	`yaml:"hostname,omitempty"`
	DomainName   	string 	`yaml:"domain name,omitempty"`
	Vnfs           	string 	`yaml:"vnfs,omitempty"`
	Ipxe           	string 	`yaml:"ipxe template,omitempty"`
	KernelVersion  	string 	`yaml:"kernel version,omitempty"`
	KernelArgs     	string 	`yaml:"kernel args,omitempty"`
	IpmiUserName   	string 	`yaml:"ipmi username,omitempty"`
	IpmiPassword   	string 	`yaml:"ipmi password,omitempty"`
	IpmiIpaddr     	string 	`yaml:"ipmi ipaddr,omitempty"`
	RuntimeOverlay 	string 	`yaml:"runtime overlay files,omitempty"`
	SystemOverlay  	string 	`yaml:"system overlay files,omitempty"`
	Profiles   		[]string `yaml:"profiles,omitempty"`
	NetDevs        	map[string]*NetDevs
}

type NetDevs struct {
	Type    		string `yaml:"type,omitempty"`
	Hwaddr  		string
	Ipaddr  		string
	Netmask 		string
	Gateway 		string `yaml:"gateway,omitempty"`
}

type NodeInfoEntry struct {
	value			string
	profile 		string
	group 			string
	def		 		string
}

type NodeInfo struct {
	Id             	NodeInfoEntry
	Gid 	       	NodeInfoEntry
	Uid 		   	NodeInfoEntry
	Comment 		NodeInfoEntry
	GroupName      	NodeInfoEntry
	HostName       	NodeInfoEntry
	Fqdn           	NodeInfoEntry
	DomainName     	NodeInfoEntry
	Vnfs           	NodeInfoEntry
	Ipxe           	NodeInfoEntry
	KernelVersion  	NodeInfoEntry
	KernelArgs     	NodeInfoEntry
	IpmiIpaddr     	NodeInfoEntry
	IpmiUserName   	NodeInfoEntry
	IpmiPassword   	NodeInfoEntry
	RuntimeOverlay 	NodeInfoEntry
	SystemOverlay  	NodeInfoEntry
	Profiles   		[]string
	GroupProfiles   []string
	NetDevs        	map[string]*NetDevs
}

type GroupInfo struct {
	Id             	string
	Comment        	string
	Vnfs           	string
	Ipxe           	string
	KernelVersion  	string
	KernelArgs     	string
	IpmiUserName   	string
	IpmiPassword   	string
	DomainName   	string
	RuntimeOverlay 	string
	SystemOverlay  	string
	Profiles   		[]string
}

type ProfileInfo struct {
	Id 				string
	Comment 		string
	Vnfs           	string
	Ipxe           	string
	KernelVersion  	string
	KernelArgs     	string
	IpmiUserName   	string
	IpmiPassword   	string
	DomainName   	string
	RuntimeOverlay 	string
	SystemOverlay  	string
}



const ConfigFile = "/etc/warewulf/nodes.conf"

func init() {
	//TODO: Check to make sure nodes.conf is found
	if util.IsFile(ConfigFile) == false {
		wwlog.Printf(wwlog.ERROR, "Configuration file not found: %s\n", ConfigFile)
		os.Exit(1)
	}
}
