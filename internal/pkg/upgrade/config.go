package upgrade

import (
	"gopkg.in/yaml.v3"

	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func ParseConfig(data []byte) (warewulfYaml *WarewulfYaml, err error) {
	warewulfYaml = new(WarewulfYaml)
	if err = yaml.Unmarshal(data, warewulfYaml); err != nil {
		return warewulfYaml, err
	}
	return warewulfYaml, nil
}

type WarewulfYaml struct {
	WWInternal      string        `yaml:"WW_INTERNAL"`
	Comment         string        `yaml:"comment"`
	Ipaddr          string        `yaml:"ipaddr"`
	Ipaddr6         string        `yaml:"ipaddr6"`
	Netmask         string        `yaml:"netmask"`
	Network         string        `yaml:"network"`
	Ipv6net         string        `yaml:"ipv6net"`
	Fqdn            string        `yaml:"fqdn"`
	Warewulf        *WarewulfConf `yaml:"warewulf"`
	DHCP            *DHCPConf     `yaml:"dhcp"`
	TFTP            *TFTPConf     `yaml:"tftp"`
	NFS             *NFSConf      `yaml:"nfs"`
	SSH             *SSHConf      `yaml:"ssh"`
	MountsImage     []*MountEntry `yaml:"image mounts"`
	MountsContainer []*MountEntry `yaml:"container mounts"`
	Paths           *BuildConfig  `yaml:"paths"`
	WWClient        *WWClientConf `yaml:"wwclient"`
}

func (legacy *WarewulfYaml) Upgrade() (upgraded *config.WarewulfYaml) {
	upgraded = new(config.WarewulfYaml)
	if legacy.WWInternal != "" {
		logIgnore("WW_INTERNAL", legacy.WWInternal, "obsolete")
	}
	upgraded.Comment = legacy.Comment
	upgraded.Ipaddr = legacy.Ipaddr
	upgraded.Ipaddr6 = legacy.Ipaddr6
	upgraded.Netmask = legacy.Netmask
	upgraded.Network = legacy.Network
	upgraded.Ipv6net = legacy.Ipv6net
	upgraded.Fqdn = legacy.Fqdn
	if legacy.Warewulf != nil {
		upgraded.Warewulf = legacy.Warewulf.Upgrade()
	}
	if legacy.DHCP != nil {
		upgraded.DHCP = legacy.DHCP.Upgrade()
	}
	if legacy.TFTP != nil {
		upgraded.TFTP = legacy.TFTP.Upgrade()
	}
	if legacy.NFS != nil {
		upgraded.NFS = legacy.NFS.Upgrade()
	}
	if legacy.SSH != nil {
		upgraded.SSH = legacy.SSH.Upgrade()
	}
	upgraded.MountsImage = make([]*config.MountEntry, 0)
	for _, mount := range legacy.MountsImage {
		upgraded.MountsImage = append(upgraded.MountsImage, mount.Upgrade())
	}
	if len(upgraded.MountsImage) == 0 {
		for _, mount := range legacy.MountsContainer {
			upgraded.MountsImage = append(upgraded.MountsImage, mount.Upgrade())
		}
	}
	if legacy.Paths != nil {
		upgraded.Paths = legacy.Paths.Upgrade()
	}
	if legacy.WWClient != nil {
		upgraded.WWClient = legacy.WWClient.Upgrade()
	}
	if legacy.Warewulf != nil && legacy.Warewulf.DataStore != "" {
		if upgraded.Paths == nil {
			upgraded.Paths = new(config.BuildConfig)
		}
		if upgraded.Paths.Datadir == "" {
			upgraded.Paths.Datadir = legacy.Warewulf.DataStore
		}
	}
	return upgraded
}

type WarewulfConf struct {
	Port              int    `yaml:"port"`
	Secure            *bool  `yaml:"secure"`
	UpdateInterval    int    `yaml:"update interval"`
	AutobuildOverlays *bool  `yaml:"autobuild overlays"`
	EnableHostOverlay *bool  `yaml:"host overlay"`
	Syslog            *bool  `yaml:"syslog"`
	DataStore         string `yaml:"datastore"`
	GrubBoot          *bool  `yaml:"grubboot"`
}

func (legacy *WarewulfConf) Upgrade() (upgraded *config.WarewulfConf) {
	upgraded = new(config.WarewulfConf)
	upgraded.Port = legacy.Port
	upgraded.SecureP = legacy.Secure
	upgraded.UpdateInterval = legacy.UpdateInterval
	upgraded.AutobuildOverlaysP = legacy.AutobuildOverlays
	upgraded.EnableHostOverlayP = legacy.EnableHostOverlay
	if legacy.Syslog != nil {
		wwlog.Warn("syslog configuration ignored: all logs now go to stdout/stderr")
	}
	upgraded.GrubBootP = legacy.GrubBoot
	return upgraded
}

type DHCPConf struct {
	Enabled     *bool  `yaml:"enabled"`
	Template    string `yaml:"template"`
	RangeStart  string `yaml:"range start"`
	RangeEnd    string `yaml:"range end"`
	SystemdName string `yaml:"systemd name"`
}

func (legacy *DHCPConf) Upgrade() (upgraded *config.DHCPConf) {
	upgraded = new(config.DHCPConf)
	upgraded.EnabledP = legacy.Enabled
	upgraded.Template = legacy.Template
	upgraded.RangeStart = legacy.RangeStart
	upgraded.RangeEnd = legacy.RangeEnd
	upgraded.SystemdName = legacy.SystemdName
	return upgraded
}

type TFTPConf struct {
	Enabled      *bool             `yaml:"enabled"`
	TftpRoot     string            `yaml:"tftproot"`
	SystemdName  string            `yaml:"systemd name"`
	IpxeBinaries map[string]string `yaml:"ipxe"`
}

func (legacy *TFTPConf) Upgrade() (upgraded *config.TFTPConf) {
	upgraded = new(config.TFTPConf)
	upgraded.EnabledP = legacy.Enabled
	upgraded.TftpRoot = legacy.TftpRoot
	upgraded.SystemdName = legacy.SystemdName
	upgraded.IpxeBinaries = make(map[string]string)
	for name, binary := range legacy.IpxeBinaries {
		upgraded.IpxeBinaries[name] = binary
	}
	return upgraded
}

type NFSConf struct {
	Enabled         *bool            `yaml:"enabled"`
	Exports         []string         `yaml:"exports"`
	ExportsExtended []*NFSExportConf `yaml:"export paths"`
	SystemdName     string           `yaml:"systemd name"`
}

func (legacy *NFSConf) Upgrade() (upgraded *config.NFSConf) {
	upgraded = new(config.NFSConf)
	upgraded.EnabledP = legacy.Enabled
	upgraded.ExportsExtended = make([]*config.NFSExportConf, 0)
	for _, export := range legacy.Exports {
		extendedExport := new(config.NFSExportConf)
		extendedExport.Path = export
		upgraded.ExportsExtended = append(upgraded.ExportsExtended, extendedExport)
	}
	for _, export := range legacy.ExportsExtended {
		upgraded.ExportsExtended = append(upgraded.ExportsExtended, export.Upgrade())
	}
	upgraded.SystemdName = legacy.SystemdName
	return upgraded
}

type NFSExportConf struct {
	Path          string `yaml:"path"`
	ExportOptions string `yaml:"export options"`
	MountOptions  string `yaml:"mount options"`
	Mount         *bool  `yaml:"mount"`
}

func (legacy *NFSExportConf) Upgrade() (upgraded *config.NFSExportConf) {
	upgraded = new(config.NFSExportConf)
	upgraded.Path = legacy.Path
	upgraded.ExportOptions = legacy.ExportOptions
	if legacy.Mount != nil && *(legacy.Mount) {
		wwlog.Warn("Legacy mount configured for NFS export %s: use `wwctl upgrade nodes --with-warewulfconf=<original file>` to port to nodes.conf", legacy.Path)
	}
	return upgraded
}

type SSHConf struct {
	KeyTypes []string `yaml:"key types"`
}

func (legacy *SSHConf) Upgrade() (upgraded *config.SSHConf) {
	upgraded = new(config.SSHConf)
	upgraded.KeyTypes = append([]string{}, legacy.KeyTypes...)
	return upgraded
}

type MountEntry struct {
	Source   string `yaml:"source"`
	Dest     string `yaml:"dest"`
	ReadOnly *bool  `yaml:"readonly"`
	Options  string `yaml:"options"`
	Copy     *bool  `yaml:"copy"`
}

func (legacy *MountEntry) Upgrade() (upgraded *config.MountEntry) {
	upgraded = new(config.MountEntry)
	upgraded.Source = legacy.Source
	upgraded.Dest = legacy.Dest
	upgraded.ReadOnlyP = legacy.ReadOnly
	upgraded.Options = legacy.Options
	upgraded.CopyP = legacy.Copy
	return upgraded
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
	Datadir        string
	WWOverlaydir   string
	WWChrootdir    string
	WWProvisiondir string
	WWClientdir    string
}

func (legacy *BuildConfig) Upgrade() (upgraded *config.BuildConfig) {
	upgraded = new(config.BuildConfig)
	upgraded.Bindir = legacy.Bindir
	upgraded.Sysconfdir = legacy.Sysconfdir
	upgraded.Localstatedir = legacy.Localstatedir
	upgraded.Cachedir = legacy.Cachedir
	upgraded.Ipxesource = legacy.Ipxesource
	upgraded.Srvdir = legacy.Srvdir
	upgraded.Firewallddir = legacy.Firewallddir
	upgraded.Datadir = legacy.Datadir
	upgraded.WWOverlaydir = legacy.WWOverlaydir
	upgraded.WWChrootdir = legacy.WWChrootdir
	upgraded.WWProvisiondir = legacy.WWProvisiondir
	upgraded.WWClientdir = legacy.WWClientdir
	return upgraded
}

type WWClientConf struct {
	Port uint16 `yaml:"port"`
}

func (legacy *WWClientConf) Upgrade() (upgraded *config.WWClientConf) {
	upgraded = new(config.WWClientConf)
	upgraded.Port = legacy.Port
	return upgraded
}
