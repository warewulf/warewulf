package wwlog

import (
	"fmt"
	"os"
	"io"
	"strings"
	"time"
	"runtime"
	"reflect"
)

const (
	SECCRITICAL = 51
	CRITICAL    = 50
	SECERROR    = 41
	ERROR       = 40
	SECWARN     = 31
	WARN        = 30
	SECINFO     = 21
	INFO        = 20
	SECVERBOSE  = 16
	VERBOSE     = 15
	SECDEBUG    = 11
	DEBUG       = 10
	NOTSET      = 0
)

type LogRecord struct {
	Level int
	Err error
	Msg string
	Args []interface{}
	Pc uintptr
	File string
	Line int
	Time time.Time
}

/*
	Format a log message from a record
	rec.level >= logLevel
*/
type LogFormatter func(logLevel int, rec *LogRecord) string

/*
	Get string level name for level number
*/
func LevelName(level int) string {
	if level >= SECCRITICAL {
    return "SECCRITICAL"
  }
	if level >= CRITICAL {
    return "CRITICAL"
  }
	if level >= SECERROR {
    return "SECERROR"
  }
	if level >= ERROR {
    return "ERROR"
  }
	if level >= SECWARN {
    return "SECWARN"
  }
	if level >= WARN {
    return "WARN"
  }
	if level >= SECINFO {
    return "SECINFO"
  }
	if level >= INFO {
    return "INFO"
  }
	if level >= SECVERBOSE {
    return "SECVERBOSE"
  }
	if level >= VERBOSE {
    return "VERBOSE"
  }
	if level >= SECDEBUG {
    return "SECDEBUG"
  }
	if level >= DEBUG {
    return "DEBUG"
  }

	return "NOTSET"
}

func DefaultFormatter(logLevel int, rec *LogRecord) string {
	message := fmt.Sprintf(rec.Msg, rec.Args...)

	if ( !strings.HasSuffix(message, "\n") ) {
		// ensure written messages are separated by at least one newline
		message += "\n"
	}

	if rec.Err != nil {
		if logLevel <= VERBOSE {
			// when debugging errors, add file and line number, and any stack trace
			message += fmt.Sprintf("%s:%d\n%+v\n", rec.File, rec.Line, rec.Err )

		}else{
			message += fmt.Sprintf("%v\n", rec.Err )
		}
	}

	if rec.Level == INFO && logLevel == INFO {
		// NOTE: this is a bit strange, but for user-friendliness it makes sense
		// to not pollute the messages unless something bad happens (by default).
		// if logLevel > INFO, then level == INFO messages should not be printed anyway,
		// and if logLevel < INFO, then it seems like all messages should get prefixed
		return message
	}

	return fmt.Sprintf("%-7s: %s", LevelName(rec.Level), message)
}

var logLevel = INFO
var logOut io.Writer = os.Stdout
var logErr io.Writer = os.Stderr
var logFormatter LogFormatter = DefaultFormatter

func EnabledForLevel(level int) bool {
	return level >= logLevel
}

/*
Set the central log level. Uneven values are security related
lof entries.
*/
func SetLogLevel(level int) {
	logLevel = level

	Debug("Set log level to: %d, %s", logLevel, LevelName(logLevel))
}

func GetLogLevel() int {
	return logLevel
}

/*
Set the log output writers
By default they are set to os.Stdout and os.Stderr
*/
func SetLogWriters(out io.Writer, err io.Writer) {
	logOut = out
	logErr = err
	Debug("Set log writers")
}

func GetLogWriters() (io.Writer, io.Writer) {
	return logOut, logErr
}

/*
Set the log record formatter
By default this is set to DefaultFormatter
*/
func SetLogFormatter(formatter LogFormatter) {
	logFormatter = formatter
	Debug("Set log formatter: %s", runtime.FuncForPC(reflect.ValueOf(formatter).Pointer()).Name())
}

func GetLogFormatter() LogFormatter {
	return logFormatter
}

/*
	Internal method to create a log record
*/
func recordLog(level int, err error, message string, a ...interface{}) {

	if EnabledForLevel(level) {
		pc, file, line, ok := runtime.Caller(2)
		if !ok {
			file = "[unknown]"
		}

		rec := LogRecord{
			Level : level,
 			Err : err,
			Msg : message,
			Args : a,
			Pc : pc,
 			File : file,
 			Line : line,
			Time : time.Now() }

		message = logFormatter(logLevel, &rec)

		if level >= ERROR {
			fmt.Fprintf(logErr, message)
		} else {
			fmt.Fprintf(logOut, message)
		}
	}
}

func Println(level int, message string) {
	recordLog(level, nil, message)
}

func Printf(level int, message string, a ...interface{}) {
	recordLog(level, nil, message, a...)
}

/*******************************************************************************
	Named log level functions
*/
func Log(level int, message string, a ...interface{}) {
	recordLog(level, nil, message, a...)
}

func LogExc(level int, err error, message string, a ...interface{}) {
	recordLog(level, err, message, a...)
}

func Debug(message string, a ...interface{}) {
	recordLog(DEBUG, nil, message, a...)
}

func DebugExc(err error, message string, a ...interface{}) {
	recordLog(DEBUG, err, message, a...)
}

func SecDebug(message string, a ...interface{}) {
	recordLog(SECDEBUG, nil, message, a...)
}

func Verbose(message string, a ...interface{}) {
	recordLog(VERBOSE, nil, message, a...)
}

func VerboseExc(err error, message string, a ...interface{}) {
	recordLog(VERBOSE, err, message, a...)
}

func SecVerbose(message string, a ...interface{}) {
	recordLog(SECVERBOSE, nil, message, a...)
}

func Info(message string, a ...interface{}) {
	recordLog(INFO, nil, message, a...)
}

func InfoExc(err error, message string, a ...interface{}) {
	recordLog(INFO, err, message, a...)
}

func SecInfo(message string, a ...interface{}) {
	recordLog(SECINFO, nil, message, a...)
}

func Warn(message string, a ...interface{}) {
	recordLog(WARN, nil, message, a...)
}

func WarnExc(err error, message string, a ...interface{}) {
	recordLog(WARN, err, message, a...)
}

func SecWarn(message string, a ...interface{}) {
	recordLog(SECWARN, nil, message, a...)
}

func Error(message string, a ...interface{}) {
	recordLog(ERROR, nil, message, a...)
}

func ErrorExc(err error, message string, a ...interface{}) {
	recordLog(ERROR, err, message, a...)
}

func SecError(message string, a ...interface{}) {
	recordLog(SECERROR, nil, message, a...)
}

func Critical(message string, a ...interface{}) {
	recordLog(CRITICAL, nil, message, a...)
}

func CriticalExc(err error, message string, a ...interface{}) {
	recordLog(CRITICAL, err, message, a...)
}

func SecCritical(message string, a ...interface{}) {
	recordLog(SECCRITICAL, nil, message, a...)
}
