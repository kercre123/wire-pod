package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogFileRotates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wire-pod.log")
	if err := os.WriteFile(path, make([]byte, maxLogFileSize), 0644); err != nil {
		t.Fatal(err)
	}
	if err := SetLogFile(path); err != nil {
		t.Fatal(err)
	}
	defer func() {
		logFileMu.Lock()
		logFile.Close()
		logFile = nil
		logFileMu.Unlock()
	}()
	if info, err := os.Stat(path + ".1"); err != nil || info.Size() != maxLogFileSize {
		t.Fatalf("oversized log was not rotated on open: %v", err)
	}
	LogUI("hello from the UI")
	Println("hello from Println")
	Println("said twice")
	LogUI("said twice")
	b, _ := os.ReadFile(path)
	if strings.Count(string(b), "said twice") != 1 {
		t.Fatalf("a line logged to both Println and LogUI was written twice: %q", b)
	}
	if !strings.Contains(string(b), "hello from the UI") || !strings.Contains(string(b), "hello from Println") {
		t.Fatalf("log file missing lines: %q", b)
	}

	// rotates again once it fills up while running
	logFileMu.Lock()
	logFileSize = maxLogFileSize - 1
	logFileMu.Unlock()
	Println("this line triggers rotation")
	Println("first line of the new file")
	b, _ = os.ReadFile(path)
	if strings.Contains(string(b), "hello from Println") || !strings.Contains(string(b), "first line of the new file") {
		t.Fatalf("log was not rotated while running: %q", b)
	}
}
