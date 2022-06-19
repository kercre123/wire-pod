package robot

/*

#include "crash_reporter.h"
#include <stdlib.h>

*/
import "C"
import "unsafe"

func InstallCrashReporter(procname string) {
	cproc := C.CString(procname)
	defer C.free(unsafe.Pointer(cproc))
	C.InstallCrashReporter(cproc)
}

func UninstallCrashReporter() {
	C.UninstallCrashReporter()
}
