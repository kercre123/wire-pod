// +build vicosBROKEN

package log

import (
	"fmt"
	"unsafe"
)

/*
#cgo LDFLAGS: -llog
#include <android/log.h>
#include <stdlib.h>

// Go cannot call variadic C functions directly
static int android_log(int prio, const char* tag, const char* str) {
	return __android_log_print(prio, tag, "%s", str);
}
*/
import "C"

func androidLog(level C.int, tag string, str string) int {
	ctag := C.CString(tag)
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(ctag))
	defer C.free(unsafe.Pointer(cstr))
	ret := C.android_log(level, ctag, cstr)
	return int(ret)
}

func Println(a ...interface{}) (int, error) {
	str := fmt.Sprintln(a...)
	return androidLog(C.ANDROID_LOG_INFO, Tag, str), nil
}

func Printf(format string, a ...interface{}) (int, error) {
	str := fmt.Sprintf(format, a...)
	return androidLog(C.ANDROID_LOG_INFO, Tag, str), nil
}

func Errorln(a ...interface{}) (int, error) {
	str := fmt.Sprintln(a...)
	return androidLog(C.ANDROID_LOG_ERROR, Tag, str), nil
}

func Errorf(format string, a ...interface{}) (int, error) {
	str := fmt.Sprintf(format, a...)
	return androidLog(C.ANDROID_LOG_ERROR, Tag, str), nil
}

func init() {
	logVicos = Println
}
