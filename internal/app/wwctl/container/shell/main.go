//go:build linux
// +build linux

package shell

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	cntexec "github.com/warewulf/warewulf/internal/app/wwctl/container/exec"
	"github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	containerName := args[0]
	var allargs []string

	if !container.ValidSource(containerName) {
		return fmt.Errorf("unknown Warewulf container: %s", containerName)
	}
	shellName := os.Getenv("SHELL")
	if !container.ValidSource(containerName) {
		return fmt.Errorf("unknown Warewulf container: %s", containerName)
	}
	var shells []string
	if shellName == "" {
		shells = append(shells, "/bin/bash")
	} else {
		shells = append(shells, shellName, "/bin/bash")
	}
	for _, s := range shells {
		if _, err := os.Stat(path.Join(container.RootFsDir(containerName), s)); err == nil {
			shellName = s
			break
		}
	}
	args = append(args, shellName)
	allargs = append(allargs, args...)
	wwlog.Debug("Calling exec with args: %s", allargs)
	cntexec.SetBinds(binds)
	cntexec.SetNode(nodeName)
	cntexec.SyncUser = syncUser
	cntexec.Build = build
	if cntexec.Build {
		wwlog.Info("Container image build will be skipped if the shell ends with a non-zero exit code.")
	}
	return cntexec.CobraRunE(cmd, allargs)
}
