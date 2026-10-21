//go:build go1.23

package logger

import (
	"os"
	"runtime/debug"
)

func setCrashOutput(f *os.File) {
	debug.SetCrashOutput(f, debug.CrashOptions{})
}
