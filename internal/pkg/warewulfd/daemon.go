package warewulfd

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

const (
	WAREWULFD_PIDFILE = "/tmp/warewulfd.pid"
	WAREWULFD_LOGFILE = "/tmp/warewulfd.log"
)

func DaemonStart() error {

	if os.Getenv("WAREWULFD_BACKGROUND") == "1" {
		RunServer()

	} else {
		os.Setenv("WAREWULFD_BACKGROUND", "1")

		f, err := os.OpenFile(WAREWULFD_LOGFILE, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return err
		}

		p, err := os.OpenFile(WAREWULFD_PIDFILE, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}

		cmd := exec.Command(os.Args[0], "server", "start")
		cmd.Stdout = f
		cmd.Stderr = f
		cmd.Start()
		pid := cmd.Process.Pid

		fmt.Fprintf(p, "%d", pid)

		p.Close()

		time.Sleep(1 * time.Second)

		DaemonStatus()
	}

	return nil
}

func DaemonStatus() error {

	if util.IsFile(WAREWULFD_PIDFILE) == false {
		wwlog.Printf(wwlog.INFO, "Warewulf daemon process not running (%s)\n", WAREWULFD_PIDFILE)
		return nil
	}

	dat, err := ioutil.ReadFile(WAREWULFD_PIDFILE)
	if err != nil {
		return err
	}

	pid, _ := strconv.Atoi(string(dat))
	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("Failed to find process: %s\n", err)
		return err
	} else {
		err := process.Signal(syscall.Signal(0))
		if err != nil {
			fmt.Printf("SIGCONT on pid %d returned: %v\n", pid, err)
			return err
		} else {
			fmt.Printf("Warewulf daemon is running at PID: %d\n", pid)
		}
	}

	return nil
}

func DaemonReload() error {
	if util.IsFile(WAREWULFD_PIDFILE) == false {
		wwlog.Printf(wwlog.INFO, "Warewulf daemon process not running (%s)\n", WAREWULFD_PIDFILE)
		return nil
	}

	dat, err := ioutil.ReadFile(WAREWULFD_PIDFILE)
	if err != nil {
		return err
	}

	pid, _ := strconv.Atoi(string(dat))
	process, err := os.FindProcess(pid)

	if err != nil {
		fmt.Printf("Failed to find process: %s\n", err)
		return err
	} else {
		err := process.Signal(syscall.Signal(syscall.SIGHUP))
		if err != nil {
			fmt.Printf("SIGCONT on pid %d returned: %v\n", pid, err)
			return err
		}
	}

	return nil
}

func DaemonStop() error {

	if util.IsFile(WAREWULFD_PIDFILE) == false {
		wwlog.Printf(wwlog.INFO, "Warewulf daemon process not running (%s)\n", WAREWULFD_PIDFILE)
		return nil
	}

	dat, err := ioutil.ReadFile(WAREWULFD_PIDFILE)
	if err != nil {
		return err
	}

	pid, _ := strconv.Atoi(string(dat))
	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("Failed to find process: %s\n", err)
	} else {
		err := process.Signal(syscall.Signal(15))
		if err != nil {
			fmt.Printf("SIGCONT on pid %d returned: %v\n", pid, err)
		} else {
			fmt.Printf("Terminated Warewulf process at PID: %d\n", pid)
		}
	}

	os.Remove(WAREWULFD_PIDFILE)

	return nil
}
