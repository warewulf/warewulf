package warewulfconf

import "github.com/hpcng/warewulf/internal/pkg/wwlog"

func (conf *ControllerConf) BINDIR() string {
	wwlog.Debug("BINDIR = '%s'", conf.Paths.Bindir)
	return conf.Paths.Bindir
}

func (conf *ControllerConf) DATADIR() string {
	wwlog.Debug("DATADIR = '%s'", conf.Paths.Datadir)
	return conf.Paths.Datadir
}

func (conf *ControllerConf) SYSCONFDIR() string {
	wwlog.Debug("SYSCONFDIR = '%s'", conf.Paths.Sysconfdir)
	return conf.Paths.Sysconfdir
}

func (conf *ControllerConf) LOCALSTATEDIR() string {
	wwlog.Debug("LOCALSTATEDIR = '%s'", conf.Paths.Localstatedir)
	return conf.Paths.Localstatedir
}

func (conf *ControllerConf) SRVDIR() string {
	wwlog.Debug("SRVDIR = '%s'", conf.Paths.Srvdir)
	return conf.Paths.Srvdir
}

func (conf *ControllerConf) TFTPDIR() string {
	wwlog.Debug("TFTPDIR = '%s'", conf.Paths.Tftpdir)
	return conf.Paths.Tftpdir
}

func (conf *ControllerConf) FIREWALLDDIR() string {
	wwlog.Debug("FIREWALLDDIR = '%s'", conf.Paths.Firewallddir)
	return conf.Paths.Firewallddir
}

func (conf *ControllerConf) SYSTEMDDIR() string {
	wwlog.Debug("SYSTEMDDIR = '%s'", conf.Paths.Systemddir)
	return conf.Paths.Systemddir
}

func (conf *ControllerConf) WWOVERLAYDIR() string {
	wwlog.Debug("WWOVERLAYDIR = '%s'", conf.Paths.Wwoverlaydir)
	return conf.Paths.Wwoverlaydir
}

func (conf *ControllerConf) WWCHROOTDIR() string {
	wwlog.Debug("WWCHROOTDIR = '%s'", conf.Paths.Wwchrootdir)
	return conf.Paths.Wwchrootdir
}

func (conf *ControllerConf) WWPROVISIONDIR() string {
	wwlog.Debug("WWPROVISIONDIR = '%s'", conf.Paths.Wwprovisiondir)
	return conf.Paths.Wwprovisiondir
}

func (conf *ControllerConf) VERSION() string {
	wwlog.Debug("VERSION = '%s'", conf.Paths.Version)
	return conf.Paths.Version
}

func (conf *ControllerConf) RELEASE() string {
	wwlog.Debug("RELEASE = '%s'", conf.Paths.Release)
	return conf.Paths.Release
}

func (conf *ControllerConf) WWCLIENTDIR() string {
	wwlog.Debug("WWCLIENTDIR = '%s'", conf.Paths.Wwclientdir)
	return conf.Paths.Wwclientdir
}
