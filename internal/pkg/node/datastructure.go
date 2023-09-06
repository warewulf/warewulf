package node

/******
 * YAML data representations
 ******/
/*
Structure of which goes to disk
*/
type NodeYaml struct {
	WWInternal   int `yaml:"WW_INTERNAL"`
	NodeProfiles map[string]*NodeConf
	Nodes        map[string]*NodeConf
}

/*
NodeConf is the datastructure describing a node and a profile which in disk format.
*/
type NodeConf struct {
	Comment       string `yaml:"comment,omitempty" lopt:"comment" comment:"Set arbitrary string comment"`
	ClusterName   string `yaml:"cluster name,omitempty" lopt:"cluster" sopt:"c" comment:"Set cluster group"`
	ContainerName string `yaml:"container name,omitempty" lopt:"container" sopt:"C" comment:"Set container name"`
	Ipxe          string `yaml:"ipxe template,omitempty" lopt:"ipxe" comment:"Set the iPXE template name"`
	// Deprecated start
	// Kernel settings here are deprecated and here for backward compatibility
	KernelVersion  string `yaml:"kernel version,omitempty"`
	KernelOverride string `yaml:"kernel override,omitempty"`
	KernelArgs     string `yaml:"kernel args,omitempty"`
	// Ipmi settings herer are deprecated and here for backward compatibility
	IpmiUserName  string `yaml:"ipmi username,omitempty"`
	IpmiPassword  string `yaml:"ipmi password,omitempty"`
	IpmiIpaddr    string `yaml:"ipmi ipaddr,omitempty"`
	IpmiNetmask   string `yaml:"ipmi netmask,omitempty"`
	IpmiPort      string `yaml:"ipmi port,omitempty"`
	IpmiGateway   string `yaml:"ipmi gateway,omitempty"`
	IpmiInterface string `yaml:"ipmi interface,omitempty"`
	IpmiWrite     string `yaml:"ipmi write,omitempty"`
	// Deprecated end
	RuntimeOverlay []string               `yaml:"runtime overlay,omitempty" lopt:"runtime" sopt:"R" comment:"Set the runtime overlay"`
	SystemOverlay  []string               `yaml:"system overlay,omitempty" lopt:"wwinit" sopt:"O" comment:"Set the system overlay"`
	Kernel         *KernelConf            `yaml:"kernel,omitempty"`
	Ipmi           *IpmiConf              `yaml:"ipmi,omitempty"`
	Init           string                 `yaml:"init,omitempty" lopt:"init" sopt:"i" comment:"Define the init process to boot the container"`
	Root           string                 `yaml:"root,omitempty" lopt:"root" comment:"Define the rootfs" `
	AssetKey       string                 `yaml:"asset key,omitempty" lopt:"asset" comment:"Set the node's Asset tag (key)"`
	Discoverable   string                 `yaml:"discoverable,omitempty" lopt:"discoverable" sopt:"e" comment:"Make discoverable in given network (true/false)" type:"bool"`
	Profiles       []string               `yaml:"profiles,omitempty" lopt:"profile" sopt:"P" comment:"Set the node's profile members (comma separated)"`
	NetDevs        map[string]*NetDevs    `yaml:"network devices,omitempty"`
	Tags           map[string]string      `yaml:"tags,omitempty" lopt:"tagadd" comment:"base key"`
	TagsDel        []string               `yaml:"tagsdel,omitempty" lopt:"tagdel" comment:"remove this tags"` // should not go to disk only to wire
	Keys           map[string]string      `yaml:"keys,omitempty"`                                             // Reverse compatibility
	PrimaryNetDev  string                 `yaml:"primary network,omitempty" lopt:"primarynet" sopt:"p" comment:"Set the primary network interface"`
	Disks          map[string]*Disk       `yaml:"disks,omitempty"`
	FileSystems    map[string]*FileSystem `yaml:"filesystems,omitempty"`
}

type IpmiConf struct {
	UserName  string            `yaml:"username,omitempty" lopt:"ipmiuser" comment:"Set the IPMI username"`
	Password  string            `yaml:"password,omitempty" lopt:"ipmipass" comment:"Set the IPMI password"`
	Ipaddr    string            `yaml:"ipaddr,omitempty" lopt:"ipmiaddr" comment:"Set the IPMI IP address" type:"IP"`
	Netmask   string            `yaml:"netmask,omitempty" lopt:"ipminetmask" comment:"Set the IPMI netmask" type:"IP"`
	Port      string            `yaml:"port,omitempty" lopt:"ipmiport" comment:"Set the IPMI port"`
	Gateway   string            `yaml:"gateway,omitempty" lopt:"ipmigateway" comment:"Set the IPMI gateway" type:"IP"`
	Interface string            `yaml:"interface,omitempty" lopt:"ipmiinterface" comment:"Set the node's IPMI interface (defaults: 'lan')"`
	Write     string            `yaml:"write,omitempty" lopt:"ipmiwrite" comment:"Enable the write of impi configuration (true/false)" type:"bool"`
	Tags      map[string]string `yaml:"tags,omitempty" lopt:"ipmitagadd" comment:"add ipmitags"`
	TagsDel   []string          `yaml:"tagsdel,omitempty" lopt:"ipmitagdel" comment:"remove ipmitags"` // should not go to disk only to wire
}
type KernelConf struct {
	Version  string `yaml:"version,omitempty"`
	Override string `yaml:"override,omitempty" lopt:"kerneloverride" sopt:"K" comment:"Set kernel override version"`
	Args     string `yaml:"args,omitempty" lopt:"kernelargs" sopt:"A" comment:"Set Kernel argument"`
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
	Primary string            `yaml:"primary,omitempty" lopt:"primary" comment:"Enable/disable network device as primary (true/false)" type:"bool"`
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
	Ipaddr    Entry
	Netmask   Entry
	Port      Entry
	Gateway   Entry
	UserName  Entry
	Password  Entry
	Interface Entry
	Write     Entry
	Tags      map[string]*Entry
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

