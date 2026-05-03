package capture

import (
	"path/filepath"
	"strings"

	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/overlaydiff"
)

const snapshotFileName = "capture.json"

func defaultStateFilePath() string {
	conf := config.Get()
	base := conf.Paths.Localstatedir
	if base == "" {
		base = "/var/lib"
	}
	return filepath.Join(base, "warewulf", "overlaydiff", snapshotFileName)
}

func resolveStateFilePath(custom string) string {
	if strings.TrimSpace(custom) == "" {
		return defaultStateFilePath()
	}
	return custom
}

func defaultExcludeHelp() string {
	return "Exclude path prefix from scan (repeatable). Defaults include: " + strings.Join(overlaydiff.DefaultExcludes, ", ")
}
