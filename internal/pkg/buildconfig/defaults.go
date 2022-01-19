package buildconfig

import "github.com/hpcng/warewulf/internal/pkg/wwlog"

var (
	bindir         string = "UNDEF"
	sysconfdir     string = "UNDEF"
	localstatedir  string = "UNDEF"
	srvdir         string = "UNDEF"
	tftpdir        string = "UNDEF"
	firewallddir   string = "UNDEF"
	systemddir     string = "UNDEF"
	wwoverlaydir   string = "UNDEF"
	wwchrootdir    string = "UNDEF"
	wwprovisiondir string = "UNDEF"
	version        string = "UNDEF"
	release        string = "UNDEF"
)

func BINDIR() string {
	wwlog.Printf(wwlog.DEBUG, "BINDIR = '%s'\n", bindir)
	return bindir
}

func SYSCONFDIR() string {
	wwlog.Printf(wwlog.DEBUG, "SYSCONFDIR = '%s'\n", sysconfdir)
	return sysconfdir
}

func LOCALSTATEDIR() string {
	wwlog.Printf(wwlog.DEBUG, "LOCALSTATEDIR = '%s'\n", localstatedir)
	return localstatedir
}

func SRVDIR() string {
	wwlog.Printf(wwlog.DEBUG, "SRVDIR = '%s'\n", srvdir)
	return srvdir
}

func TFTPDIR() string {
	wwlog.Printf(wwlog.DEBUG, "TFTPDIR = '%s'\n", tftpdir)
	return tftpdir
}

func FIREWALLDDIR() string {
	wwlog.Printf(wwlog.DEBUG, "FIREWALLDDIR = '%s'\n", firewallddir)
	return firewallddir
}

func SYSTEMDDIR() string {
	wwlog.Printf(wwlog.DEBUG, "SYSTEMDDIR = '%s'\n", systemddir)
	return systemddir
}

func WWOVERLAYDIR() string {
	wwlog.Printf(wwlog.DEBUG, "WWOVERLAYDIR = '%s'\n", wwoverlaydir)
	return wwoverlaydir
}

func WWCHROOTDIR() string {
	wwlog.Printf(wwlog.DEBUG, "WWCHROOTDIR = '%s'\n", wwchrootdir)
	return wwchrootdir
}

func WWPROVISIONDIR() string {
	wwlog.Printf(wwlog.DEBUG, "WWPROVISIONDIR = '%s'\n", wwprovisiondir)
	return wwprovisiondir
}

func VERSION() string {
	wwlog.Printf(wwlog.DEBUG, "VERSION = '%s'\n", version)
	return version
}

func RELEASE() string {
	wwlog.Printf(wwlog.DEBUG, "RELEASE = '%s'\n", release)
	return release
}
