package loguploader

/*

#include "loguploader.h"
#include <stdlib.h>

*/
import "C"
import (
	"errors"
	"unsafe"
)

func UploadDebugLogs() (string, error) {
	var outstr *C.char
	result := C.UploadDebugLogs(&outstr)
	str := C.GoString(outstr)
	C.free(unsafe.Pointer(outstr))
	if result != 0 {
		return "", errors.New(str)
	}
	return str, nil
}
