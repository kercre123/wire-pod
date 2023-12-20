//go:build !android
// +build !android

package logger

import (
	"github.com/ncruces/zenity"
)

func WarnMsg(msg string) {
	zenity.Warning(
		msg,
		zenity.WarningIcon,
		zenity.Title("WirePod"),
	)
}

func ErrMsg(msg string) {
	zenity.Error(
		msg,
		zenity.ErrorIcon,
		zenity.Title("WirePod"),
	)
}
