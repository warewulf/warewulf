package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/pkg/errors"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

const (
	WAREWULFD_PIDFILE = "/var/run/warewulfd.pid"
)

// allow to run without daemon for tests
var nodaemon bool

func init() {
	nodaemon = false
}

// run without daemon
func SetNoDaemon() {
	nodaemon = true
}

func DaemonStatus() error {
	if nodaemon {
		return nil
	}

	if !util.IsFile(WAREWULFD_PIDFILE) {
		return errors.New("Warewulf server is not running")
	}

	dat, err := os.ReadFile(WAREWULFD_PIDFILE)
	if err != nil {
		return fmt.Errorf("could not read Warewulfd PID file: %w", err)
	}

	pid, _ := strconv.Atoi(string(dat))
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find running PID: %w", err)
	} else {
		err := process.Signal(syscall.Signal(0))
		if err != nil {
			return fmt.Errorf("failed to send process SIGCONT: %w", err)
		} else {
			wwlog.Serv("Warewulf server is running at PID: %d", pid)
		}
	}

	return nil
}

func DaemonReload() error {
	if nodaemon {
		return nil
	}
	cmd := exec.Command("/usr/sbin/service", "warewulfd", "reload")
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to reload warewulfd: %w", err)
	}
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("failed to reload warewulfd: %w", err)
	}
	return nil
}
