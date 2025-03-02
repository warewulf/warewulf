package nodedb

import "github.com/warewulf/warewulf/internal/pkg/wwlog"

func Reload() {
	if err := LoadNodeDB(); err != nil {
		wwlog.Error("Could not load node DB: %s", err)
	}

	if err := LoadNodeStatus(); err != nil {
		wwlog.Error("Could not prepopulate node status DB: %s", err)
	}
}
