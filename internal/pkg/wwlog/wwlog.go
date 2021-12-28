package wwlog

import (
	"fmt"
	"os"
)

const (
	CRITICAL = 0
	ERROR    = 1
	WARN     = 2
	INFO     = 3
	VERBOSE  = 4
	DEBUG    = 5
)

var (
	logLevel = INFO
)

func SetLevel(level int) {
	logLevel = level

	Printf(DEBUG, "Set log level to: %d\n", logLevel)
}

func prefixGen(level int) string {
	switch level {
	case DEBUG:
		return "[DEBUG]   : "
	case VERBOSE:
		return "[VERBOSE] : "
	case INFO:
		return "[INFO]    : "
	case WARN:
		return "[WARNING] : "
	case ERROR:
		return "[ERROR]   : "
	case CRITICAL:
		return "[CRITICAL]: "
	}
	return "[UNDEF]   : "
}

func printlog(level int, message string) {
	if level == INFO && logLevel <= INFO {
		fmt.Printf(message)
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
