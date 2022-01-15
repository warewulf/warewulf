package warewulfconf

const defaultPort int = 9983

var defaultDataStore string = "/var/lib/warewulf"

func defaultConfig() *ControllerConf {
	Warewulf := &WarewulfConf{
		Port:              defaultPort,
		Secure:            true,
		UpdateInterval:    60,
		AutobuildOverlays: true,
		Syslog:            false,
		DataStore:         defaultDataStore,
	}
	Dhcp := &DhcpConf{
		Enabled:     true,
		Template:    "default",
		RangeStart:  "192.168.200.50",
		RangeEnd:    "192.168.200.99",
		SystemdName: "dhcpd",
		ConfigFile:  "/etc/dhcp/dhcpd.conf",
	}
	Tftp := &TftpConf{
		Enabled:     true,
		TftpRoot:    "/var/lib/tftpboot",
		SystemdName: "tftp",
	}
	Nfs := &NfsConf{
		Enabled:     true,
		Exports:     []string{"/home"},
		SystemdName: "nfs-server",
	}

	return &ControllerConf{
		Warewulf: Warewulf,
		Dhcp:     Dhcp,
		Tftp:     Tftp,
		Nfs:      Nfs,
		current:  false,
	}
}
