package upgrade

import (
	"gopkg.in/yaml.v3"
)

func Parse(data []byte) (nodeYaml NodeYaml, err error) {
	if err = yaml.Unmarshal(data, &nodeYaml); err != nil {
		return nodeYaml, err
	}
	if nodeYaml.Nodes == nil {
		nodeYaml.Nodes = map[string]*Node{}
	}
	if nodeYaml.NodeProfiles == nil {
		nodeYaml.NodeProfiles = map[string]*Profile{}
	}
	return nodeYaml, nil
}

type NodeYaml struct {
	NodeProfiles map[string]*Profile
	Nodes        map[string]*Node
}

type Node struct {
	Profile `yaml:"-,inline"`
}

type Profile struct {
	AssetKey       string                 `yaml:"asset key,omitempty"`
	ClusterName    string                 `yaml:"cluster name,omitempty"`
	Comment        string                 `yaml:"comment,omitempty"`
	ContainerName  string                 `yaml:"container name,omitempty"`
	Disabled       string                 `yaml:"disabled,omitempty"`
	Discoverable   string                 `yaml:"discoverable,omitempty"`
	Disks          map[string]*Disk       `yaml:"disks,omitempty"`
	FileSystems    map[string]*FileSystem `yaml:"filesystems,omitempty"`
	Init           string                 `yaml:"init,omitempty"`
	Ipmi           *IpmiConf              `yaml:"ipmi,omitempty"`
	IpmiEscapeChar string                 `yaml:"ipmi escapechar,omitempty"`
	IpmiGateway    string                 `yaml:"ipmi gateway,omitempty"`
	IpmiInterface  string                 `yaml:"ipmi interface,omitempty"`
	IpmiIpaddr     string                 `yaml:"ipmi ipaddr,omitempty"`
	IpmiNetmask    string                 `yaml:"ipmi netmask,omitempty"`
	IpmiPassword   string                 `yaml:"ipmi password,omitempty"`
	IpmiPort       string                 `yaml:"ipmi port,omitempty"`
	IpmiUserName   string                 `yaml:"ipmi username,omitempty"`
	IpmiWrite      string                 `yaml:"ipmi write,omitempty"`
	Ipxe           string                 `yaml:"ipxe template,omitempty"`
	Kernel         *KernelConf            `yaml:"kernel,omitempty"`
	KernelArgs     string                 `yaml:"kernel args,omitempty"`
	KernelOverride string                 `yaml:"kernel override,omitempty"`
	KernelVersion  string                 `yaml:"kernel version,omitempty"`
	Keys           map[string]string      `yaml:"keys,omitempty"`
	NetDevs        map[string]*NetDev     `yaml:"network devices,omitempty"`
	PrimaryNetDev  string                 `yaml:"primary network,omitempty"`
	Profiles       []string               `yaml:"profiles,omitempty"`
	Root           string                 `yaml:"root,omitempty"`
	RuntimeOverlay []string               `yaml:"runtime overlay,omitempty"`
	SystemOverlay  []string               `yaml:"system overlay,omitempty"`
	Tags           map[string]string      `yaml:"tags,omitempty"`
	TagsDel        []string               `yaml:"tagsdel,omitempty"`
}

type IpmiConf struct {
	EscapeChar string            `yaml:"escapechar,omitempty"`
	Gateway    string            `yaml:"gateway,omitempty"`
	Interface  string            `yaml:"interface,omitempty"`
	Ipaddr     string            `yaml:"ipaddr,omitempty"`
	Netmask    string            `yaml:"netmask,omitempty"`
	Password   string            `yaml:"password,omitempty"`
	Port       string            `yaml:"port,omitempty"`
	Tags       map[string]string `yaml:"tags,omitempty"`
	TagsDel    []string          `yaml:"tagsdel,omitempty"`
	UserName   string            `yaml:"username,omitempty"`
	Write      string            `yaml:"write,omitempty"`
}

type KernelConf struct {
	Args     string `yaml:"args,omitempty"`
	Override string `yaml:"override,omitempty"`
	Version  string `yaml:"version,omitempty"`
}

type NetDev struct {
	Default string            `yaml:"default"`
	Device  string            `yaml:"device,omitempty"`
	Gateway string            `yaml:"gateway,omitempty"`
	Hwaddr  string            `yaml:"hwaddr,omitempty"`
	IpCIDR  string            `yaml:"ipcidr,omitempty"`
	Ipaddr  string            `yaml:"ipaddr,omitempty"`
	Ipaddr6 string            `yaml:"ip6addr,omitempty"`
	MTU     string            `yaml:"mtu,omitempty"`
	Netmask string            `yaml:"netmask,omitempty"`
	OnBoot  string            `yaml:"onboot,omitempty"`
	Prefix  string            `yaml:"prefix,omitempty"`
	Primary string            `yaml:"primary,omitempty"`
	Tags    map[string]string `yaml:"tags,omitempty"`
	TagsDel []string          `yaml:"tagsdel,omitempty"`
	Type    string            `yaml:"type,omitempty"`
}

type Disk struct {
	Partitions map[string]*Partition `yaml:"partitions,omitempty"`
	WipeTable  string                `yaml:"wipe_table,omitempty"`
}

type Partition struct {
	Guid               string `yaml:"guid,omitempty"`
	Number             string `yaml:"number,omitempty"`
	Resize             string `yaml:"resize,omitempty"`
	ShouldExist        string `yaml:"should_exist,omitempty"`
	SizeMiB            string `yaml:"size_mib,omitempty"`
	StartMiB           string `yaml:"start_mib,omitempty"`
	TypeGuid           string `yaml:"type_guid,omitempty"`
	WipePartitionEntry string `yaml:"wipe_partition_entry,omitempty"`
}

type FileSystem struct {
	Format       string `yaml:"format,omitempty"`
	Label        string `yaml:"label,omitempty"`
	MountOptions string `yaml:"mount_options,omitempty"`
	//MountOptions   []string `yaml:"mount_options,omitempty"`
	Options        []string `yaml:"options,omitempty"`
	Path           string   `yaml:"path,omitempty"`
	Uuid           string   `yaml:"uuid,omitempty"`
	WipeFileSystem string   `yaml:"wipe_filesystem,omitempty"`
}
