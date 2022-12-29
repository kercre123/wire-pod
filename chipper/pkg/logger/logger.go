package logger

import (
	"fmt"
	"os"
)

var debugLogging bool = true

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
