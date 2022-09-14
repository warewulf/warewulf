package buildconfig

import "github.com/hpcng/warewulf/internal/pkg/wwlog"

const WWVer = 43

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
	wwclientdir    string = "UNDEF"
	datadir        string = "UNDEF"
)

func BINDIR() string {
	wwlog.Debug("BINDIR = '%s'\n", bindir)
	return bindir
}

func DATADIR() string {
	wwlog.Debug("DATADIR = '%s'\n", bindir)
	return datadir
}

func SYSCONFDIR() string {
	wwlog.Debug("SYSCONFDIR = '%s'\n", sysconfdir)
	return sysconfdir
}

func LOCALSTATEDIR() string {
	wwlog.Debug("LOCALSTATEDIR = '%s'\n", localstatedir)
	return localstatedir
}

func SRVDIR() string {
	wwlog.Debug("SRVDIR = '%s'\n", srvdir)
	return srvdir
}

func TFTPDIR() string {
	wwlog.Debug("TFTPDIR = '%s'\n", tftpdir)
	return tftpdir
}

func FIREWALLDDIR() string {
	wwlog.Debug("FIREWALLDDIR = '%s'\n", firewallddir)
	return firewallddir
}

func SYSTEMDDIR() string {
	wwlog.Debug("SYSTEMDDIR = '%s'\n", systemddir)
	return systemddir
}

func WWOVERLAYDIR() string {
	wwlog.Debug("WWOVERLAYDIR = '%s'\n", wwoverlaydir)
	return wwoverlaydir
}

func WWCHROOTDIR() string {
	wwlog.Debug("WWCHROOTDIR = '%s'\n", wwchrootdir)
	return wwchrootdir
}

func WWPROVISIONDIR() string {
	wwlog.Debug("WWPROVISIONDIR = '%s'\n", wwprovisiondir)
	return wwprovisiondir
}

func VERSION() string {
	wwlog.Debug("VERSION = '%s'\n", version)
	return version
}

func RELEASE() string {
	wwlog.Debug("RELEASE = '%s'\n", release)
	return release
}

func WWCLIENTDIR() string {
	wwlog.Debug("WWCLIENTDIR = '%s'\n", wwclientdir)
	return wwclientdir
}
