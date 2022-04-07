package wwlog

import (
	"fmt"
	"os"
)

const (
	CRITICAL    = 0
	SECCRITICAL = 1
	ERROR       = 2
	SECERROR    = 3
	WARN        = 4
	SECWARN     = 5
	INFO        = 6
	SECINFO     = 7
	VERBOSE     = 8
	SECVERBOSE  = 9
	DEBUG       = 10
	SECDEBUG    = 11
)

var (
	logLevel = INFO
)

/*
Set the central log level. Uneven values are security related
lof entries.
*/
func SetLevel(level int) {
	logLevel = level

	Printf(DEBUG, "Set log level to: %d\n", logLevel)
}

/*
generate the prefix for log level
*/
func prefixGen(level int) string {
	switch level {
	case DEBUG:
		return "[DEBUG]      : "
	case SECDEBUG:
		return "[SECDEBUG]   : "
	case VERBOSE:
		return "[VERBOSE]    : "
	case SECVERBOSE:
		return "[SECVERBOSE] : "
	case INFO:
		return "[INFO]       : "
	case SECINFO:
		return "[SECINFO]    : "
	case WARN:
		return "[WARNING]    : "
	case SECWARN:
		return "[SECWARNING] : "
	case ERROR:
		return "[ERROR]      : "
	case SECERROR:
		return "[SECERROR]   : "
	case CRITICAL:
		return "[CRITICAL]   : "
	case SECCRITICAL:
		return "[SECCRITICAL]: "
	}
	return "[UNDEF]   : "
}

func printlog(level int, message string) {
	if level == INFO && logLevel <= INFO {
		fmt.Print(message)
	} else if level <= logLevel {
		if level < INFO {
			fmt.Fprintf(os.Stderr, prefixGen(level)+message)
		} else {
			fmt.Printf(prefixGen(level) + message)
		}
	}
}

func Println(level int, message string) {
	printlog(level, message)
}

func Printf(level int, message string, a ...interface{}) {
	printlog(level, fmt.Sprintf(message, a...))
}
