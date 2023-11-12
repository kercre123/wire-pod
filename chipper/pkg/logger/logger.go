package logger

import (
	"fmt"
	"os"
	"time"
)

var debugLogging bool = true
var LogList string
var LogArray []string

var TrayLogList string
var TrayLogArray []string
var LogChan chan string

func GetChan() chan string {
	return LogChan
}

func Init() {
	LogChan = make(chan string)
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
	LogTray(a...)
	select {
	case LogChan <- fmt.Sprint(a...):
	default:
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
	TrayLogArray = append(TrayLogArray, time.Now().Format("2006.01.02 15:04:05")+": "+fmt.Sprint(a...)+"\n")
	if len(TrayLogArray) >= 50 {
		TrayLogArray = TrayLogArray[1:]
	}
	TrayLogList = ""
	for _, b := range TrayLogArray {
		TrayLogList = TrayLogList + b
	}
}
