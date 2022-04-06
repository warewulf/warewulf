package warewulfconf

import "github.com/hpcng/warewulf/internal/pkg/buildconfig"

const defaultPort int = 9983

func defaultConfig() *ControllerConf {
	Warewulf := &WarewulfConf{
		Port:              defaultPort,
		Secure:            true,
		UpdateInterval:    60,
		AutobuildOverlays: true,
		EnableHostOverlay: true,
		Syslog:            false,
		DataStore:         buildconfig.LOCALSTATEDIR(),
	}
	Dhcp := &DhcpConf{
		Enabled:     true,
		Template:    "default",
		RangeStart:  "192.168.200.50",
		RangeEnd:    "192.168.200.99",
		SystemdName: "dhcpd",
	}
	Tftp := &TftpConf{
		Enabled:     true,
		TftpRoot:    buildconfig.TFTPDIR(),
		SystemdName: "tftp",
	}

	return &ControllerConf{
		WWInternal: buildconfig.WWVer,
		Warewulf:   Warewulf,
		Dhcp:       Dhcp,
		Tftp:       Tftp,
		current:    false,
	}
}
