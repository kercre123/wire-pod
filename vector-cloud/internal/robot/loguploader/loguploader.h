//
// anki/robot/loguploader/loguploader.h
//
// Declare an extern "C" interface to C++ internals.
// This file provides a bridge between loguploader.go
// (compiled as C) and loguploader.cpp (compiled as C++).
//

#ifndef __anki_robot_loguploader_loguploader_h
#define __anki_robot_loguploader_loguploader_h

#ifdef __cplusplus
extern "C" {
#endif

//
// Returns 0 and outstr=url on success
// Returns non-zero and outstr=error on failure
// Caller is responsible for releasing outstr
//
int UploadDebugLogs(char** outstr);

#ifdef __cplusplus
}
#endif

#endif // __anki_robot_loguploader_loguploader_h
