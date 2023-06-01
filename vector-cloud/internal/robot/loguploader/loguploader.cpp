//
// GO anki/robot/loguploader/loguploader.cpp
//
// Wrapper methods so C++ library functions can be called from go.
//

#include "loguploader.h"

#include "platform/robotLogUploader/robotLogUploader.h"

extern "C"
{

//
// Returns 0 and outstr=url on success
// Returns non-zero and outstr=error on failure
// Caller is responsible for releasing outstr
//
int UploadDebugLogs(char ** outstr)
{
  std::string status;
  const int result = Anki::Vector::RobotLogUploader::UploadDebugLogs(status);
  if (outstr != nullptr) {
    *outstr = strdup(status.c_str());
  }
  return result;
}

} // end extern "C"
