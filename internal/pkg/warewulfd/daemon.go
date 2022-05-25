package warewulfd

import (
	"fmt"
	"io/ioutil"
	"log/syslog"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/version"
	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/pkg/errors"
)

const (
	WAREWULFD_PIDFILE = "/var/run/warewulfd.pid"
	WAREWULFD_LOGFILE = "/var/log/warewulfd.log"
)

var loginit bool

func DaemonFormatter(logLevel int, rec *wwlog.LogRecord) string {
	return "[" + rec.Time.Format(time.UnixDate) + "] " + wwlog.DefaultFormatter(logLevel, rec)
}

func DaemonInitLogging() error {
	if loginit {
		return nil
	}

	wwlog.SetLogFormatter(DaemonFormatter)

	level_str, ok := os.LookupEnv("WAREWULFD_LOGLEVEL")
	if ok {
		level, err := strconv.Atoi(level_str)
		if err == nil {
			wwlog.SetLogLevel(level)
		}
	}else{
		wwlog.SetLogLevel(wwlog.SERV)
	}

	conf, err := warewulfconf.New()
	if err != nil {
		return errors.Wrap(err, "Could not read Warewulf configuration file")
	}

	if conf.Warewulf.Syslog {

		wwlog.Debug("Changingq log output to syslog")

		logwriter, err := syslog.New(syslog.LOG_NOTICE, "warewulfd")
		if err != nil {
			return errors.Wrap(err, "Could not create syslog writer")
		}

		wwlog.SetLogFormatter(wwlog.DefaultFormatter)
		wwlog.SetLogWriters(logwriter, logwriter)

	}

	loginit = true

	return nil
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

		logLevel := wwlog.GetLogLevel()
		if logLevel == wwlog.INFO {
			os.Setenv("WAREWULFD_LOGLEVEL", strconv.Itoa(wwlog.SERV))
		}else{
			os.Setenv("WAREWULFD_LOGLEVEL", strconv.Itoa(logLevel))
		}

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

		wwlog.Serv("Started Warewulf (%s) server at PID: %d", version.GetVersion(), pid)

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
			wwlog.Serv("Warewulf server is running at PID: %d", pid)
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

	logLevel := wwlog.GetLogLevel()
	if logLevel == wwlog.INFO {
		os.Setenv("WAREWULFD_LOGLEVEL", strconv.Itoa(wwlog.SERV))
	}else{
		os.Setenv("WAREWULFD_LOGLEVEL", strconv.Itoa(logLevel))
	}

	return nil
}

func DaemonStop() error {
	if !util.IsFile(WAREWULFD_PIDFILE) {
		wwlog.Warn("Warewulf daemon process not running")
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
			wwlog.Serv("Terminated Warewulf server at PID: %d", pid)
		}
	}

	return nil
}
