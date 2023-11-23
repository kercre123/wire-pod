package logger

import (
	"fmt"
	"os"
	"time"
)

var debugLogging bool = true
var LogList string
var LogArray []string

var LogTrayList string
var LogTrayArray []string
var LogTrayChan chan string

func GetLogTrayChan() chan string {
	return LogTrayChan
}

func Init() {
	LogTrayChan = make(chan string)
	if os.Getenv("DEBUG_LOGGING") == "true" {
		debugLogging = true
	} else {
		debugLogging = false
	}
}

func Println(a ...any) {
	LogTray(a...)
	if debugLogging {
		fmt.Println(a...)
	}
}

func LogUI(a ...any) {
	LogArray = append(LogArray, time.Now().Format("2006.01.02 15:04:05")+": "+fmt.Sprint(a...)+"\n")
	if len(LogArray) >= 30 {
		LogArray = LogArray[1:]
	}
	LogList = ""
	for _, b := range LogArray {
		LogList = LogList + b
	}
}

func LogTray(a ...any) {
	LogTrayArray = append(LogTrayArray, time.Now().Format("2006.01.02 15:04:05")+": "+fmt.Sprint(a...)+"\n")
	if len(LogTrayArray) >= 30 {
		LogTrayArray = LogTrayArray[1:]
	}
	LogTrayList = ""
	for _, b := range LogTrayArray {
		LogTrayList = LogTrayList + b
	}
	select {
	case LogTrayChan <- time.Now().Format("2006.01.02 15:04:05") + ": " + fmt.Sprint(a...) + "\n":
	default:
	}
}
