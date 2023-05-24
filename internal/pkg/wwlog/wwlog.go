package wwlog

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"
)

type LogRecord struct {
	Level int
	Err   error
	Msg   string
	Args  []interface{}
	Pc    uintptr
	File  string
	Line  int
	Time  time.Time
}

/*
Format a log message from a record
Only called if rec.level >= logLevel
*/
type LogFormatter func(logLevel int, rec *LogRecord) string

var (
	SECCRITICAL = SetLevelName(51, "SECCRITICAL")
	CRITICAL    = SetLevelName(50, "CRITICAL")
	DENIED      = SetLevelName(42, "DENIED")
	SECERROR    = SetLevelName(41, "SECERROR")
	ERROR       = SetLevelName(40, "ERROR")
	SECWARN     = SetLevelName(31, "SECWARN")
	WARN        = SetLevelName(30, "WARN")
	SEND        = SetLevelName(27, "SEND")
	RECV        = SetLevelName(26, "RECV")
	SERV        = SetLevelName(25, "SERV")
	SECINFO     = SetLevelName(21, "SECINFO")
	INFO        = SetLevelName(20, "INFO")
	SECVERBOSE  = SetLevelName(16, "SECVERBOSE")
	VERBOSE     = SetLevelName(15, "VERBOSE")
	SECDEBUG    = SetLevelName(11, "SECDEBUG")
	DEBUG       = SetLevelName(10, "DEBUG")
)

var (
	levelNums                 = []int{0}
	levelNames                = []string{"NOTSET"}
	logLevel                  = INFO
	logErr       io.Writer    = os.Stderr
	logFormatter LogFormatter = DefaultFormatter
)

func LevelNameEff(level int) (int, int, string) {
	n := len(levelNums)
	idx := sort.SearchInts(levelNums, level)

	if idx >= n {
		idx = n - 1
	}

	eff_level := levelNums[idx]
	eff_name := levelNames[idx]

	return idx, eff_level, eff_name
}

func LevelName(level int) string {
	_, _, name := LevelNameEff(level)
	return name
}

func SetLevelName(level int, name string) int {
	n := len(levelNums)
	idx := sort.SearchInts(levelNums, level)

	if idx < n && levelNums[idx] == level {
		levelNames[idx] = name
	} else {

		levelNums = append(levelNums, level)
		levelNames = append(levelNames, name)

		if idx < n {
			copy(levelNums[idx+1:], levelNums[idx:])
			copy(levelNames[idx+1:], levelNames[idx:])

			levelNums[idx] = level
			levelNames[idx] = name
		}
	}

	return level
}

func DefaultFormatter(logLevel int, rec *LogRecord) string {
	message := fmt.Sprintf(rec.Msg, rec.Args...)

	if !strings.HasSuffix(message, "\n") {
		// ensure written messages are separated by at least one newline
		message += "\n"
	}

	if rec.Err != nil {
		if logLevel < VERBOSE {
			// when debugging errors, add file and line number, and any stack trace
			message += fmt.Sprintf("%s:%d\n%+v\n", rec.File, rec.Line, rec.Err)
		} else {
			message += fmt.Sprintf("%v\n", rec.Err)
		}
	}

	if rec.Level == INFO && logLevel == INFO {
		// NOTE: this is a bit strange, but for user-friendliness it makes sense
		// to not pollute the messages unless something bad happens (by default).
		// if logLevel > INFO, then level == INFO messages should not be printed anyway,
		// and if logLevel < INFO, then it seems like all messages should get prefixed
		return message
	}

	name := LevelName(rec.Level)

	if len(name) <= 7 {
		return fmt.Sprintf("%-7s: %s", name, message)
	}

	return fmt.Sprintf("%-11s: %s", name, message)
}

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
Set the log output writer
By default they are set to output writer
*/
func SetLogWriter(err io.Writer) {
	logErr = err
}

func GetLogWriter() io.Writer {
	return logErr
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
func LogCaller(level int, skip int, err error, message string, a ...interface{}) {
	if EnabledForLevel(level) {
		pc, file, line, ok := runtime.Caller(skip + 1)
		if !ok {
			file = "[unknown]"
		}

		rec := LogRecord{
			Level: level,
			Err:   err,
			Msg:   message,
			Args:  a,
			Pc:    pc,
			File:  file,
			Line:  line,
			Time:  time.Now(),
		}

		message = logFormatter(logLevel, &rec)

		fmt.Fprint(logErr, message)
	}
}

func Println(level int, message string) {
	LogCaller(level, 1, nil, message)
}

func Printf(level int, message string, a ...interface{}) {
	LogCaller(level, 1, nil, message, a...)
}

/*
******************************************************************************
Named log level functions
*/
func Log(level int, message string, a ...interface{}) {
	LogCaller(level, 1, nil, message, a...)
}

func LogExc(level int, err error, message string, a ...interface{}) {
	LogCaller(level, 1, err, message, a...)
}

func Debug(message string, a ...interface{}) {
	LogCaller(DEBUG, 1, nil, message, a...)
}

func DebugExc(err error, message string, a ...interface{}) {
	LogCaller(DEBUG, 1, err, message, a...)
}

func SecDebug(message string, a ...interface{}) {
	LogCaller(SECDEBUG, 1, nil, message, a...)
}

func Verbose(message string, a ...interface{}) {
	LogCaller(VERBOSE, 1, nil, message, a...)
}

func VerboseExc(err error, message string, a ...interface{}) {
	LogCaller(VERBOSE, 1, err, message, a...)
}

func SecVerbose(message string, a ...interface{}) {
	LogCaller(SECVERBOSE, 1, nil, message, a...)
}

func Info(message string, a ...interface{}) {
	LogCaller(INFO, 1, nil, message, a...)
}

func InfoExc(err error, message string, a ...interface{}) {
	LogCaller(INFO, 1, err, message, a...)
}

func SecInfo(message string, a ...interface{}) {
	LogCaller(SECINFO, 1, nil, message, a...)
}

func Serv(message string, a ...interface{}) {
	LogCaller(SERV, 1, nil, message, a...)
}

func Recv(message string, a ...interface{}) {
	LogCaller(RECV, 1, nil, message, a...)
}

func Send(message string, a ...interface{}) {
	LogCaller(SEND, 1, nil, message, a...)
}

func Warn(message string, a ...interface{}) {
	LogCaller(WARN, 1, nil, message, a...)
}

func WarnExc(err error, message string, a ...interface{}) {
	LogCaller(WARN, 1, err, message, a...)
}

func SecWarn(message string, a ...interface{}) {
	LogCaller(SECWARN, 1, nil, message, a...)
}

func Error(message string, a ...interface{}) {
	LogCaller(ERROR, 1, nil, message, a...)
}

func ErrorExc(err error, message string, a ...interface{}) {
	LogCaller(ERROR, 1, err, message, a...)
}

func SecError(message string, a ...interface{}) {
	LogCaller(SECERROR, 1, nil, message, a...)
}

func Denied(message string, a ...interface{}) {
	LogCaller(DENIED, 1, nil, message, a...)
}

func Critical(message string, a ...interface{}) {
	LogCaller(CRITICAL, 1, nil, message, a...)
}

func CriticalExc(err error, message string, a ...interface{}) {
	LogCaller(CRITICAL, 1, err, message, a...)
}

func SecCritical(message string, a ...interface{}) {
	LogCaller(SECCRITICAL, 1, nil, message, a...)
}
