package upgrade

type RootConf struct {
	WWInternal      int           `yaml:"WW_INTERNAL"`
	Comment         string        `yaml:"comment,omitempty"`
	Ipaddr          string        `yaml:"ipaddr"`
	Ipaddr6         string        `yaml:"ipaddr6,omitempty"`
	Netmask         string        `yaml:"netmask"`
	Network         string        `yaml:"network,omitempty"`
	Ipv6net         string        `yaml:"ipv6net,omitempty"`
	Fqdn            string        `yaml:"fqdn,omitempty"`
	Warewulf        *WarewulfConf `yaml:"warewulf"`
	DHCP            *DHCPConf     `yaml:"dhcp"`
	TFTP            *TFTPConf     `yaml:"tftp"`
	NFS             *NFSConf      `yaml:"nfs"`
	SSH             *SSHConf      `yaml:"ssh,omitempty"`
	MountsContainer []*MountEntry `yaml:"container mounts"`
	Paths           *BuildConfig  `yaml:"paths"`
	WWClient        *WWClientConf `yaml:"wwclient"`
}

type WarewulfConf struct {
	Port              int    `yaml:"port"`
	Secure            bool   `yaml:"secure"`
	UpdateInterval    int    `yaml:"update interval"`
	AutobuildOverlays bool   `yaml:"autobuild overlays"`
	EnableHostOverlay bool   `yaml:"host overlay"`
	Syslog            bool   `yaml:"syslog"`
	DataStore         string `yaml:"datastore"`
	GrubBoot          bool   `yaml:"grubboot"`
}

type DHCPConf struct {
	Enabled     bool   `yaml:"enabled"`
	Template    string `yaml:"template"`
	RangeStart  string `yaml:"range start,omitempty"`
	RangeEnd    string `yaml:"range end,omitempty"`
	SystemdName string `yaml:"systemd name"`
}

type TFTPConf struct {
	Enabled     bool   `yaml:"enabled"`
	TftpRoot    string `yaml:"tftproot"`
	SystemdName string `yaml:"systemd name"`

	IpxeBinaries map[string]string `yaml:"ipxe"`
}

type NFSConf struct {
	Enabled         bool             `yaml:"enabled"`
	ExportsExtended []*NFSExportConf `yaml:"export paths"`
	SystemdName     string           `yaml:"systemd name"`
}

type NFSExportConf struct {
	Path          string `yaml:"path"`
	ExportOptions string `yaml:"export options"`
	MountOptions  string `yaml:"mount options"`
	Mount         bool   `yaml:"mount"`
}

type SSHConf struct {
	KeyTypes []string `yaml:"key types"`
}

type MountEntry struct {
	Source   string `yaml:"source"`
	Dest     string `yaml:"dest,omitempty"`
	ReadOnly bool   `yaml:"readonly,omitempty"`
	Options  string `yaml:"options,omitempty"`
	Copy     bool   `yaml:"copy,omitempty"`
}

type BuildConfig struct {
	Bindir         string
	Sysconfdir     string
	Localstatedir  string
	Cachedir       string
	Ipxesource     string
	Srvdir         string
	Firewallddir   string
	Systemddir     string
	WWOverlaydir   string
	WWChrootdir    string
	WWProvisiondir string
	WWClientdir    string
}

type WWClientConf struct {
	Port uint16 `yaml:"port"`
}
