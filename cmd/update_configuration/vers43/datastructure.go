package vers43

/******
 * YAML data representations
 ******/

//type nodeYaml struct {
type NodeYaml struct { // <- needs to be exported
	WWInternal   int `yaml:"WW_INTERNAL"`
	NodeProfiles map[string]*NodeConf
	Nodes        map[string]*NodeConf
}

/*
NodeConf is the datastructure which is stored on disk.
*/
type NodeConf struct {
	Comment        string              `yaml:"comment,omitempty"`
	ClusterName    string              `yaml:"cluster name,omitempty"`
	ContainerName  string              `yaml:"container name,omitempty"`
	Ipxe           string              `yaml:"ipxe template,omitempty"`
	RuntimeOverlay []string            `yaml:"runtime overlay,omitempty"`
	SystemOverlay  []string            `yaml:"system overlay,omitempty"`
	Kernel         *KernelConf         `yaml:"kernel,omitempty"`
	Ipmi           *IpmiConf           `yaml:"ipmi,omitempty"`
	Init           string              `yaml:"init,omitempty"`
	Root           string              `yaml:"root,omitempty"`
	AssetKey       string              `yaml:"asset key,omitempty"`
	Discoverable   string              `yaml:"discoverable,omitempty"`
	Profiles       []string            `yaml:"profiles,omitempty"`
	NetDevs        map[string]*NetDevs `yaml:"network devices,omitempty"`
	Tags           map[string]string   `yaml:"tags,omitempty"`
	Keys           map[string]string   `yaml:"keys,omitempty"` // Reverse compatibility
}

type IpmiConf struct {
	UserName  string `yaml:"username,omitempty"`
	Password  string `yaml:"password,omitempty"`
	Ipaddr    string `yaml:"ipaddr,omitempty"`
	Netmask   string `yaml:"netmask,omitempty"`
	Port      string `yaml:"port,omitempty"`
	Gateway   string `yaml:"gateway,omitempty"`
	Interface string `yaml:"interface,omitempty"`
	Write     bool   `yaml:"write,omitempty"`
}
type KernelConf struct {
	Version  string `yaml:"version,omitempty"`
	Override string `yaml:"override,omitempty"`
	Args     string `yaml:"args,omitempty"`
}

type NetDevs struct {
	Type    string            `yaml:"type,omitempty"`
	OnBoot  string            `yaml:"onboot,omitempty"`
	Device  string            `yaml:"device,omitempty"`
	Hwaddr  string            `yaml:"hwaddr,omitempty"`
	Ipaddr  string            `yaml:"ipaddr,omitempty"`
	IpCIDR  string            `yaml:"ipcidr,omitempty"`
	Ipaddr6 string            `yaml:"ip6addr,omitempty"`
	Prefix  string            `yaml:"prefix,omitempty"`
	Netmask string            `yaml:"netmask,omitempty"`
	Gateway string            `yaml:"gateway,omitempty"`
	Default string            `yaml:"default,omitempty"`
	Tags    map[string]string `yaml:"tags,omitempty"`
}
