//go:build linux
// +build linux

package shell

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	cntexec "github.com/warewulf/warewulf/internal/app/wwctl/image/exec"
	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	imageName := args[0]

	if !image.ValidSource(imageName) {
		return fmt.Errorf("unknown Warewulf image: %s", imageName)
	}
	if !image.ValidSource(imageName) {
		return fmt.Errorf("unknown Warewulf image: %s", imageName)
	}

	var shells []string
	if os.Getenv("SHELL") != "" {
		shells = append(shells, os.Getenv("SHELL"))
	}
	shells = append(shells, "/bin/bash", "/bin/sh")

	shellName := ""
	for _, s := range shells {
		if util.IsFile(path.Join(image.RootFsDir(imageName), s)) {
			shellName = s
			break
		}
	}
	if shellName == "" {
		return fmt.Errorf("no shell found in image: %s", imageName)
	}

	args = append(args, shellName)
	wwlog.Debug("%s: exec with args: %v", imageName, args)
	cntexec.SetBinds(binds)
	cntexec.SetNode(nodeName)
	cntexec.SyncUser = syncUser
	cntexec.Build = build
	if cntexec.Build {
		wwlog.Info("Image build will be skipped if the shell ends with a non-zero exit code.")
	}
	return cntexec.CobraRunE(cmd, args)
}
