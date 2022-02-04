//go:build linux
// +build linux

package exec

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
)

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
	if len(args) == 1 {
		allargs = append(allargs, "/usr/bin/bash")
	}

	c := exec.Command("/proc/self/exe", append([]string{"container", "exec", "__child"}, allargs...)...)

	//c := exec.Command("/bin/sh")
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

	fmt.Printf("Rebuilding container...\n")
	err := container.Build(containerName, false)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not build container %s: %s\n", containerName, err)
		os.Exit(1)
	}

	return nil
}
