package logger

import (
	"fmt"
	"os"
	"time"
)

var debugLogging bool = true
var LogList string
var LogArray []string

func Init() {
	if os.Getenv("DEBUG_LOGGING") == "true" {
		debugLogging = true
	} else {
		debugLogging = false
	}
}

func Println(a ...any) {
	if debugLogging {
		fmt.Println(a...)
	}
}

func LogMatch(a ...any) {
	LogArray = append(LogArray, time.Now().Format("2006.01.02 15:04:05")+": "+fmt.Sprint(a...)+"\n")
	if len(LogArray) >= 30 {
		LogArray = LogArray[1:]
	}
	LogList = ""
	for _, b := range LogArray {
		LogList = LogList + b
	}
}
