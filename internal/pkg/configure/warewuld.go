package configure

import (
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func WAREWULFD() (err error) {
	controller := warewulfconf.Get()
	wwlog.Info("Enabling and restarting the WAREWULFD service")
	err = util.SystemdStart(controller.Warewulf.SystemdName)
	if err != nil {
		return
	}
	return nil
}
