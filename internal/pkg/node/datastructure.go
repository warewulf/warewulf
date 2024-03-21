package node

/******
 * YAML data representations
 ******/
/*
Structure of which goes to disk
*/
type NodeYaml struct {
	WWInternal   int `yaml:"WW_INTERNAL,omitempty" json:"WW_INTERNAL,omitempty"`
	NodeProfiles map[string]*NodeConf
	Nodes        map[string]*NodeConf
}

/*
NodeConf is the datastructure describing a node and a profile which in disk format.
*/
type NodeConf struct {
	Comment       string `yaml:"comment,omitempty" lopt:"comment" comment:"Set arbitrary string comment" json:"comment,omitempty"`
	ClusterName   string `yaml:"cluster name,omitempty" lopt:"cluster" sopt:"c" comment:"Set cluster group" json:"cluster name,omitempty"`
	ContainerName string `yaml:"container name,omitempty" lopt:"container" sopt:"C" comment:"Set container name" json:"container name,omitempty"`
	Ipxe          string `yaml:"ipxe template,omitempty" lopt:"ipxe" comment:"Set the iPXE template name" json:"ipxe template,omitempty"`
	// Deprecated start
	// Kernel settings here are deprecated and here for backward compatibility
	KernelVersion  string `yaml:"kernel version,omitempty" json:"kernel version,omitempty"`
	KernelOverride string `yaml:"kernel override,omitempty" json:"kernel override,omitempty"`
	KernelArgs     string `yaml:"kernel args,omitempty" json:"kernel args,omitempty"`
	// Ipmi settings herer are deprecated and here for backward compatibility
	IpmiUserName   string `yaml:"ipmi username,omitempty" json:"ipmi username,omitempty"`
	IpmiPassword   string `yaml:"ipmi password,omitempty" json:"ipmi password,omitempty"`
	IpmiIpaddr     string `yaml:"ipmi ipaddr,omitempty" json:"ipmi ipaddr,omitempty"`
	IpmiNetmask    string `yaml:"ipmi netmask,omitempty" json:"ipmi netmask,omitempty"`
	IpmiPort       string `yaml:"ipmi port,omitempty" json:"ipmi port,omitempty"`
	IpmiGateway    string `yaml:"ipmi gateway,omitempty" json:"ipmi gateway,omitempty"`
	IpmiInterface  string `yaml:"ipmi interface,omitempty" json:"ipmi interface,omitempty"`
	IpmiEscapeChar string `yaml:"ipmi escapechar,omitempty" json:"ipmi escapechar,omitempty"`
	IpmiWrite      string `yaml:"ipmi write,omitempty" json:"ipmi write,omitempty"`
	// Deprecated end
	RuntimeOverlay []string               `yaml:"runtime overlay,omitempty" lopt:"runtime" sopt:"R" comment:"Set the runtime overlay" json:"runtime overlay,omitempty"`
	SystemOverlay  []string               `yaml:"system overlay,omitempty" lopt:"wwinit" sopt:"O" comment:"Set the system overlay" json:"system overlay,omitempty"`
	Kernel         *KernelConf            `yaml:"kernel,omitempty" json:"kernel,omitempty"`
	Ipmi           *IpmiConf              `yaml:"ipmi,omitempty" json:"ipmi,omitempty"`
	Init           string                 `yaml:"init,omitempty" lopt:"init" sopt:"i" comment:"Define the init process to boot the container" json:"init,omitempty"`
	Root           string                 `yaml:"root,omitempty" lopt:"root" comment:"Define the rootfs" json:"root,omitempty"`
	AssetKey       string                 `yaml:"asset key,omitempty" lopt:"asset" comment:"Set the node's Asset tag (key)" json:"asset key,omitempty"`
	Discoverable   string                 `yaml:"discoverable,omitempty" lopt:"discoverable" sopt:"e" comment:"Make discoverable in given network  (true/false)" type:"bool" json:"discoverable,omitempty"`
	Profiles       []string               `yaml:"profiles,omitempty" lopt:"profile" sopt:"P" comment:"Set the node's profile members (comma separated)" json:"profiles,omitempty"`
	NetDevs        map[string]*NetDevs    `yaml:"network devices,omitempty" json:"network devices,omitempty"`
	Tags           map[string]string      `yaml:"tags,omitempty" lopt:"tagadd" comment:"base key" json:"tags,omitempty"`
	TagsDel        []string               `yaml:"tagsdel,omitempty" lopt:"tagdel" comment:"remove this tags" json:"tagsdel,omitempty"` // should not go to disk only to wire
	Keys           map[string]string      `yaml:"keys,omitempty" json:"keys,omitempty"`                                                // Reverse compatibility
	PrimaryNetDev  string                 `yaml:"primary network,omitempty" lopt:"primarynet" sopt:"p" comment:"Set the primary network interface" json:"primary network,omitempty"`
	Disks          map[string]*Disk       `yaml:"disks,omitempty" json:"disks,omitempty"`
	FileSystems    map[string]*FileSystem `yaml:"filesystems,omitempty" json:"filesystems,omitempty"`
}

type IpmiConf struct {
	UserName   string            `yaml:"username,omitempty" lopt:"ipmiuser" comment:"Set the IPMI username" json:"username,omitempty"`
	Password   string            `yaml:"password,omitempty" lopt:"ipmipass" comment:"Set the IPMI password" json:"password,omitempty"`
	Ipaddr     string            `yaml:"ipaddr,omitempty" lopt:"ipmiaddr" comment:"Set the IPMI IP address" type:"IP" json:"ipaddr,omitempty"`
	Netmask    string            `yaml:"netmask,omitempty" lopt:"ipminetmask" comment:"Set the IPMI netmask" type:"IP" json:"netmask,omitempty"`
	Port       string            `yaml:"port,omitempty" lopt:"ipmiport" comment:"Set the IPMI port" json:"port,omitempty"`
	Gateway    string            `yaml:"gateway,omitempty" lopt:"ipmigateway" comment:"Set the IPMI gateway" type:"IP" json:"gateway,omitempty"`
	Interface  string            `yaml:"interface,omitempty" lopt:"ipmiinterface" comment:"Set the node's IPMI interface (defaults: 'lan')" json:"interface,omitempty"`
	EscapeChar string            `yaml:"escapechar,omitempty" lopt:"ipmiescapechar" comment:"Set the IPMI escape character (defaults: '~')" json:"escapechar,omitempty"`
	Write      string            `yaml:"write,omitempty" lopt:"ipmiwrite" comment:"Enable the write of impi configuration (true/false)" type:"bool" json:"write,omitempty"`
	Tags       map[string]string `yaml:"tags,omitempty" lopt:"ipmitagadd" comment:"add ipmitags" json:"tags,omitempty"`
	TagsDel    []string          `yaml:"tagsdel,omitempty" lopt:"ipmitagdel" comment:"remove ipmitags" json:"tagsdel,omitempty"` // should not go to disk only to wire
}
type KernelConf struct {
	Version  string `yaml:"version,omitempty" json:"version,omitempty"`
	Override string `yaml:"override,omitempty" lopt:"kerneloverride" sopt:"K" comment:"Set kernel override version" json:"override,omitempty"`
	Args     string `yaml:"args,omitempty" lopt:"kernelargs" sopt:"A" comment:"Set Kernel argument" json:"args,omitempty"`
}

type NetDevs struct {
	Type    string            `yaml:"type,omitempty" lopt:"type" sopt:"T" comment:"Set device type of given network"`
	OnBoot  string            `yaml:"onboot,omitempty" lopt:"onboot" comment:"Enable/disable network device (true/false)" type:"bool"`
	Device  string            `yaml:"device,omitempty" lopt:"netdev" sopt:"N" comment:"Set the device for given network"`
	Hwaddr  string            `yaml:"hwaddr,omitempty" lopt:"hwaddr" sopt:"H" comment:"Set the device's HW address for given network" type:"MAC"`
	Ipaddr  string            `yaml:"ipaddr,omitempty" comment:"IPv4 address in given network" sopt:"I" lopt:"ipaddr" type:"IP"`
	IpCIDR  string            `yaml:"ipcidr,omitempty"`
	Ipaddr6 string            `yaml:"ip6addr,omitempty" lopt:"ipaddr6" comment:"IPv6 address" type:"IP"`
	Prefix  string            `yaml:"prefix,omitempty"`
	Netmask string            `yaml:"netmask,omitempty" lopt:"netmask" sopt:"M" comment:"Set the networks netmask" type:"IP"`
	Gateway string            `yaml:"gateway,omitempty" lopt:"gateway" sopt:"G" comment:"Set the node's network device gateway" type:"IP"`
	MTU     string            `yaml:"mtu,omitempty" lopt:"mtu" comment:"Set the mtu" type:"uint"`
	Primary string            `yaml:"primary,omitempty" type:"bool"`
	Default string            `yaml:"default,omitempty"` /* backward compatibility */
	Tags    map[string]string `yaml:"tags,omitempty" lopt:"nettagadd" comment:"network tags"`
	TagsDel []string          `yaml:"tagsdel,omitempty" lopt:"nettagdel" comment:"delete network tags"` // should not go to disk only to wire
}

/*
Holds the disks of a node
*/
type Disk struct {
	WipeTable  string                `yaml:"wipe_table,omitempty" type:"bool" lopt:"diskwipe" comment:"whether or not the partition tables shall be wiped"`
	Partitions map[string]*Partition `yaml:"partitions,omitempty"`
}

/*
partition definition, the label must be uniq so its used as the key in the
Partitions map
*/
type Partition struct {
	Number             string `yaml:"number,omitempty" lopt:"partnumber" comment:"set the partition number, if not set next free slot is used"`
	SizeMiB            string `yaml:"size_mib,omitempty" lopt:"partsize" comment:"set the size of the partition, if not set maximal possible size is used"`
	StartMiB           string `yaml:"start_mib,omitempty" comment:"the start of the partition"`
	TypeGuid           string `yaml:"type_guid,omitempty" comment:"Linux filesystem data will be used if empty"`
	Guid               string `yaml:"guid,omitempty" comment:"the GPT unique partition GUID"`
	WipePartitionEntry string `yaml:"wipe_partition_entry,omitempty" comment:"if true, Ignition will clobber an existing partition if it does not match the config" type:"bool"`
	ShouldExist        string `yaml:"should_exist,omitempty" lopt:"partcreate" comment:"create partition if not exist" type:"bool"`
	Resize             string `yaml:"resize,omitempty" comment:" whether or not the existing partition should be resize" type:"bool"`
}

/*
Definition of a filesystem. The device is uniq so its used as key
*/
type FileSystem struct {
	Format         string   `yaml:"format,omitempty" lopt:"fsformat" comment:"format of the file system"`
	Path           string   `yaml:"path,omitempty" lopt:"fspath" comment:"the mount point of the file system"`
	WipeFileSystem string   `yaml:"wipe_filesystem,omitempty" lopt:"fswipe" comment:"wipe file system at boot" type:"bool"`
	Label          string   `yaml:"label,omitempty" comment:"the label of the filesystem"`
	Uuid           string   `yaml:"uuid,omitempty" comment:"the uuid of the filesystem"`
	Options        []string `yaml:"options,omitempty" comment:"any additional options to be passed to the format-specific mkfs utility"`
	MountOptions   []string `yaml:"mount_options,omitempty" comment:"any special options to be passed to the mount command"`
}

/******
 * Internal code data representations
 ******/
/*
Holds string values, when accessed via Get, its value
is returned which is the default or if set the value
from the profile or if set the value of the node itself
*/
type Entry struct {
	value    []string
	altvalue []string
	from     string
	def      []string
	isSlice  bool
}

/*
NodeInfo is the in memory datastructure, which can containe
a default value, which is overwritten by the overlay from the
overlay (altvalue) which is overwitten by the value of the
node itself, for all values of type Entry.
*/
type NodeInfo struct {
	Id             Entry
	Comment        Entry
	ClusterName    Entry
	ContainerName  Entry
	Ipxe           Entry
	Grub           Entry
	RuntimeOverlay Entry
	SystemOverlay  Entry
	Root           Entry
	Discoverable   Entry
	Init           Entry //TODO: Finish adding this...
	AssetKey       Entry
	Kernel         *KernelEntry
	Ipmi           *IpmiEntry
	Profiles       Entry
	PrimaryNetDev  Entry
	NetDevs        map[string]*NetDevEntry
	Tags           map[string]*Entry
	Disks          map[string]*DiskEntry
	FileSystems    map[string]*FileSystemEntry
}

type IpmiEntry struct {
	Ipaddr     Entry
	Netmask    Entry
	Port       Entry
	Gateway    Entry
	UserName   Entry
	Password   Entry
	Interface  Entry
	EscapeChar Entry
	Write      Entry
	Tags       map[string]*Entry
}

type KernelEntry struct {
	Override Entry
	Args     Entry
}

type NetDevEntry struct {
	Type    Entry
	OnBoot  Entry
	Device  Entry
	Hwaddr  Entry
	Ipaddr  Entry
	Ipaddr6 Entry
	IpCIDR  Entry
	Prefix  Entry
	Netmask Entry
	Gateway Entry
	MTU     Entry
	Primary Entry
	Tags    map[string]*Entry
}

type DiskEntry struct {
	WipeTable  Entry
	Partitions map[string]*PartitionEntry
}

type PartitionEntry struct {
	Number             Entry
	SizeMiB            Entry
	StartMiB           Entry
	TypeGuid           Entry
	Guid               Entry
	WipePartitionEntry Entry
	ShouldExist        Entry
	Resize             Entry
}

type FileSystemEntry struct {
	Format         Entry
	Path           Entry
	WipeFileSystem Entry
	Label          Entry
	Uuid           Entry
	Options        Entry
	MountOptions   Entry
}

// string which is printed if no value is set
const NoValue = "--"

func (e Entry) MarshalText() (buf []byte, err error) {
	buf = append(buf, []byte(e.Get())...)
	return buf, nil
}
