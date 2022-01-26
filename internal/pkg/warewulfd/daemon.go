package warewulfd

import (
	"fmt"
	"io/ioutil"
	"log"
	"log/syslog"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/version"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/pkg/errors"
)

const (
	WAREWULFD_PIDFILE = "/var/run/warewulfd.pid"
	WAREWULFD_LOGFILE = "/var/log/warewulfd.log"
)

var logwriter *syslog.Writer
var loginit bool

func daemonLogf(message string, a ...interface{}) {
	conf, err := warewulfconf.New()
	if err != nil {
		fmt.Printf("ERROR: Could not read Warewulf configuration file: %s\n", err)
		return
	}

	if conf.Warewulf.Syslog {
		if !loginit {
			var err error

			logwriter, err = syslog.New(syslog.LOG_NOTICE, "warewulfd")
			if err != nil {
				return
			}
			log.SetOutput(logwriter)
			loginit = true

			log.SetFlags(0)
			log.SetPrefix("")
		}

		log.Printf(message, a...)

	} else {
		prefix := fmt.Sprintf("[%s] ", time.Now().Format(time.UnixDate))
		fmt.Printf(prefix+message, a...)
	}
}

func DaemonStart() error {
	if os.Getenv("WAREWULFD_BACKGROUND") == "1" {
		err := RunServer()
		if err != nil {
			return errors.Wrap(err, "failed to run server")
		}

	} else {
		if util.IsFile(WAREWULFD_PIDFILE) {
			return errors.New("process is already running")
		}

		os.Setenv("WAREWULFD_BACKGROUND", "1")

		f, err := os.OpenFile(WAREWULFD_LOGFILE, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return err
		}

		p, err := os.OpenFile(WAREWULFD_PIDFILE, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer p.Close()

		cmd := exec.Command(os.Args[0], "server", "start")
		cmd.Stdout = f
		cmd.Stderr = f
		err = cmd.Start()
		if err != nil {
			return errors.Wrap(err, "failed to start command")
		}
		pid := cmd.Process.Pid

		fmt.Fprintf(p, "%d", pid)

		fmt.Printf("Started Warewulf (%s) server at PID: %d\n", version.GetVersion(), pid)

	}

	return nil
}

func DaemonStatus() error {
	if !util.IsFile(WAREWULFD_PIDFILE) {
		return errors.New("Warewulf server is not running")
	}

	dat, err := ioutil.ReadFile(WAREWULFD_PIDFILE)
	if err != nil {
		return errors.Wrap(err, "could not read Warewulfd PID file")
	}

	pid, _ := strconv.Atoi(string(dat))
	process, err := os.FindProcess(pid)
	if err != nil {
		return errors.Wrap(err, "failed to find running PID")
	} else {
		err := process.Signal(syscall.Signal(0))
		if err != nil {
			return errors.Wrap(err, "failed to send process SIGCONT")
		} else {
			fmt.Printf("Warewulf server is running at PID: %d\n", pid)
		}
	}

	return nil
}

func DaemonReload() error {
	if !util.IsFile(WAREWULFD_PIDFILE) {
		return errors.New("Warewulf server is not running")
	}

	dat, err := ioutil.ReadFile(WAREWULFD_PIDFILE)
	if err != nil {
		return errors.Wrap(err, "could not read Warewulfd PID file")
	}

	pid, _ := strconv.Atoi(string(dat))
	process, err := os.FindProcess(pid)
	if err != nil {
		return errors.Wrap(err, "failed to find running PID")
	} else {
		err := process.Signal(syscall.Signal(syscall.SIGHUP))
		if err != nil {
			return errors.Wrap(err, "failed to send process SIGHUP")
		}
	}

	return nil
}

func DaemonStop() error {
	if !util.IsFile(WAREWULFD_PIDFILE) {
		fmt.Printf("Warewulf daemon process not running\n")
		return nil
	}

	dat, err := ioutil.ReadFile(WAREWULFD_PIDFILE)
	if err != nil {
		return err
	}

	_ = os.Remove(WAREWULFD_PIDFILE)

	pid, _ := strconv.Atoi(string(dat))
	process, err := os.FindProcess(pid)

	if err != nil {
		return errors.Wrap(err, "failed to find running PID")
	} else {
		err := process.Signal(syscall.Signal(15))
		if err != nil {
			return errors.Wrap(err, "failed to send process SIGTERM")
		} else {
			fmt.Printf("Terminated Warewulf server at PID: %d\n", pid)
		}
	}

	return nil
}
