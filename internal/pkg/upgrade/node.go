package upgrade

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/warewulf/warewulf/internal/pkg/kernel"
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

func ParseNodes(data []byte) (nodesYaml *NodesYaml, err error) {
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

func (legacy *NodesYaml) Upgrade(addDefaults bool, replaceOverlays bool, warewulfconf *WarewulfYaml) (upgraded *node.NodesYaml) {
	upgraded = new(node.NodesYaml)
	upgraded.NodeProfiles = make(map[string]*node.Profile)
	upgraded.Nodes = make(map[string]*node.Node)
	if legacy.WWInternal != "" {
		logIgnore("WW_INTERNAL", legacy.WWInternal, "obsolete")
	}
	for name, profile := range legacy.NodeProfiles {
		upgraded.NodeProfiles[name] = profile.Upgrade(addDefaults, replaceOverlays)
	}
	for name, node := range legacy.Nodes {
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
		if len(defaultProfile.Kernel.Args) < 1 {
			defaultProfile.Kernel.Args = []string{"quiet", "crashkernel=no", "vga=791", "net.naming-scheme=v238"}
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
	if warewulfconf != nil && warewulfconf.NFS != nil {
		var fstab []map[string]string
		for _, export := range warewulfconf.NFS.Exports {
			fstab = append(fstab, map[string]string{
				"spec":    fmt.Sprintf("warewulf:%s", export),
				"file":    export,
				"vfstype": "nfs",
			})
		}
		for _, export := range warewulfconf.NFS.ExportsExtended {
			if export.Mount != nil && *(export.Mount) {
				entry := map[string]string{
					"spec":    fmt.Sprintf("warewulf:%s", export.Path),
					"file":    export.Path,
					"vfstype": "nfs",
				}
				if export.MountOptions != "" {
					entry["mntops"] = export.MountOptions
				}
				fstab = append(fstab, entry)
			}
		}
		if len(fstab) > 0 {
			if _, ok := upgraded.NodeProfiles["default"]; !ok {
				upgraded.NodeProfiles["default"] = new(node.Profile)
			}
			if upgraded.NodeProfiles["default"].Resources == nil {
				upgraded.NodeProfiles["default"].Resources = make(map[string]node.Resource)
			}
			if _, ok := upgraded.NodeProfiles["default"].Resources["fstab"]; ok {
				if prevFstab, ok := (upgraded.NodeProfiles["default"].Resources["fstab"]).([]map[string]string); ok {
					newFstab := append(prevFstab, fstab...)
					upgraded.NodeProfiles["default"].Resources["fstab"] = newFstab
				} else {
					wwlog.Warn("Unable to port NFS mounts from warewulf.conf: incompatible existing fstab resource in default profile")
				}
			} else {
				upgraded.NodeProfiles["default"].Resources["fstab"] = fstab
			}
		}
	}
	return upgraded
}

type Node struct {
	Profile `yaml:"-,inline"`
}

func (legacy *Node) Upgrade(addDefaults bool, replaceOverlays bool) (upgraded *node.Node) {
	upgraded = new(node.Node)
	upgraded.Tags = make(map[string]string)
	upgraded.Disks = make(map[string]*node.Disk)
	upgraded.FileSystems = make(map[string]*node.FileSystem)
	upgraded.Ipmi = new(node.IpmiConf)
	upgraded.Kernel = new(node.KernelConf)
	upgraded.NetDevs = make(map[string]*node.NetDev)
	upgraded.AssetKey = legacy.AssetKey
	upgraded.ClusterName = legacy.ClusterName
	upgraded.Comment = legacy.Comment
	upgraded.ImageName = legacy.ImageName
	if upgraded.ImageName == "" {
		upgraded.ImageName = legacy.ContainerName
	}
	if legacy.Disabled != "" {
		logIgnore("Disabled", legacy.Disabled, "obsolete")
	}
	if legacy.Discoverable != "" {
		warnError(upgraded.Discoverable.Set(legacy.Discoverable))
	}
	if legacy.Disks != nil {
		for name, disk := range legacy.Disks {
			upgraded.Disks[name] = disk.Upgrade()
		}
	}
	if legacy.FileSystems != nil {
		for name, fileSystem := range legacy.FileSystems {
			upgraded.FileSystems[name] = fileSystem.Upgrade()
		}
	}
	upgraded.Init = legacy.Init
	if legacy.Ipmi != nil {
		upgraded.Ipmi = legacy.Ipmi.Upgrade()
	} else {
		upgraded.Ipmi = new(node.IpmiConf)
	}
	if upgraded.Ipmi.EscapeChar == "" {
		upgraded.Ipmi.EscapeChar = legacy.IpmiEscapeChar
	}
	if upgraded.Ipmi.Gateway.Equal(net.IP{}) {
		upgraded.Ipmi.Gateway = net.ParseIP(legacy.IpmiGateway)
	}
	if upgraded.Ipmi.Interface == "" {
		upgraded.Ipmi.Interface = legacy.IpmiInterface
	}
	if upgraded.Ipmi.Ipaddr.Equal(net.IP{}) {
		upgraded.Ipmi.Ipaddr = net.ParseIP(legacy.IpmiIpaddr)
	}
	if upgraded.Ipmi.Netmask.Equal(net.IP{}) {
		upgraded.Ipmi.Netmask = net.ParseIP(legacy.IpmiNetmask)
	}
	if upgraded.Ipmi.Password == "" {
		upgraded.Ipmi.Password = legacy.IpmiPassword
	}
	if upgraded.Ipmi.Port == "" {
		upgraded.Ipmi.Port = legacy.IpmiPort
	}
	if upgraded.Ipmi.UserName == "" {
		upgraded.Ipmi.UserName = legacy.IpmiUserName
	}
	if upgraded.Ipmi.Write == "" && legacy.IpmiWrite != "" {
		warnError(upgraded.Ipmi.Write.Set(legacy.IpmiWrite))
	}
	upgraded.Ipxe = legacy.Ipxe
	if legacy.Kernel != nil {
		upgraded.Kernel = legacy.Kernel.Upgrade(upgraded.ImageName)
	} else {
		inlineKernel := &KernelConf{
			Args:     legacy.KernelArgs,
			Version:  legacy.KernelVersion,
			Override: legacy.KernelOverride,
		}
		upgraded.Kernel = inlineKernel.Upgrade(upgraded.ImageName)
	}
	if legacy.Keys != nil {
		for key, value := range legacy.Keys {
			upgraded.Tags[key] = value
		}
	}
	if legacy.NetDevs != nil {
		for name, netDev := range legacy.NetDevs {
			upgraded.NetDevs[name] = netDev.Upgrade(false)
			if addDefaults {
				if upgraded.NetDevs[name].Type == "" {
					wwlog.Warn("NetDevs[%s].Type not specified: verify default settings manually", name)
				}
				if len(upgraded.NetDevs[name].Netmask) == 0 {
					wwlog.Warn("NetDevs[%s].Netmask not specified: verify default settings manually", name)
				}
			}
		}
	}
	if legacy.PrimaryNetDev != "" {
		upgraded.PrimaryNetDev = legacy.PrimaryNetDev
	} else {
		for name, netDev := range legacy.NetDevs {
			if b, _ := strconv.ParseBool(netDev.Primary); b {
				upgraded.PrimaryNetDev = name
				break
			} else if b, _ := strconv.ParseBool(netDev.Default); b {
				upgraded.PrimaryNetDev = name
				break
			}
		}
	}
	upgraded.Profiles = append(upgraded.Profiles, legacy.Profiles...)
	if addDefaults {
		if len(upgraded.Profiles) == 0 {
			upgraded.Profiles = append(upgraded.Profiles, "default")
		}
	}
	upgraded.Root = legacy.Root
	if legacy.RuntimeOverlay != nil {
		switch overlay := legacy.RuntimeOverlay.(type) {
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
	if legacy.SystemOverlay != nil {
		switch overlay := legacy.SystemOverlay.(type) {
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
	if legacy.Tags != nil {
		for key, value := range legacy.Tags {
			upgraded.Tags[key] = value
		}
	}
	for _, tag := range legacy.TagsDel {
		delete(upgraded.Tags, tag)
	}
	return
}

type Profile struct {
	AssetKey       string                 `yaml:"asset key,omitempty"`
	ClusterName    string                 `yaml:"cluster name,omitempty"`
	Comment        string                 `yaml:"comment,omitempty"`
	ImageName      string                 `yaml:"image name,omitempty"`
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

func (legacy *Profile) Upgrade(addDefaults bool, replaceOverlays bool) (upgraded *node.Profile) {
	upgraded = new(node.Profile)
	upgraded.Tags = make(map[string]string)
	upgraded.Disks = make(map[string]*node.Disk)
	upgraded.FileSystems = make(map[string]*node.FileSystem)
	upgraded.Kernel = new(node.KernelConf)
	upgraded.NetDevs = make(map[string]*node.NetDev)
	if legacy.AssetKey != "" {
		logIgnore("AssetKey", legacy.AssetKey, "invalid for profiles")
	}
	upgraded.ClusterName = legacy.ClusterName
	upgraded.Comment = legacy.Comment
	upgraded.ImageName = legacy.ImageName
	if upgraded.ImageName == "" {
		upgraded.ImageName = legacy.ContainerName
	}
	if legacy.Disabled != "" {
		logIgnore("Disabled", legacy.Disabled, "obsolete")
	}
	if legacy.Discoverable != "" {
		logIgnore("Discoverable", legacy.Discoverable, "invalid for profiles")
	}
	if legacy.Disks != nil {
		for name, disk := range legacy.Disks {
			upgraded.Disks[name] = disk.Upgrade()
		}
	}
	if legacy.FileSystems != nil {
		for name, fileSystem := range legacy.FileSystems {
			upgraded.FileSystems[name] = fileSystem.Upgrade()
		}
	}
	upgraded.Init = legacy.Init
	upgraded.Ipmi = new(node.IpmiConf)
	if legacy.Ipmi != nil {
		upgraded.Ipmi = legacy.Ipmi.Upgrade()
	} else {
		upgraded.Ipmi = new(node.IpmiConf)
	}
	if upgraded.Ipmi.EscapeChar == "" {
		upgraded.Ipmi.EscapeChar = legacy.IpmiEscapeChar
	}
	if upgraded.Ipmi.Gateway.Equal(net.IP{}) {
		upgraded.Ipmi.Gateway = net.ParseIP(legacy.IpmiGateway)
	}
	if upgraded.Ipmi.Interface == "" {
		upgraded.Ipmi.Interface = legacy.IpmiInterface
	}
	if upgraded.Ipmi.Ipaddr.Equal(net.IP{}) {
		upgraded.Ipmi.Ipaddr = net.ParseIP(legacy.IpmiIpaddr)
	}
	if upgraded.Ipmi.Netmask.Equal(net.IP{}) {
		upgraded.Ipmi.Netmask = net.ParseIP(legacy.IpmiNetmask)
	}
	if upgraded.Ipmi.Password == "" {
		upgraded.Ipmi.Password = legacy.IpmiPassword
	}
	if upgraded.Ipmi.Port == "" {
		upgraded.Ipmi.Port = legacy.IpmiPort
	}
	if upgraded.Ipmi.UserName == "" {
		upgraded.Ipmi.UserName = legacy.IpmiUserName
	}
	if upgraded.Ipmi.Write == "" && legacy.IpmiWrite != "" {
		warnError(upgraded.Ipmi.Write.Set(legacy.IpmiWrite))
	}
	upgraded.Ipxe = legacy.Ipxe
	if legacy.Kernel != nil {
		upgraded.Kernel = legacy.Kernel.Upgrade(upgraded.ImageName)
	} else {
		inlineKernel := &KernelConf{
			Args:     legacy.KernelArgs,
			Version:  legacy.KernelVersion,
			Override: legacy.KernelOverride,
		}
		upgraded.Kernel = inlineKernel.Upgrade(upgraded.ImageName)
	}
	if legacy.Keys != nil {
		for key, value := range legacy.Keys {
			upgraded.Tags[key] = value
		}
	}
	if legacy.NetDevs != nil {
		for name, netDev := range legacy.NetDevs {
			upgraded.NetDevs[name] = netDev.Upgrade(addDefaults)
		}
	}
	if legacy.PrimaryNetDev != "" {
		upgraded.PrimaryNetDev = legacy.PrimaryNetDev
	} else {
		for name, netDev := range legacy.NetDevs {
			if b, _ := strconv.ParseBool(netDev.Primary); b {
				upgraded.PrimaryNetDev = name
				break
			} else if b, _ := strconv.ParseBool(netDev.Default); b {
				upgraded.PrimaryNetDev = name
				break
			}
		}
	}
	if upgraded.Profiles == nil {
		upgraded.Profiles = append(upgraded.Profiles, legacy.Profiles...)
	}
	upgraded.Root = legacy.Root
	if legacy.RuntimeOverlay != nil {
		switch overlay := legacy.RuntimeOverlay.(type) {
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
	if legacy.SystemOverlay != nil {
		switch overlay := legacy.SystemOverlay.(type) {
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
	if legacy.Tags != nil {
		for key, value := range legacy.Tags {
			upgraded.Tags[key] = value
		}
	}
	for _, tag := range legacy.TagsDel {
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

func (legacy *IpmiConf) Upgrade() (upgraded *node.IpmiConf) {
	upgraded = new(node.IpmiConf)
	upgraded.Tags = make(map[string]string)
	upgraded.EscapeChar = legacy.EscapeChar
	upgraded.Gateway = net.ParseIP(legacy.Gateway)
	upgraded.Interface = legacy.Interface
	upgraded.Ipaddr = net.ParseIP(legacy.Ipaddr)
	upgraded.Netmask = net.ParseIP(legacy.Netmask)
	upgraded.Password = legacy.Password
	upgraded.Port = legacy.Port
	if legacy.Tags != nil {
		for key, value := range legacy.Tags {
			upgraded.Tags[key] = value
		}
	}
	for _, tag := range legacy.TagsDel {
		delete(upgraded.Tags, tag)
	}
	upgraded.UserName = legacy.UserName
	if legacy.Write != "" {
		warnError(upgraded.Write.Set(legacy.Write))
	}
	return
}

type KernelConf struct {
	Args     interface{} `yaml:"args,omitempty"`
	Override string      `yaml:"override,omitempty"`
	Version  string      `yaml:"version,omitempty"`
}

func (legacy *KernelConf) Upgrade(imageName string) (upgraded *node.KernelConf) {
	upgraded = new(node.KernelConf)
	switch args := legacy.Args.(type) {
	case []string:
		upgraded.Args = args
	case string:
		if args != "" {
			upgraded.Args = strings.Fields(args)
		}
	default:
		wwlog.Warn("unable to parse Kernel.Args: %v", legacy.Args)
	}
	kernels := kernel.FindKernels(imageName)
	wwlog.Debug("referencing kernels: %v (imageName: %v)", kernels, imageName)
	if legacy.Override != "" {
		if version := util.ParseVersion(legacyKernelVersion(legacy.Override)); version != nil {
			for _, kernel_ := range kernels {
				wwlog.Debug("checking if kernel '%v' version '%v' from image '%v' matches override '%v'", kernel_, kernel_.Version(), imageName, legacy.Override)
				if kernel_.Version() == version.String() {
					upgraded.Version = kernel_.Path
					wwlog.Info("kernel override %v -> version %v (image %v)", legacy.Override, upgraded.Version, imageName)
				}
			}
		} else if util.IsFile((&kernel.Kernel{ImageName: imageName, Path: legacy.Override}).FullPath()) {
			upgraded.Version = legacy.Override
		}
		if upgraded.Version == "" {
			imageDisplay := "unknown"
			if imageName != "" {
				imageDisplay = imageName
			}
			wwlog.Warn("unable to resolve kernel override %v (image %v)", legacy.Override, imageDisplay)
		}
	}
	if upgraded.Version == "" {
		upgraded.Version = legacy.Version
	}
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

func (legacy *NetDev) Upgrade(addDefaults bool) (upgraded *node.NetDev) {
	upgraded = new(node.NetDev)
	upgraded.Tags = make(map[string]string)
	upgraded.Device = legacy.Device
	upgraded.Gateway = net.ParseIP(legacy.Gateway)
	upgraded.Hwaddr = legacy.Hwaddr
	upgraded.Ipaddr = net.ParseIP(legacy.Ipaddr)
	upgraded.Ipaddr6 = net.ParseIP(legacy.Ipaddr6)
	upgraded.MTU = legacy.MTU
	upgraded.Netmask = net.ParseIP(legacy.Netmask)
	if legacy.IpCIDR != "" {
		cidrIP, cidrIPNet, err := net.ParseCIDR(legacy.IpCIDR)
		if err != nil {
			wwlog.Error("%v is not a valid CIDR address: %s", legacy.IpCIDR, err)
		} else {
			if upgraded.Ipaddr == nil {
				upgraded.Ipaddr = cidrIP
			}
			if upgraded.Netmask == nil {
				upgraded.Netmask = net.IP(cidrIPNet.Mask)
			}
		}
	}
	if legacy.OnBoot != "" {
		warnError(upgraded.OnBoot.Set(legacy.OnBoot))
	}
	upgraded.Prefix = net.ParseIP(legacy.Prefix)
	if legacy.Tags != nil {
		for key, value := range legacy.Tags {
			upgraded.Tags[key] = value
		}
	}
	for _, tag := range legacy.TagsDel {
		delete(upgraded.Tags, tag)
	}
	upgraded.Type = legacy.Type
	if addDefaults {
		if upgraded.Type == "" {
			upgraded.Type = "ethernet"
		}
		if upgraded.Netmask == nil {
			upgraded.Netmask = net.IP{255, 255, 255, 0}
		}
	}
	return
}

type Disk struct {
	Partitions map[string]*Partition `yaml:"partitions,omitempty"`
	WipeTable  string                `yaml:"wipe_table,omitempty"`
}

func (legacy *Disk) Upgrade() (upgraded *node.Disk) {
	upgraded = new(node.Disk)
	upgraded.Partitions = make(map[string]*node.Partition)
	if legacy.Partitions != nil {
		for name, partition := range legacy.Partitions {
			upgraded.Partitions[name] = partition.Upgrade()
		}
	}
	upgraded.WipeTable, _ = strconv.ParseBool(legacy.WipeTable)
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

func (legacy *Partition) Upgrade() (upgraded *node.Partition) {
	upgraded = new(node.Partition)
	upgraded.Guid = legacy.Guid
	upgraded.Number = legacy.Number
	upgraded.Resize, _ = strconv.ParseBool(legacy.Resize)
	upgraded.ShouldExist, _ = strconv.ParseBool(legacy.ShouldExist)
	upgraded.SizeMiB = legacy.SizeMiB
	upgraded.StartMiB = legacy.StartMiB
	upgraded.TypeGuid = legacy.TypeGuid
	upgraded.WipePartitionEntry, _ = strconv.ParseBool(legacy.WipePartitionEntry)
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

func (legacy *FileSystem) Upgrade() (upgraded *node.FileSystem) {
	upgraded = new(node.FileSystem)
	upgraded.Options = make([]string, 0)
	upgraded.Format = legacy.Format
	upgraded.Label = legacy.Label
	if legacy.MountOptions != nil {
		switch mountOptions := legacy.MountOptions.(type) {
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
	upgraded.Options = append(upgraded.Options, legacy.Options...)
	upgraded.Path = legacy.Path
	upgraded.Uuid = legacy.Uuid
	upgraded.WipeFileSystem, _ = strconv.ParseBool(legacy.WipeFileSystem)
	return
}
