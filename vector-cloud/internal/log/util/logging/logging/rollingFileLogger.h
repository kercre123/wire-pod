/**
* File: rollingFileLogger
*
* Author: Lee Crippen
* Created: 3/29/2016
*
* Description: 
*
* Copyright: Anki, inc. 2016
*
*/
#ifndef __Util_Logging_RollingFileLogger_H_
#define __Util_Logging_RollingFileLogger_H_

#include "util/dispatchQueue/dispatchQueue.h"
#include "util/helpers/noncopyable.h"

#include <fstream>
#include <cstdlib>
#include <string>

namespace Anki {
namespace Util {
  
class RollingFileLogger : noncopyable {
public:
  using ClockType = std::chrono::steady_clock;

  static constexpr std::size_t  kDefaultMaxFileSize = 1024 * 1024 * 20;
  static const char * const     kDefaultFileExtension;
  
  // use an existing queue for the file logger
  RollingFileLogger(Dispatch::Queue* queue, const std::string& baseDirectory, const std::string& extension = kDefaultFileExtension, std::size_t maxFileSize = kDefaultMaxFileSize);
  // create a new queue that will be owned by this file logger
  RollingFileLogger(Dispatch::create_queue_t, const std::string& baseDirectory, const std::string& extension = kDefaultFileExtension, std::size_t maxFileSize = kDefaultMaxFileSize);
  virtual ~RollingFileLogger();
  
  void Write(std::string message);
  void Flush();
  
  static std::string GetDateTimeString(const ClockType::time_point& time);
  static time_t GetTimeT(const ClockType::time_point& time);
  static std::chrono::system_clock::time_point GetSystemClockTimePoint(const ClockType::time_point& time);
  
private:
  void ExecuteBlock(const std::function<void()>& block);
  
  Dispatch::Queue*      _dispatchQueue;
  Dispatch::QueueHandle _ownedQueue;
  std::string       _baseDirectory;
  std::string       _extension;
  std::string       _currentFileName;
  std::size_t       _maxFileSize;
  std::size_t       _numBytesWritten = 0;
  std::ofstream     _currentLogFileHandle;
  
  void WriteInternal(const std::string& message);
  void FlushInternal();
  void RollLogFile();
  std::string GetNextFileName();
};

} // end namespace Util
} // end namespace Anki


#endif //__Util_Logging_RollingFileLogger_H_
