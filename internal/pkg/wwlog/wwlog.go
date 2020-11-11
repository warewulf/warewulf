package wwlog

import (
	"log"
	"strings"
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
	Indent string
)

func SetLevel(level uint) {
	logLevel = level

	if level == DEBUG {
		log.SetFlags(log.Lmicroseconds)
	} else {
		log.SetFlags(0)
	}

	Printf(DEBUG, "Set log level to: %d\n", logLevel)
}

func SetIndent(i int) {
	Indent = strings.Repeat(" ", i)
}

func prefixLevel(level uint) {
	if level == DEBUG {
		log.SetPrefix("[DEBUG]    "+Indent)
	} else if level == VERBOSE {
		log.SetPrefix("[VERBOSE]  "+Indent)
	} else if level == INFO {
		log.SetPrefix("[INFO]     "+Indent)
	} else if level == WARN {
		log.SetPrefix("[WARN]     "+Indent)
	} else if level == ERROR {
		log.SetPrefix("[ERROR] 	  "+Indent)
	} else if level == CRITICAL {
		log.SetPrefix("[CRITICAL] "+Indent)
	}
}

func Println(level uint, message string) {
	if level <= logLevel {
		prefixLevel(level)
		log.Println(message)
	}

	log.SetPrefix("[LOG]      "+Indent)
}

func Printf(level uint, message string, a...interface{}) {
	if level <= logLevel {
		prefixLevel(level)
		log.Printf(message, a...)
	}

	log.SetPrefix("[LOG]      "+Indent)
}
