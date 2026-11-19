package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

var debugLogging bool = true

var (
	logMu       sync.Mutex
	logArray    []string
	logTrayList []string
)

var LogTrayChan chan string

func GetLogTrayChan() chan string {
	return LogTrayChan
}

func Init() {
	LogTrayChan = make(chan string)
	debugLogging = os.Getenv("DEBUG_LOGGING") == "true"
}

func GetLogs() string {
	logMu.Lock()
	defer logMu.Unlock()
	return strings.Join(logArray, "")
}

func GetDebugLogs() string {
	logMu.Lock()
	defer logMu.Unlock()
	return strings.Join(logTrayList, "")
}

func appendCapped(list []string, line string, max int) []string {
	list = append(list, line)
	if len(list) >= max {
		list = list[1:]
	}
	return list
}

func Println(a ...any) {
	LogTray(a...)
	if debugLogging {
		fmt.Println(a...)
	}
}

func LogUI(a ...any) {
	line := time.Now().Format("2006.01.02 15:04:05") + ": " + fmt.Sprint(a...) + "\n"
	logMu.Lock()
	logArray = appendCapped(logArray, line, 50)
	logMu.Unlock()
}

func LogTray(a ...any) {
	line := time.Now().Format("2006.01.02 15:04:05") + ": " + fmt.Sprint(a...) + "\n"
	logMu.Lock()
	logTrayList = appendCapped(logTrayList, line, 200)
	logMu.Unlock()
	select {
	case LogTrayChan <- line:
	default:
	}
}
