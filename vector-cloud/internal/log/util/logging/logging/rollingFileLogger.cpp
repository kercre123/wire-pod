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

#include "util/logging/rollingFileLogger.h"
#include "util/dispatchQueue/dispatchQueue.h"
#include "util/fileUtils/fileUtils.h"
#include "assert.h"

#include <sstream>
#include <iomanip>

#include <ctime>
#include <time.h> // Needed for the POSIX thread-safe version of local_time and put_time

// This alternative error logging define imported from our DAS library implementation
#define LOGD(...) {fprintf(stderr, __VA_ARGS__); fprintf(stderr, "\n"); fflush(stderr); }
#define LOGO(...) {fprintf(stdout, __VA_ARGS__); fprintf(stdout, "\n"); fflush(stdout); }

namespace Anki {
namespace Util {
  
const char * const RollingFileLogger::kDefaultFileExtension = ".log";

RollingFileLogger::RollingFileLogger(Dispatch::Queue* queue, const std::string& baseDirectory, const std::string& extension, std::size_t maxFileSize)
: _dispatchQueue(queue)
, _baseDirectory(baseDirectory)
, _extension(extension)
, _maxFileSize(maxFileSize)
{
  FileUtils::CreateDirectory(baseDirectory);
}

RollingFileLogger::RollingFileLogger(Dispatch::create_queue_t, const std::string& baseDirectory, const std::string& extension, std::size_t maxFileSize)
: RollingFileLogger(nullptr, baseDirectory, extension, maxFileSize)
{
  _ownedQueue.create("RFL");
  _dispatchQueue = _ownedQueue.get();
}
  
RollingFileLogger::~RollingFileLogger()
{
  _ownedQueue.reset();
  _currentLogFileHandle.close();
}

void RollingFileLogger::ExecuteBlock(const std::function<void ()>& block)
{
  if (nullptr != _dispatchQueue)
  {
    Dispatch::Async(_dispatchQueue, block);
  }
  else
  {
    block();
  }
}

void RollingFileLogger::Write(std::string message)
{
  ExecuteBlock([this, message = std::move(message)] {
    WriteInternal(message);
  });
}


// Note that WriteInternal and everything it calls will occur on another thread.
void RollingFileLogger::WriteInternal(const std::string& message)
{
  auto messageSize = message.length();
  // If file handle is closed or we've run out of space, open a new one
  if (!_currentLogFileHandle.is_open() || (messageSize + _numBytesWritten) > _maxFileSize)
  {
    RollLogFile();
    _numBytesWritten = 0;
  }
  
  assert(_currentLogFileHandle);
  
  if(!_currentLogFileHandle.is_open())
  {
    return;
  }
  
  _currentLogFileHandle << message;
  _currentLogFileHandle.flush();
  _numBytesWritten += messageSize;
}
  
void RollingFileLogger::RollLogFile()
{
  if (_currentLogFileHandle.is_open())
  {
    _currentLogFileHandle.close();
  }
  std::string nextFilename = GetNextFileName();
  _currentLogFileHandle.open(nextFilename, std::ofstream::out | std::ofstream::app);
  
  if (!_currentLogFileHandle)
  {
    LOGD("Error getting handle for file %s: %s !!", nextFilename.c_str(), strerror(errno));
  } else {
    LOGO("New log file created '%s'", nextFilename.c_str());
  }
}

std::string RollingFileLogger::GetNextFileName()
{
  std::ostringstream pathStream;
  if (!_baseDirectory.empty())
  {
    pathStream << _baseDirectory << '/';
  }
  
  pathStream << GetDateTimeString(ClockType::now()) << _extension;
  return pathStream.str();
}
  
std::string RollingFileLogger::GetDateTimeString(const ClockType::time_point& time)
{
  std::ostringstream stringStream;
  auto currTime_t = GetTimeT(time);
  auto numSecs = std::chrono::duration_cast<std::chrono::seconds>(time.time_since_epoch());
  auto millisLeft = std::chrono::duration_cast<std::chrono::milliseconds>((time - numSecs).time_since_epoch());
  
  struct tm localTime; // This is local scoped to make it thread safe
  localtime_r(&currTime_t, &localTime);
  
  // Use the old fashioned strftime for thread safety, instead of std::put_time
  char formatTimeBuffer[256];
  strftime(formatTimeBuffer, sizeof(formatTimeBuffer), "%FT%H-%M-%S-", &localTime);
  
  stringStream << formatTimeBuffer << std::setfill('0') << std::setw(3) << millisLeft.count();
  return stringStream.str();
}
  
time_t RollingFileLogger::GetTimeT(const ClockType::time_point& time)
{
  return std::chrono::system_clock::to_time_t(GetSystemClockTimePoint(time));
}
  
std::chrono::system_clock::time_point RollingFileLogger::GetSystemClockTimePoint(const ClockType::time_point& time)
{
  static const std::chrono::system_clock::time_point systemClockNow = std::chrono::system_clock::now();
  static const ClockType::time_point clockTypeNow = ClockType::now();
  
  const auto timeDiff = std::chrono::duration_cast<std::chrono::system_clock::duration>(time.time_since_epoch() - clockTypeNow.time_since_epoch());
  
  return systemClockNow + timeDiff;
}
  
void RollingFileLogger::FlushInternal()
{
  if (_currentLogFileHandle.is_open()) {
    _currentLogFileHandle.flush();
  }
}
  
void RollingFileLogger::Flush()
{
  if (nullptr != _dispatchQueue)
  {
    Dispatch::Sync(_dispatchQueue, [this] {
      FlushInternal();
    });
  }
  else
  {
    FlushInternal();
  }
}


} // end namespace Util
} // end namespace Anki
