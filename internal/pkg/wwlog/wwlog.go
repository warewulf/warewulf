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
	if level == DEBUG {
		return "[DEBUG]"
	} else if level == VERBOSE {
		return "[VERBOSE]"
	} else if level == INFO {
		return "[INFO]"
	} else if level == WARN {
		return "[WARNING]"
	} else if level == ERROR {
		return "[ERROR]"
	} else if level == CRITICAL {
		return "[CRITICAL]"
	}
	return "[UNDEF]"
}

func Println(level int, message string) {
	if level <= logLevel {
		fmt.Fprintln(os.Stderr, prefixGen(level)+" "+message)
	}
}

func Printf(level int, message string, a ...interface{}) {
	if level <= logLevel {
		fmt.Fprintf(os.Stderr, prefixGen(level)+" "+message, a...)
	}
}
