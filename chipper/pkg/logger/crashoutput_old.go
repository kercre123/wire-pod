//go:build !go1.23

package logger

import "os"

// debug.SetCrashOutput needs Go 1.23; older builds only log to the file.
func setCrashOutput(f *os.File) {}
