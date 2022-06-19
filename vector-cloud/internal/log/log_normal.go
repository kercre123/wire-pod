// THIS IS DISABLED PLUSbuild !vicos

package log

import (
	"fmt"
	"os"
)

func Println(a ...interface{}) (int, error) {
	return fmt.Println(a...)
}

func Printf(format string, a ...interface{}) (int, error) {
	return fmt.Printf(format, a...)
}

func Errorln(a ...interface{}) (int, error) {
	return fmt.Fprintln(os.Stderr, a...)
}

func Errorf(format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(os.Stderr, format, a...)
}
