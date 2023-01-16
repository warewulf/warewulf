//go:build linux
// +build linux

package shell

import (
	"os"

	cntexec "github.com/hpcng/warewulf/internal/app/wwctl/container/exec"
	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	containerName := args[0]
	var allargs []string

	if !container.ValidSource(containerName) {
		wwlog.Error("Unknown Warewulf container: %s", containerName)
		os.Exit(1)
	}
	/*
		for _, b := range binds {
			allargs = append(allargs, "--bind", b)
		}
	*/
	args = append(args, "/usr/bin/bash")
	allargs = append(allargs, args...)
	wwlog.Debug("Calling exec with args: %s", allargs)
	cntexec.SetBinds(binds)
	return cntexec.CobraRunE(cmd, allargs)
	/*
		c := exec.Command("/proc/self/exe", append([]string{"container", "exec"}, allargs...)...)

		//c := exec.Command("/bin/sh")
		c.SysProcAttr = &syscall.SysProcAttr{
			Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		}
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		os.Setenv("WW_CONTAINER_SHELL", containerName)

		if err := c.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	*/
	return nil
}
