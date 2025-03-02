package warewulfd

import (
	"os"
	"strconv"
	"time"

	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

const (
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
	}

	loginit = true

	return nil
}
