package logging

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	LogFile *os.File
	Logger  *log.Logger
)

func Init(filename string) {
	tokens := strings.Split(filename, "/")
	filename = tokens[len(tokens)-1] + ".log"

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Failed to open log file")
	}

	LogFile = file
	Logger = log.New(LogFile, "", log.Ltime)
}
