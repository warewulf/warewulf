package util

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/opencontainers/selinux/go-selinux"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// RestoreSELinuxContext restores the SELinux context for a path and all its children
// based on the system's SELinux policy, equivalent to running restorecon -R
func RestoreSELinuxContext(rootPath string) error {
	if !selinux.GetEnabled() {
		wwlog.Debug("SELinux not enabled, skipping context restoration")
		return nil
	}

	wwlog.Info("Restoring SELinux contexts for: %s", rootPath)

	cmd := exec.Command("restorecon", "-vR", rootPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("restorecon failed: %w: %s", err, string(output))
	}

	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line != "" {
			wwlog.Debug("restorecon output: %s", line)
		}
	}
	return nil
}
