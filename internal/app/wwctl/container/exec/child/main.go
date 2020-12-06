// + build linux

package child

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/container"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
	"syscall"
)

func CobraRunE(cmd *cobra.Command, args []string) error {

	if os.Getpid() != 1 {
		wwlog.Printf(wwlog.ERROR, "PID is not 1: %d\n", os.Getpid())
		os.Exit(1)
	}

	containerName := args[0]

	if container.ValidSource(containerName) == false {
		wwlog.Printf(wwlog.ERROR, "Unknown Warewulf container: %s\n", containerName)
		os.Exit(1)
	}

	syscall.Mount("", "/", "", syscall.MS_PRIVATE, "")

	containerPath := container.RootFsDir(containerName)

	syscall.Chroot(containerPath)
	os.Chdir("/")

	syscall.Mount("rootfs", "rootfs", "", syscall.MS_BIND, "")
	syscall.Mount("/proc", "/proc", "proc", 0, "")

	ps1string := fmt.Sprintf("[%s] Warewulf> ", containerName)
	os.Setenv("PS1", ps1string)

	err := syscall.Exec(args[1], args[1:], os.Environ())
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	return nil
}
