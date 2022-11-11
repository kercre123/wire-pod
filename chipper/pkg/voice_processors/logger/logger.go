package logger

import "fmt"

var DebugLogging bool

func Log(a ...any) {
	if DebugLogging {
		fmt.Println(a...)
	}
}
