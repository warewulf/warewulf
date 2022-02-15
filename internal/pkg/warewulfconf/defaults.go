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
	}
	Tftp := &TftpConf{
		Enabled:     true,
		TftpRoot:    "/var/lib/tftpboot",
		SystemdName: "tftp",
	}

	return &ControllerConf{
		Warewulf: Warewulf,
		Dhcp:     Dhcp,
		Tftp:     Tftp,
		current:  false,
	}
}
