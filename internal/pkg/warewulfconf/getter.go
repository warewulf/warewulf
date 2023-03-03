package warewulfconf

import "github.com/hpcng/warewulf/internal/pkg/wwlog"

func (conf *ControllerConf) BINDIR() string {
	wwlog.Debug("BINDIR = '%s'", conf.Paths.bindir)
	return conf.Paths.bindir
}

func (conf *ControllerConf) DATADIR() string {
	wwlog.Debug("DATADIR = '%s'", conf.Paths.datadir)
	return conf.Paths.datadir
}

func (conf *ControllerConf) SYSCONFDIR() string {
	wwlog.Debug("SYSCONFDIR = '%s'", conf.Paths.sysconfdir)
	return conf.Paths.sysconfdir
}

func (conf *ControllerConf) LOCALSTATEDIR() string {
	wwlog.Debug("LOCALSTATEDIR = '%s'", conf.Paths.localstatedir)
	return conf.Paths.localstatedir
}

func (conf *ControllerConf) SRVDIR() string {
	wwlog.Debug("SRVDIR = '%s'", conf.Paths.srvdir)
	return conf.Paths.srvdir
}

func (conf *ControllerConf) TFTPDIR() string {
	wwlog.Debug("TFTPDIR = '%s'", conf.Paths.tftpdir)
	return conf.Paths.tftpdir
}

func (conf *ControllerConf) FIREWALLDDIR() string {
	wwlog.Debug("FIREWALLDDIR = '%s'", conf.Paths.firewallddir)
	return conf.Paths.firewallddir
}

func (conf *ControllerConf) SYSTEMDDIR() string {
	wwlog.Debug("SYSTEMDDIR = '%s'", conf.Paths.systemddir)
	return conf.Paths.systemddir
}

func (conf *ControllerConf) WWOVERLAYDIR() string {
	wwlog.Debug("WWOVERLAYDIR = '%s'", conf.Paths.wwoverlaydir)
	return conf.Paths.wwoverlaydir
}

func (conf *ControllerConf) WWCHROOTDIR() string {
	wwlog.Debug("WWCHROOTDIR = '%s'", conf.Paths.wwchrootdir)
	return conf.Paths.wwchrootdir
}

func (conf *ControllerConf) WWPROVISIONDIR() string {
	wwlog.Debug("WWPROVISIONDIR = '%s'", conf.Paths.wwprovisiondir)
	return conf.Paths.wwprovisiondir
}

func (conf *ControllerConf) VERSION() string {
	wwlog.Debug("VERSION = '%s'", conf.Paths.version)
	return conf.Paths.version
}

func (conf *ControllerConf) RELEASE() string {
	wwlog.Debug("RELEASE = '%s'", conf.Paths.release)
	return conf.Paths.release
}

func (conf *ControllerConf) WWCLIENTDIR() string {
	wwlog.Debug("WWCLIENTDIR = '%s'", conf.Paths.wwclientdir)
	return conf.Paths.wwclientdir
}
