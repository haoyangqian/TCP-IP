package logging

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	LogFile       *os.File
	Logger        *log.Logger
	EnableLogging bool
)

func Init(filename string, enableLogging bool) {
	tokens := strings.Split(filename, "/")
	filename = tokens[len(tokens)-1] + ".log"

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Failed to open log file")
	}

	LogFile = file
	Logger = log.New(LogFile, "", log.Ltime)
	EnableLogging = enableLogging
}

func Println(a ...interface{}) {
	if EnableLogging {
		Logger.Println(a...)
	}
}

func Printf(formatString string, a ...interface{}) {
	if EnableLogging {
		Logger.Printf(formatString, a...)

	}
}
