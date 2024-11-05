package upgrade

import (
	"net"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

var wwinitSplitOverlays = []string{
	"wwinit",
	"wwclient",
	"fstab",
	"hostname",
	"ssh.host_keys",
	"issue",
	"resolv",
	"udev.netname",
	"systemd.netname",
	"ifcfg",
	"NetworkManager",
	"debian.interfaces",
	"wicked",
	"ignition",
}

var genericSplitOverlays = []string{
	"hosts",
	"ssh.authorized_keys",
	"syncuser",
}

func indexOf[T comparable](slice []T, item T) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}

func replaceSliceElement[T any](original []T, index int, replacement []T) []T {
	if index < 0 || index >= len(original) {
		return original
	}
	return append(original[:index], append(replacement, original[index+1:]...)...)
}

func logIgnore(name string, value interface{}, reason string) {
	wwlog.Warn("ignore: %s: %v (%s)", name, value, reason)
}

func warnError(err error) {
	if err != nil {
		wwlog.Warn("%s", err)
	}
}

func Parse(data []byte) (nodesYaml *NodesYaml, err error) {
	nodesYaml = new(NodesYaml)
	if err = yaml.Unmarshal(data, nodesYaml); err != nil {
		return nodesYaml, err
	}
	return nodesYaml, nil
}

type NodesYaml struct {
	WWInternal   string `yaml:"WW_INTERNAL"`
	NodeProfiles map[string]*Profile
	Nodes        map[string]*Node
}

func (this *NodesYaml) Upgrade(addDefaults bool, replaceOverlays bool) (upgraded *node.NodesYaml) {
	upgraded = new(node.NodesYaml)
	upgraded.NodeProfiles = make(map[string]*node.Profile)
	upgraded.Nodes = make(map[string]*node.Node)
	if this.WWInternal != "" {
		logIgnore("WW_INTERNAL", this.WWInternal, "obsolete")
	}
	for name, profile := range this.NodeProfiles {
		upgraded.NodeProfiles[name] = profile.Upgrade(addDefaults, replaceOverlays)
	}
	for name, node := range this.Nodes {
		upgraded.Nodes[name] = node.Upgrade(addDefaults, replaceOverlays)
		if addDefaults && !util.InSlice(upgraded.Nodes[name].Profiles, "default") {
			wwlog.Warn("node %s does not include the default profile: verify default settings manually", name)
		}
	}
	if addDefaults {
		if _, ok := upgraded.NodeProfiles["default"]; !ok {
			upgraded.NodeProfiles["default"] = new(node.Profile)
			upgraded.NodeProfiles["default"].Kernel = new(node.KernelConf)
		}
		defaultProfile := upgraded.NodeProfiles["default"]
		if len(defaultProfile.SystemOverlay) == 0 {
			defaultProfile.SystemOverlay = append(
				defaultProfile.SystemOverlay, wwinitSplitOverlays...)
		}
		if len(defaultProfile.RuntimeOverlay) == 0 {
			defaultProfile.RuntimeOverlay = append(
				defaultProfile.RuntimeOverlay, genericSplitOverlays...)
		}
		if defaultProfile.Kernel.Args == "" {
			defaultProfile.Kernel.Args = "quiet crashkernel=no vga=791 net.naming-scheme=v238"
		}
		if defaultProfile.Init == "" {
			defaultProfile.Init = "/sbin/init"
		}
		if defaultProfile.Root == "" {
			defaultProfile.Root = "initramfs"
		}
		if defaultProfile.Ipxe == "" {
			defaultProfile.Ipxe = "default"
		}
	}
	return upgraded
}

type Node struct {
	Profile `yaml:"-,inline"`
}

func (this *Node) Upgrade(addDefaults bool, replaceOverlays bool) (upgraded *node.Node) {
	upgraded = new(node.Node)
	upgraded.Tags = make(map[string]string)
	upgraded.Disks = make(map[string]*node.Disk)
	upgraded.FileSystems = make(map[string]*node.FileSystem)
	upgraded.Ipmi = new(node.IpmiConf)
	upgraded.Kernel = new(node.KernelConf)
	upgraded.NetDevs = make(map[string]*node.NetDev)
	upgraded.AssetKey = this.AssetKey
	upgraded.ClusterName = this.ClusterName
	upgraded.Comment = this.Comment
	upgraded.ContainerName = this.ContainerName
	if this.Disabled != "" {
		logIgnore("Disabled", this.Disabled, "obsolete")
	}
	if this.Discoverable != "" {
		warnError(upgraded.Discoverable.Set(this.Discoverable))
	}
	if this.Disks != nil {
		for name, disk := range this.Disks {
			upgraded.Disks[name] = disk.Upgrade()
		}
	}
	if this.FileSystems != nil {
		for name, fileSystem := range this.FileSystems {
			upgraded.FileSystems[name] = fileSystem.Upgrade()
		}
	}
	upgraded.Init = this.Init
	if this.Ipmi != nil {
		upgraded.Ipmi = this.Ipmi.Upgrade()
	} else {
		upgraded.Ipmi = new(node.IpmiConf)
	}
	if upgraded.Ipmi.EscapeChar == "" {
		upgraded.Ipmi.EscapeChar = this.IpmiEscapeChar
	}
	if upgraded.Ipmi.Gateway.Equal(net.IP{}) {
		upgraded.Ipmi.Gateway = net.ParseIP(this.IpmiGateway)
	}
	if upgraded.Ipmi.Interface == "" {
		upgraded.Ipmi.Interface = this.IpmiInterface
	}
	if upgraded.Ipmi.Ipaddr.Equal(net.IP{}) {
		upgraded.Ipmi.Ipaddr = net.ParseIP(this.IpmiIpaddr)
	}
	if upgraded.Ipmi.Netmask.Equal(net.IP{}) {
		upgraded.Ipmi.Netmask = net.ParseIP(this.IpmiNetmask)
	}
	if upgraded.Ipmi.Password == "" {
		upgraded.Ipmi.Password = this.IpmiPassword
	}
	if upgraded.Ipmi.Port == "" {
		upgraded.Ipmi.Port = this.IpmiPort
	}
	if upgraded.Ipmi.UserName == "" {
		upgraded.Ipmi.UserName = this.IpmiUserName
	}
	if upgraded.Ipmi.Write == "" && this.IpmiWrite != "" {
		warnError(upgraded.Ipmi.Write.Set(this.IpmiWrite))
	}
	upgraded.Ipxe = this.Ipxe
	if this.Kernel != nil {
		upgraded.Kernel = this.Kernel.Upgrade()
	} else {
		upgraded.Kernel = new(node.KernelConf)
	}
	if upgraded.Kernel.Args == "" {
		upgraded.Kernel.Args = this.KernelArgs
	}
	if upgraded.Kernel.Override == "" {
		upgraded.Kernel.Override = this.KernelOverride
	}
	if upgraded.Kernel.Version == "" {
		upgraded.Kernel.Version = this.KernelVersion
	}
	if this.Keys != nil {
		for key, value := range this.Keys {
			upgraded.Tags[key] = value
		}
	}
	if this.NetDevs != nil {
		for name, netDev := range this.NetDevs {
			upgraded.NetDevs[name] = netDev.Upgrade(addDefaults)
		}
	}
	if this.PrimaryNetDev != "" {
		upgraded.PrimaryNetDev = this.PrimaryNetDev
	} else {
		for name, netDev := range this.NetDevs {
			if b, _ := strconv.ParseBool(netDev.Primary); b {
				upgraded.PrimaryNetDev = name
				break
			} else if b, _ := strconv.ParseBool(netDev.Default); b {
				upgraded.PrimaryNetDev = name
				break
			}
		}
	}
	upgraded.Profiles = append(upgraded.Profiles, this.Profiles...)
	if addDefaults {
		if len(upgraded.Profiles) == 0 {
			upgraded.Profiles = append(upgraded.Profiles, "default")
		}
	}
	upgraded.Root = this.Root
	if this.RuntimeOverlay != nil {
		switch overlay := this.RuntimeOverlay.(type) {
		case string:
			upgraded.RuntimeOverlay = append(upgraded.RuntimeOverlay, strings.Split(overlay, ",")...)
		case []interface{}:
			for _, each := range overlay {
				upgraded.RuntimeOverlay = append(upgraded.RuntimeOverlay, each.(string))
			}
		default:
			wwlog.Error("unparsable RuntimeOverlay: %v", overlay)
		}
	}
	if this.SystemOverlay != nil {
		switch overlay := this.SystemOverlay.(type) {
		case string:
			upgraded.SystemOverlay = append(upgraded.SystemOverlay, strings.Split(overlay, ",")...)
		case []interface{}:
			for _, each := range overlay {
				upgraded.SystemOverlay = append(upgraded.SystemOverlay, each.(string))
			}
		default:
			wwlog.Error("unparsable SystemOverlay: %v", overlay)
		}
	}
	if replaceOverlays {
		if indexOf(upgraded.SystemOverlay, "wwinit") != -1 {
			upgraded.SystemOverlay = replaceSliceElement(
				upgraded.SystemOverlay,
				indexOf(upgraded.SystemOverlay, "wwinit"),
				wwinitSplitOverlays)
		}
		if indexOf(upgraded.RuntimeOverlay, "generic") != -1 {
			upgraded.RuntimeOverlay = replaceSliceElement(
				upgraded.RuntimeOverlay,
				indexOf(upgraded.RuntimeOverlay, "generic"),
				genericSplitOverlays)
		}
	}
	if this.Tags != nil {
		for key, value := range this.Tags {
			upgraded.Tags[key] = value
		}
	}
	for _, tag := range this.TagsDel {
		delete(upgraded.Tags, tag)
	}
	return
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
	RuntimeOverlay interface{}            `yaml:"runtime overlay,omitempty"`
	SystemOverlay  interface{}            `yaml:"system overlay,omitempty"`
	Tags           map[string]string      `yaml:"tags,omitempty"`
	TagsDel        []string               `yaml:"tagsdel,omitempty"`
}

func (this *Profile) Upgrade(addDefaults bool, replaceOverlays bool) (upgraded *node.Profile) {
	upgraded = new(node.Profile)
	upgraded.Tags = make(map[string]string)
	upgraded.Disks = make(map[string]*node.Disk)
	upgraded.FileSystems = make(map[string]*node.FileSystem)
	upgraded.Kernel = new(node.KernelConf)
	upgraded.NetDevs = make(map[string]*node.NetDev)
	if this.AssetKey != "" {
		logIgnore("AssetKey", this.AssetKey, "invalid for profiles")
	}
	upgraded.ClusterName = this.ClusterName
	upgraded.Comment = this.Comment
	upgraded.ContainerName = this.ContainerName
	if this.Disabled != "" {
		logIgnore("Disabled", this.Disabled, "obsolete")
	}
	if this.Discoverable != "" {
		logIgnore("Discoverable", this.Discoverable, "invalid for profiles")
	}
	if this.Disks != nil {
		for name, disk := range this.Disks {
			upgraded.Disks[name] = disk.Upgrade()
		}
	}
	if this.FileSystems != nil {
		for name, fileSystem := range this.FileSystems {
			upgraded.FileSystems[name] = fileSystem.Upgrade()
		}
	}
	upgraded.Init = this.Init
	upgraded.Ipmi = new(node.IpmiConf)
	if this.Ipmi != nil {
		upgraded.Ipmi = this.Ipmi.Upgrade()
	} else {
		upgraded.Ipmi = new(node.IpmiConf)
	}
	if upgraded.Ipmi.EscapeChar == "" {
		upgraded.Ipmi.EscapeChar = this.IpmiEscapeChar
	}
	if upgraded.Ipmi.Gateway.Equal(net.IP{}) {
		upgraded.Ipmi.Gateway = net.ParseIP(this.IpmiGateway)
	}
	if upgraded.Ipmi.Interface == "" {
		upgraded.Ipmi.Interface = this.IpmiInterface
	}
	if upgraded.Ipmi.Ipaddr.Equal(net.IP{}) {
		upgraded.Ipmi.Ipaddr = net.ParseIP(this.IpmiIpaddr)
	}
	if upgraded.Ipmi.Netmask.Equal(net.IP{}) {
		upgraded.Ipmi.Netmask = net.ParseIP(this.IpmiNetmask)
	}
	if upgraded.Ipmi.Password == "" {
		upgraded.Ipmi.Password = this.IpmiPassword
	}
	if upgraded.Ipmi.Port == "" {
		upgraded.Ipmi.Port = this.IpmiPort
	}
	if upgraded.Ipmi.UserName == "" {
		upgraded.Ipmi.UserName = this.IpmiUserName
	}
	if upgraded.Ipmi.Write == "" && this.IpmiWrite != "" {
		warnError(upgraded.Ipmi.Write.Set(this.IpmiWrite))
	}
	upgraded.Ipxe = this.Ipxe
	if this.Kernel != nil {
		upgraded.Kernel = this.Kernel.Upgrade()
	} else {
		upgraded.Kernel = new(node.KernelConf)
	}
	if upgraded.Kernel.Args == "" {
		upgraded.Kernel.Args = this.KernelArgs
	}
	if upgraded.Kernel.Override == "" {
		upgraded.Kernel.Override = this.KernelOverride
	}
	if upgraded.Kernel.Version == "" {
		upgraded.Kernel.Version = this.KernelVersion
	}
	if this.Keys != nil {
		for key, value := range this.Keys {
			upgraded.Tags[key] = value
		}
	}
	if this.NetDevs != nil {
		for name, netDev := range this.NetDevs {
			upgraded.NetDevs[name] = netDev.Upgrade(addDefaults)
		}
	}
	if this.PrimaryNetDev != "" {
		upgraded.PrimaryNetDev = this.PrimaryNetDev
	} else {
		for name, netDev := range this.NetDevs {
			if b, _ := strconv.ParseBool(netDev.Primary); b {
				upgraded.PrimaryNetDev = name
				break
			} else if b, _ := strconv.ParseBool(netDev.Default); b {
				upgraded.PrimaryNetDev = name
				break
			}
		}
	}
	if this.Profiles != nil {
		logIgnore("Profiles", this.Profiles, "invalid for profiles")
	}
	upgraded.Root = this.Root
	if this.RuntimeOverlay != nil {
		switch overlay := this.RuntimeOverlay.(type) {
		case string:
			upgraded.RuntimeOverlay = append(upgraded.RuntimeOverlay, strings.Split(overlay, ",")...)
		case []interface{}:
			for _, each := range overlay {
				upgraded.RuntimeOverlay = append(upgraded.RuntimeOverlay, each.(string))
			}
		default:
			wwlog.Error("unparsable RuntimeOverlay: %v", overlay)
		}
	}
	if this.SystemOverlay != nil {
		switch overlay := this.SystemOverlay.(type) {
		case string:
			upgraded.SystemOverlay = append(upgraded.SystemOverlay, strings.Split(overlay, ",")...)
		case []interface{}:
			for _, each := range overlay {
				upgraded.SystemOverlay = append(upgraded.SystemOverlay, each.(string))
			}
		default:
			wwlog.Error("unparsable SystemOverlay: %v", overlay)
		}
	}
	if replaceOverlays {
		if indexOf(upgraded.SystemOverlay, "wwinit") != -1 {
			upgraded.SystemOverlay = replaceSliceElement(
				upgraded.SystemOverlay,
				indexOf(upgraded.SystemOverlay, "wwinit"),
				wwinitSplitOverlays)
		}
		if indexOf(upgraded.RuntimeOverlay, "generic") != -1 {
			upgraded.RuntimeOverlay = replaceSliceElement(
				upgraded.RuntimeOverlay,
				indexOf(upgraded.RuntimeOverlay, "generic"),
				genericSplitOverlays)
		}
	}
	if this.Tags != nil {
		for key, value := range this.Tags {
			upgraded.Tags[key] = value
		}
	}
	for _, tag := range this.TagsDel {
		delete(upgraded.Tags, tag)
	}
	return
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

func (this *IpmiConf) Upgrade() (upgraded *node.IpmiConf) {
	upgraded = new(node.IpmiConf)
	upgraded.Tags = make(map[string]string)
	upgraded.EscapeChar = this.EscapeChar
	upgraded.Gateway = net.ParseIP(this.Gateway)
	upgraded.Interface = this.Interface
	upgraded.Ipaddr = net.ParseIP(this.Ipaddr)
	upgraded.Netmask = net.ParseIP(this.Netmask)
	upgraded.Password = this.Password
	upgraded.Port = this.Port
	if this.Tags != nil {
		for key, value := range this.Tags {
			upgraded.Tags[key] = value
		}
	}
	for _, tag := range this.TagsDel {
		delete(upgraded.Tags, tag)
	}
	upgraded.UserName = this.UserName
	if this.Write != "" {
		warnError(upgraded.Write.Set(this.Write))
	}
	return
}

type KernelConf struct {
	Args     string `yaml:"args,omitempty"`
	Override string `yaml:"override,omitempty"`
	Version  string `yaml:"version,omitempty"`
}

func (this *KernelConf) Upgrade() (upgraded *node.KernelConf) {
	upgraded = new(node.KernelConf)
	upgraded.Args = this.Args
	upgraded.Override = this.Override
	upgraded.Version = this.Version
	return
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

func (this *NetDev) Upgrade(addDefaults bool) (upgraded *node.NetDev) {
	upgraded = new(node.NetDev)
	upgraded.Tags = make(map[string]string)
	upgraded.Device = this.Device
	upgraded.Gateway = net.ParseIP(this.Gateway)
	upgraded.Hwaddr = this.Hwaddr
	upgraded.Ipaddr = net.ParseIP(this.Ipaddr)
	upgraded.Ipaddr6 = net.ParseIP(this.Ipaddr6)
	upgraded.MTU = this.MTU
	upgraded.Netmask = net.ParseIP(this.Netmask)
	if this.IpCIDR != "" {
		cidrIP, cidrIPNet, err := net.ParseCIDR(this.IpCIDR)
		if err != nil {
			wwlog.Error("%v is not a valid CIDR address: %w", this.IpCIDR, err)
		} else {
			if upgraded.Ipaddr == nil {
				upgraded.Ipaddr = cidrIP
			}
			if upgraded.Netmask == nil {
				upgraded.Netmask = net.IP(cidrIPNet.Mask)
			}
		}
	}
	if this.OnBoot != "" {
		warnError(upgraded.OnBoot.Set(this.OnBoot))
	}
	upgraded.Prefix = net.ParseIP(this.Prefix)
	if this.Tags != nil {
		for key, value := range this.Tags {
			upgraded.Tags[key] = value
		}
	}
	for _, tag := range this.TagsDel {
		delete(upgraded.Tags, tag)
	}
	upgraded.Type = this.Type
	if addDefaults {
		if upgraded.Type == "" {
			upgraded.Type = "ethernet"
		}
		if upgraded.Netmask == nil {
			upgraded.Netmask = net.ParseIP("255.255.255.0")
		}
	}
	return
}

type Disk struct {
	Partitions map[string]*Partition `yaml:"partitions,omitempty"`
	WipeTable  string                `yaml:"wipe_table,omitempty"`
}

func (this *Disk) Upgrade() (upgraded *node.Disk) {
	upgraded = new(node.Disk)
	upgraded.Partitions = make(map[string]*node.Partition)
	if this.Partitions != nil {
		for name, partition := range this.Partitions {
			upgraded.Partitions[name] = partition.Upgrade()
		}
	}
	upgraded.WipeTable, _ = strconv.ParseBool(this.WipeTable)
	return
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

func (this *Partition) Upgrade() (upgraded *node.Partition) {
	upgraded = new(node.Partition)
	upgraded.Guid = this.Guid
	upgraded.Number = this.Number
	upgraded.Resize, _ = strconv.ParseBool(this.Resize)
	upgraded.ShouldExist, _ = strconv.ParseBool(this.ShouldExist)
	upgraded.SizeMiB = this.SizeMiB
	upgraded.StartMiB = this.StartMiB
	upgraded.TypeGuid = this.TypeGuid
	upgraded.WipePartitionEntry, _ = strconv.ParseBool(this.WipePartitionEntry)
	return
}

type FileSystem struct {
	Format         string      `yaml:"format,omitempty"`
	Label          string      `yaml:"label,omitempty"`
	MountOptions   interface{} `yaml:"mount_options,omitempty"`
	Options        []string    `yaml:"options,omitempty"`
	Path           string      `yaml:"path,omitempty"`
	Uuid           string      `yaml:"uuid,omitempty"`
	WipeFileSystem string      `yaml:"wipe_filesystem,omitempty"`
}

func (this *FileSystem) Upgrade() (upgraded *node.FileSystem) {
	upgraded = new(node.FileSystem)
	upgraded.Options = make([]string, 0)
	upgraded.Format = this.Format
	upgraded.Label = this.Label
	if this.MountOptions != nil {
		switch mountOptions := this.MountOptions.(type) {
		case string:
			upgraded.MountOptions = mountOptions
		case []interface{}:
			mountOptionsStrings := make([]string, 0)
			for _, option := range mountOptions {
				mountOptionsStrings = append(mountOptionsStrings, option.(string))
			}
			upgraded.MountOptions = strings.Join(mountOptionsStrings, " ")
		default:
			wwlog.Error("unparsable MountOptions: %v", mountOptions)
		}
	}
	upgraded.Options = append(upgraded.Options, this.Options...)
	upgraded.Path = this.Path
	upgraded.Uuid = this.Uuid
	upgraded.WipeFileSystem, _ = strconv.ParseBool(this.WipeFileSystem)
	return
}
