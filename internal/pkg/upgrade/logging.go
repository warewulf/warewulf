package upgrade

import (
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func logIgnore(name string, value interface{}, reason string) {
	wwlog.Warn("ignore: %s: %v (%s)", name, value, reason)
}

func warnError(err error) {
	if err != nil {
		wwlog.Warn("%s", err)
	}
}
