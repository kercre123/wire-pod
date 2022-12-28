package logger

import (
	"fmt"
	"os"
)

func Println(a ...any) {
	if os.Getenv("DEBUG_LOGGING") == "true" {
		fmt.Println(a...)
	}
}
