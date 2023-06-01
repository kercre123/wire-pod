/**
* File: androidLogPrintLogger_vicos.cpp
*
* Description: Implements ILoggerProvider for __android_log_print()
*
* Copyright: Anki, inc. 2017
*
*/

#include "util/logging/androidLogPrintLogger_vicos.h"
#include "util/logging/iLoggerProvider.h"
#include <android/log.h>
#include <assert.h>

using LogLevel = Anki::Util::ILoggerProvider::LogLevel;

namespace Anki {
namespace Util {

AndroidLogPrintLogger::AndroidLogPrintLogger(const std::string& tag) :
  _tag(tag)
{

}

static inline android_LogPriority GetPriority(LogLevel level)
{
  switch (level) {
    case LogLevel::LOG_LEVEL_DEBUG:
      return ANDROID_LOG_VERBOSE;
    case LogLevel::LOG_LEVEL_INFO:
      return ANDROID_LOG_DEBUG;
    case LogLevel::LOG_LEVEL_EVENT:
      return ANDROID_LOG_INFO;
    case LogLevel::LOG_LEVEL_WARN:
      return  ANDROID_LOG_WARN;
    case LogLevel::LOG_LEVEL_ERROR:
      /* intentional fallthrough */
    case LogLevel::_LOG_LEVEL_COUNT:
      return ANDROID_LOG_ERROR;
  }
  return ANDROID_LOG_DEFAULT;
}

void AndroidLogPrintLogger::Log(LogLevel level,
  const char * eventName,
  const std::vector<std::pair<const char *, const char *>>& keyValues,
  const char * eventValue)
{
  assert(eventName != nullptr);
  assert(eventValue != nullptr);
  __android_log_print(GetPriority(level), _tag.c_str(), "%s: %s", eventName, eventValue);
}

void AndroidLogPrintLogger::Log(LogLevel level,
  const char * channel,
  const char * eventName,
  const std::vector<std::pair<const char *, const char *>>& keyValues,
  const char * eventValue)
{
  assert(channel != nullptr);
  assert(eventName != nullptr);
  assert(eventValue != nullptr);
  __android_log_print(GetPriority(level), _tag.c_str(), "[@%s] %s: %s", channel, eventName, eventValue);
}

} // end namespace Util
} // end namespace Anki
