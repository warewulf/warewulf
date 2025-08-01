package configure

import (
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func WAREWULFD() (err error) {
	controller := warewulfconf.Get()
	if controller.Warewulf.SystemdName != "" {
		wwlog.Info("Enabling and restarting the Warewulf server")
		err = util.SystemdStart(controller.Warewulf.SystemdName)
		if err != nil {
			return err
		}
	} else {
		wwlog.Warn("Not (re)starting Warewulf server: no systemd name configured")
	}

	return nil
}
