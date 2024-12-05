package upgrade

import (
	"os"
	"path/filepath"

	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func legacyKernelVersion(kernelName string) string {
	wwlog.Debug("legacyKernelVersion(%v)", kernelName)
	if kernelName == "" {
		return ""
	}
	kernelVersion, err := os.ReadFile(legacyKernelVersionFile(kernelName))
	if err != nil {
		return ""
	}
	wwlog.Debug("legacyKernelVersion(%v) -> %v", kernelName, string(kernelVersion))
	return string(kernelVersion)
}

func legacyKernelVersionFile(kernelName string) string {
	if kernelName == "" {
		return ""
	}

	if !util.ValidString(kernelName, "^[a-zA-Z0-9-._]+$") {
		return ""
	}

	return filepath.Join(legacyKernelImageDir(kernelName), "version")
}

func legacyKernelImageDir(name string) string {
	return filepath.Join(legacyKernelImageTopDir(), name)
}

func legacyKernelImageTopDir() string {
	return filepath.Join(config.Get().Paths.WWProvisiondir, "kernel")
}
