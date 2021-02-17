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

	if container.ValidSource(containerName) == false {
		wwlog.Printf(wwlog.ERROR, "Unknown Warewulf container: %s\n", containerName)
		os.Exit(1)
	}

	c := exec.Command("/proc/self/exe", append([]string{"container", "exec", "__child"}, args...)...)

	//c := exec.Command("/bin/sh")
	c.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}

	fmt.Printf("Rebuilding container...\n")
	output, err := container.Build(containerName, false)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not build container %s: %s\n", containerName, err)
		os.Exit(1)
	} else {
		fmt.Printf("%s\n", output)
	}

	return nil
}
