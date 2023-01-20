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
	wwlog.Debug("BINDIR = '%s'", bindir)
	return bindir
}

func DATADIR() string {
	wwlog.Debug("DATADIR = '%s'", datadir)
	return datadir
}

func SYSCONFDIR() string {
	wwlog.Debug("SYSCONFDIR = '%s'", sysconfdir)
	return sysconfdir
}

func LOCALSTATEDIR() string {
	wwlog.Debug("LOCALSTATEDIR = '%s'", localstatedir)
	return localstatedir
}

func SRVDIR() string {
	wwlog.Debug("SRVDIR = '%s'", srvdir)
	return srvdir
}

func TFTPDIR() string {
	wwlog.Debug("TFTPDIR = '%s'", tftpdir)
	return tftpdir
}

func FIREWALLDDIR() string {
	wwlog.Debug("FIREWALLDDIR = '%s'", firewallddir)
	return firewallddir
}

func SYSTEMDDIR() string {
	wwlog.Debug("SYSTEMDDIR = '%s'", systemddir)
	return systemddir
}

func WWOVERLAYDIR() string {
	wwlog.Debug("WWOVERLAYDIR = '%s'", wwoverlaydir)
	return wwoverlaydir
}

func WWCHROOTDIR() string {
	wwlog.Debug("WWCHROOTDIR = '%s'", wwchrootdir)
	return wwchrootdir
}

func WWPROVISIONDIR() string {
	wwlog.Debug("WWPROVISIONDIR = '%s'", wwprovisiondir)
	return wwprovisiondir
}

func VERSION() string {
	wwlog.Debug("VERSION = '%s'", version)
	return version
}

func RELEASE() string {
	wwlog.Debug("RELEASE = '%s'", release)
	return release
}

func WWCLIENTDIR() string {
	wwlog.Debug("WWCLIENTDIR = '%s'", wwclientdir)
	return wwclientdir
}
