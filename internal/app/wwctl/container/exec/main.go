//go:build linux
// +build linux

package exec

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func runContainedCmd(args []string) error {
	wwlog.Printf(wwlog.VERBOSE, "Running contained command: %s\n", args[1:])
	c := exec.Command("/proc/self/exe", append([]string{"container", "exec", "__child"}, args...)...)

	c.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return nil
}

func CobraRunE(cmd *cobra.Command, args []string) error {

	containerName := args[0]
	var allargs []string

	if !container.ValidSource(containerName) {
		wwlog.Printf(wwlog.ERROR, "Unknown Warewulf container: %s\n", containerName)
		os.Exit(1)
	}

	for _, b := range binds {
		allargs = append(allargs, "--bind", b)
	}
	allargs = append(allargs, args...)

	err := runContainedCmd(allargs)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Failed executing container command: %s\n", err)
		os.Exit(1)
	}

	if util.IsFile(path.Join(container.RootFsDir(allargs[0]), "/etc/warewulf/container_exit.sh")) {
		wwlog.Printf(wwlog.VERBOSE, "Found clean script: /etc/warewulf/container_exit.sh\n")
		err = runContainedCmd([]string{allargs[0], "/bin/sh", "/etc/warewulf/container_exit.sh"})
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Failed executing exit script: %s\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Rebuilding container...\n")
	err = container.Build(containerName, false)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not build container %s: %s\n", containerName, err)
		os.Exit(1)
	}

	return nil
}
