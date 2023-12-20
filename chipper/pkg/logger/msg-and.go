//go:build android
// +build android

package logger

import (
	"fmt"
)

func WarnMsg(msg string) {
	fmt.Println(msg)
}

func ErrMsg(msg string) {
	fmt.Println(msg)
}
