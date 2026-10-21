package logger

// Optional log file, so logs survive a crash or restart. Packaged builds
// (macOS, Windows, Android apps) have no terminal, so without this the only
// copy of the log is in memory and is lost when wire-pod exits.

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// the file is rotated to <path>.1 when it grows past this
const maxLogFileSize = 5 * 1024 * 1024

var (
	logFileMu   sync.Mutex
	logFile     *os.File
	logFilePath string
	logFileSize int64
	lastLogLine string
)

// SetLogFile starts appending every log line to path, rotating it at 5 MB.
// It also sends the output of an unrecovered panic to the same file where
// the Go version supports it.
func SetLogFile(path string) error {
	logFileMu.Lock()
	defer logFileMu.Unlock()
	if logFile != nil {
		logFile.Close()
	}
	logFilePath = path
	if err := openLogFile(); err != nil {
		return err
	}
	setCrashOutput(logFile)
	return nil
}

// openLogFile opens logFilePath for appending, rotating it first if it is
// already too big. Call with logFileMu held.
func openLogFile() error {
	if info, err := os.Stat(logFilePath); err == nil && info.Size() >= maxLogFileSize {
		os.Rename(logFilePath, logFilePath+".1")
	}
	f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logFile = nil
		return err
	}
	logFile = f
	logFileSize = 0
	if info, err := f.Stat(); err == nil {
		logFileSize = info.Size()
	}
	return nil
}

func writeLogFile(a ...any) {
	logFileMu.Lock()
	defer logFileMu.Unlock()
	if logFile == nil {
		return
	}
	// some messages go to both Println and LogUI; only write them once
	line := time.Now().Format("2006.01.02 15:04:05") + ": " + fmt.Sprint(a...)
	if line == lastLogLine {
		return
	}
	lastLogLine = line
	n, _ := fmt.Fprintln(logFile, line)
	logFileSize += int64(n)
	if logFileSize >= maxLogFileSize {
		logFile.Close()
		os.Rename(logFilePath, logFilePath+".1")
		if openLogFile() == nil {
			setCrashOutput(logFile)
		}
	}
}
