package wwlog

import (
	"log"
)

const (
	CRITICAL 	= 0
	ERROR 		= 1
	WARN	 	= 2
	INFO		= 3
	VERBOSE		= 4
	DEBUG 		= 5
)

var (
	logLevel uint
)

func SetLevel(level uint) {
	logLevel = level

	if logLevel == DEBUG {
		log.SetFlags(log.Lmicroseconds | log.Llongfile)
	} else {
		log.SetFlags(0)
	}
	Printf(VERBOSE, "Set log level to: %d\n", logLevel)
}

func prefixLevel(level uint) {
	if level == DEBUG {
		log.SetPrefix("[DEBUG]    ")
	} else if level == VERBOSE {
		log.SetPrefix("[VERBOSE]  ")
	} else if level == INFO {
		log.SetPrefix("[INFO]   ")
	} else if level == WARN {
		log.SetPrefix("[WARN]     ")
	} else if level == ERROR {
		log.SetPrefix("[ERROR] 	  ")
	} else if level == CRITICAL {
		log.SetPrefix("[CRITICAL] ")
	}
}

func Println(level uint, message string) {
	if level <= logLevel {
		prefixLevel(level)
		log.Println(message)
	}
}

func Printf(level uint, message string, a...interface{}) {
	if level <= logLevel {
		prefixLevel(level)
		log.Printf(message, a...)
	}
}
