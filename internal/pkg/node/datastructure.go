package node

import (
	"encoding/gob"
	"net"

	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwtype"
)

func init() {
	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})
	gob.Register(map[interface{}]interface{}{})
}

/******
 * YAML data representations
 ******/
/*
Structure of which goes to disk
*/
type NodesYaml struct {
	NodeProfiles map[string]*Profile `yaml:"nodeprofiles"`
	Nodes        map[string]*Node    `yaml:"nodes"`
}

/*
Node is the datastructure describing a node and a profile which in disk format.
*/
type Node struct {
	id    string
	valid bool // Is set true, if called by the constructor
	// exported values
	Discoverable wwtype.WWbool     `yaml:"discoverable,omitempty" json:"discoverable,omitempty" lopt:"discoverable" sopt:"e" comment:"discoverable in given network (true/false)"`
	AssetKey     string            `yaml:"asset key,omitempty"    json:"asset key,omitempty"    lopt:"asset"                 comment:"the node's Asset tag (key)"`
	Profile      `yaml:"-,inline"` // include all values set in the profile, but inline them in yaml output if these are part of Node
}

/*
Holds the data which can be set for profiles and nodes.
*/
type Profile struct {
	id string
	// exported values
	Profiles       []string               `yaml:"profiles,omitempty"         json:"profiles,omitempty"         lopt:"profile"             sopt:"P" comment:"the node's profile members (comma separated)"`
	Comment        string                 `yaml:"comment,omitempty"          json:"comment,omitempty"          lopt:"comment"                      comment:"arbitrary string comment"`
	ClusterName    string                 `yaml:"cluster name,omitempty"     json:"cluster name,omitempty"     lopt:"cluster"             sopt:"c" comment:"cluster group"`
	ImageName      string                 `yaml:"image name,omitempty"       json:"image name,omitempty"       lopt:"image"                        comment:"image name"`
	Ipxe           string                 `yaml:"ipxe template,omitempty"    json:"ipxe template,omitempty"    lopt:"ipxe"                         comment:"the iPXE template name"`
	RuntimeOverlay []string               `yaml:"runtime overlay,omitempty"  json:"runtime overlay,omitempty"  lopt:"runtime-overlays"    sopt:"R" comment:"the runtime overlay"`
	SystemOverlay  []string               `yaml:"system overlay,omitempty"   json:"system overlay,omitempty"   lopt:"system-overlays"     sopt:"O" comment:"the system overlay"`
	Kernel         *KernelConf            `yaml:"kernel,omitempty"           json:"kernel,omitempty"`
	Ipmi           *IpmiConf              `yaml:"ipmi,omitempty"             json:"ipmi,omitempty"`
	Init           string                 `yaml:"init,omitempty"             json:"init,omitempty"             lopt:"init"                sopt:"i" comment:"the init process to boot the image"`
	Root           string                 `yaml:"root,omitempty"             json:"root,omitempty"             lopt:"root"                         comment:"the rootfs" `
	NetDevs        map[string]*NetDev     `yaml:"network devices,omitempty"  json:"network devices,omitempty"`
	Tags           map[string]string      `yaml:"tags,omitempty"             json:"tags,omitempty"`
	PrimaryNetDev  string                 `yaml:"primary network,omitempty"  json:"primary network,omitempty"  lopt:"primarynet"          sopt:"p" comment:"the primary network interface"`
	Disks          map[string]*Disk       `yaml:"disks,omitempty"            json:"disks,omitempty"`
	FileSystems    map[string]*FileSystem `yaml:"filesystems,omitempty"      json:"filesystems,omitempty"`
	Resources      map[string]Resource    `yaml:"resources,omitempty"        json:"resources,omitempty"`
}

type IpmiConf struct {
	UserName   string            `yaml:"username,omitempty"   json:"username,omitempty"   lopt:"ipmiuser"       comment:"the IPMI username"`
	Password   string            `yaml:"password,omitempty"   json:"password,omitempty"   lopt:"ipmipass"       comment:"the IPMI password"`
	Ipaddr     net.IP            `yaml:"ipaddr,omitempty"     json:"ipaddr,omitempty"     lopt:"ipmiaddr"       comment:"the IPMI IP address" type:"IP"`
	Gateway    net.IP            `yaml:"gateway,omitempty"    json:"gateway,omitempty"    lopt:"ipmigateway"    comment:"the IPMI gateway" type:"IP"`
	Netmask    net.IP            `yaml:"netmask,omitempty"    json:"netmask,omitempty"    lopt:"ipminetmask"    comment:"the IPMI netmask" type:"IP"`
	Port       string            `yaml:"port,omitempty"       json:"port,omitempty"       lopt:"ipmiport"       comment:"the IPMI port"`
	Interface  string            `yaml:"interface,omitempty"  json:"interface,omitempty"  lopt:"ipmiinterface"  comment:"the node's IPMI interface (defaults: 'lan')"`
	EscapeChar string            `yaml:"escapechar,omitempty" json:"escapechar,omitempty" lopt:"ipmiescapechar" comment:"the IPMI escape character (defaults: '~')"`
	Write      wwtype.WWbool     `yaml:"write,omitempty"      json:"write,omitempty"      lopt:"ipmiwrite"      comment:"writing of IPMI configuration (true/false)"`
	Template   string            `yaml:"template,omitempty"   json:"template,omitempty"   lopt:"ipmitemplate"   comment:"template used for ipmi command"`
	Tags       map[string]string `yaml:"tags,omitempty"       json:"tags,omitempty"`
}

type KernelConf struct {
	Version string   `yaml:"version,omitempty" json:"version,omitempty" lopt:"kernelversion"          comment:"kernel version"`
	Args    []string `yaml:"args,omitempty"    json:"args,omitempty"    lopt:"kernelargs"    sopt:"A" comment:"kernel arguments"`
}

type NetDev struct {
	Type       string            `yaml:"type,omitempty"       json:"type,omitempty"       lopt:"type"       sopt:"T" comment:"device type of given network"                       scope:"net"`
	OnBoot     wwtype.WWbool     `yaml:"onboot,omitempty"     json:"onbot,omitempty"      lopt:"onboot"              comment:"network device (true/false)"                          scope:"net"`
	Device     string            `yaml:"device,omitempty"     json:"device,omitempty"     lopt:"netdev"     sopt:"N" comment:"the device for given network"                         scope:"net"`
	Hwaddr     string            `yaml:"hwaddr,omitempty"     json:"hwaddr,omitempty"     lopt:"hwaddr"     sopt:"H" comment:"the device's HW address for given network" type:"MAC" scope:"net"`
	Ipaddr     net.IP            `yaml:"ipaddr,omitempty"     json:"ipaddr,omitempty"     lopt:"ipaddr"     sopt:"I" comment:"IPv4 address in given network" type:"IP"              scope:"net"`
	Netmask    net.IP            `yaml:"netmask,omitempty"    json:"netmask,omitempty"    lopt:"netmask"    sopt:"M" comment:"the network's netmask" type:"IP"                      scope:"net"`
	Gateway    net.IP            `yaml:"gateway,omitempty"    json:"gateway,omitempty"    lopt:"gateway"    sopt:"G" comment:"the node's IPv4 network device gateway" type:"IP"     scope:"net"`
	Ipaddr6    net.IP            `yaml:"ipaddr6,omitempty"    json:"ipaddr6,omitempty"    lopt:"ipaddr6"             comment:"IPv6 address in given network" type:"IP"              scope:"net"`
	PrefixLen6 string            `yaml:"prefixlen6,omitempty" json:"prefixlen6,omitempty" lopt:"prefixlen6"          comment:"the network's IPv6 prefix length" type:"uint"          scope:"net"`
	Gateway6   net.IP            `yaml:"gateway6,omitempty"   json:"gateway6,omitempty"   lopt:"gateway6"            comment:"the node's IPv6 network device gateway" type:"IP"     scope:"net"`
	MTU        string            `yaml:"mtu,omitempty"        json:"mtu,omitempty"        lopt:"mtu"                 comment:"the MTU" type:"uint"                                  scope:"net"`
	Tags       map[string]string `yaml:"tags,omitempty"       json:"tags,omitempty"`
	primary    bool
}

/*
Holds the disks of a node
*/
type Disk struct {
	id         string                `yaml:"-"                    json:"-"`
	WipeTableP *bool                 `yaml:"wipe_table,omitempty" json:"wipe_table,omitempty" lopt:"diskwipe" comment:"whether or not the partition tables shall be wiped" name:"WipeTable" scope:"disk"`
	Partitions map[string]*Partition `yaml:"partitions,omitempty" json:"partitions,omitempty"`
}

/*
partition definition, the label must be uniq so its used as the key in the
Partitions map
*/
type Partition struct {
	id                  string `yaml:"-"                              json:"-"`
	Number              string `yaml:"number,omitempty"               json:"number,omitempty"               lopt:"partnumber" comment:"the partition number (if not set, next free slot is used)" type:"uint"  scope:"disk,part"`
	SizeMiB             string `yaml:"size_mib,omitempty"             json:"size_mib,omitempty"             lopt:"partsize"   comment:"the partition size (if not set, maximum possible size is used)" type:"uint" scope:"disk,part"`
	StartMiB            string `yaml:"start_mib,omitempty"            json:"start_mib,omitempty"                              comment:"the start of the partition" type:"uint"`
	TypeGuid            string `yaml:"type_guid,omitempty"            json:"type_guid,omitempty"            lopt:"parttype"   comment:"the partition type GUID"                                                         scope:"disk,part"`
	Guid                string `yaml:"guid,omitempty"                 json:"guid,omitempty"                                   comment:"the GPT unique partition GUID"`
	WipePartitionEntryP *bool  `yaml:"wipe_partition_entry,omitempty" json:"wipe_partition_entry,omitempty" lopt:"partwipe"   comment:"if true, Ignition will clobber an existing partition if it does not match the config" name:"WipePartitionEntry" scope:"disk,part"`
	ShouldExistP        *bool  `yaml:"should_exist,omitempty"         json:"should_exist,omitempty"         lopt:"partcreate" comment:"create the partition if it does not exist" name:"ShouldExist"                        scope:"disk,part"`
	ResizeP             *bool  `yaml:"resize,omitempty"               json:"resize,omitempty"                                 comment:"whether or not the existing partition should be resize" name:"Resize"`
}

/*
Definition of a filesystem. The device is uniq so its used as key
*/
type FileSystem struct {
	id              string   `yaml:"-"                         json:"-"`
	Format          string   `yaml:"format,omitempty"          json:"format,omitempty"          lopt:"fsformat" comment:"format of the file system"                                                   scope:"fs"`
	Path            string   `yaml:"path,omitempty"            json:"path,omitempty"            lopt:"fspath"   comment:"the mount point of the file system"                                           scope:"fs"`
	WipeFileSystemP *bool    `yaml:"wipe_filesystem,omitempty" json:"wipe_filesystem,omitempty" lopt:"fswipe"   comment:"wipe file system at boot" name:"WipeFileSystem"                               scope:"fs"`
	Label           string   `yaml:"label,omitempty"           json:"label,omitempty"                           comment:"the label of the filesystem"`
	Uuid            string   `yaml:"uuid,omitempty"            json:"uuid,omitempty"                            comment:"the uuid of the filesystem"`
	Options         []string `yaml:"options,omitempty"         json:"options,omitempty"                         comment:"any additional options to be passed to the format-specific mkfs utility"`
	MountOptions    string   `yaml:"mount_options,omitempty"   json:"mount_options,omitempty"                   comment:"any special options to be passed to the mount command"`
}

type Resource interface{}

// Disk methods
func (disk *Disk) WipeTable() bool {
	return util.BoolP(disk.WipeTableP)
}

// Partition methods
func (partition *Partition) WipePartitionEntry() bool {
	return util.BoolP(partition.WipePartitionEntryP)
}

func (partition *Partition) ShouldExist() bool {
	return util.BoolP(partition.ShouldExistP)
}

func (partition *Partition) Resize() bool {
	return util.BoolP(partition.ResizeP)
}

// FileSystem methods
func (fs *FileSystem) WipeFileSystem() bool {
	return util.BoolP(fs.WipeFileSystemP)
}
