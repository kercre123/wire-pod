/**
 * File: iFormattedLoggerProvider
 *
 * Author: Trevor Dasch
 * Created: 2/10/16
 *
 * Description: abstract class
 * that formats the log into a c-string
 * and calls its Log function.
 *
 * Copyright: Anki, inc. 2016
 *
 */
#include "util/logging/iFormattedLoggerProvider.h"
#include "util/logging/logging.h"
#include "util/math/numericCast.h"
#include "json/json.h"
#include <sstream>
#include <string>
#include <iomanip>

#define PRINT_TID 1
#define PRINT_DAS_EXTRAS_BEFORE_EVENT 0
#define PRINT_DAS_EXTRAS_AFTER_EVENT 1
#define PRINT_DAS_EXTRAS_FOR_EVENT_LEVEL_ONLY 0

#if (PRINT_TID)
#include <stdarg.h>
#include <cstdint>
#include <ctime>
#include <thread>
#include <atomic>
#include <cassert>
#include <pthread.h>
#endif

namespace Anki {
namespace Util {
    
#if (PRINT_TID)
    static std::atomic<uint32_t> thread_max {0};
    static pthread_key_t thread_id_key;
    static pthread_once_t thread_id_once = PTHREAD_ONCE_INIT;
    
    static void thread_id_init()
    {
      pthread_key_create(&thread_id_key, nullptr);
    }
#endif

const char* kLevelListKey = "levels";
const char* kLevelNameKey = "level";
const char* kLevelEnabledKey = "enabled";

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
IFormattedLoggerProvider::IFormattedLoggerProvider()
{
  _logLevelEnabledFlags.resize(_LOG_LEVEL_COUNT, false);
  SetMinLogLevel(LOG_LEVEL_DEBUG);
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void IFormattedLoggerProvider::SetMinLogLevel(LogLevel logLevel)
{
  DEV_ASSERT(logLevel < _logLevelEnabledFlags.size(), "IFormattedLoggerProvider.SetMinLogLevel.InvalidLogLevel");
  for (int i = 0; i < logLevel; ++i) {
    _logLevelEnabledFlags[i] = false;
  }
  for (int i = logLevel; i < _LOG_LEVEL_COUNT; ++i) {
    _logLevelEnabledFlags[i] = true;
  }
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void IFormattedLoggerProvider::SetLogLevelEnabled(LogLevel logLevel, bool enabled)
{
  DEV_ASSERT(logLevel < _logLevelEnabledFlags.size(), "IFormattedLoggerProvider.SetLogLevelEnabled.InvalidLogLevel");
  _logLevelEnabledFlags[logLevel] = enabled;
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void IFormattedLoggerProvider::ParseLogLevelSettings(const Json::Value& config)
{
  // parse config
  if ( !config.isNull() ) {
    for( const auto& logLevelInfo : config[kLevelListKey] )
    {
      // parse channel name
      DEV_ASSERT(logLevelInfo[kLevelNameKey].isString(), "IFormattedLoggerProvider.ParseLogLevelSettings.BadName");
      const std::string& logLevelName = logLevelInfo[kLevelNameKey].asString();
      LogLevel logLevelVal = GetLogLevelValue(logLevelName);
      if (logLevelVal < _LOG_LEVEL_COUNT)
      {
        // parse value
        DEV_ASSERT(logLevelInfo[kLevelEnabledKey].isBool(), "IFormattedLoggerProvider.ParseLogLevelSettings.BadEnableFlag");
        const bool logLevelEnabled = logLevelInfo[kLevelEnabledKey].asBool();
      
        // Register channel
        SetLogLevelEnabled(logLevelVal, logLevelEnabled);
      }
    }
  }
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
bool IFormattedLoggerProvider::IsLogLevelEnabled(LogLevel logLevel) const
{
  DEV_ASSERT(logLevel < _logLevelEnabledFlags.size(), "IFormattedLoggerProvider.IsLogLevelEnabled.InvalidLogLevel");
  const bool ret = _logLevelEnabledFlags[logLevel];
  return ret;
}
  
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void IFormattedLoggerProvider::FormatAndLog(IFormattedLoggerProvider::LogLevel logLevel, const char* eventName,
                                  const std::vector<std::pair<const char*, const char*>>& keyValues,
                                  const char* eventValue)
{
  FormatAndLogChanneled(logLevel, "", eventName, keyValues, eventValue);
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void IFormattedLoggerProvider::FormatAndLogChanneled(ILoggerProvider::LogLevel logLevel, const char* channel,
            const char* eventName,
            const std::vector<std::pair<const char*, const char*>>& keyValues,
            const char* eventValue)
{
  #if (PRINT_TID)
    pthread_once(&thread_id_once, thread_id_init);
    
    uint32_t thread_id = numeric_cast_clamped<uint32_t>((uintptr_t)pthread_getspecific(thread_id_key));
    if(0 == thread_id) {
      thread_id = ++thread_max;
      pthread_setspecific(thread_id_key, (void*)((uintptr_t)thread_id));
    }
  #endif
      
  #if (PRINT_DAS_EXTRAS_BEFORE_EVENT || PRINT_DAS_EXTRAS_AFTER_EVENT)
    const bool printDasExtras = !PRINT_DAS_EXTRAS_FOR_EVENT_LEVEL_ONLY ||
                                logLevel == ILoggerProvider::LogLevel::LOG_LEVEL_EVENT;
    const size_t kMaxStringBufferSize = 1024;
    char logString[kMaxStringBufferSize]{0};
    if(printDasExtras) {
      for (const auto& keyValuePair : keyValues) {
        snprintf(logString, kMaxStringBufferSize, "%s[%s: %s] ", logString, keyValuePair.first, keyValuePair.second);
      }
    }
  #endif
  
  std::ostringstream stream;
      
  #if (PRINT_TID)
    stream << "(t:" << std::setw(2) << std::setfill('0') << thread_id << ") ";
  #endif
      
  stream << "[" << GetLogLevelString(logLevel) << "]";
      
  #if (PRINT_DAS_EXTRAS_BEFORE_EVENT)
    if(printDasExtras) {
      stream << " " << logString;
    }
  #endif
  
  ASSERT_NAMED(eventName!=nullptr, "IFormattedLoggerProvider.FormatAndLogChanneled logging null eventName");
  ASSERT_NAMED(eventValue!=nullptr, "IFormattedLoggerProvider.FormatAndLogChanneled logging null eventValue");
  static const char* emptyStr = "";
  if( eventName == nullptr ) {
    eventName = emptyStr;
  }
  if( eventValue == nullptr ) {
    eventValue = emptyStr;
  }

  std::string channelStr(channel);
  if (!channelStr.empty()) {
    stream << "[@" << channelStr << "] " << eventName << " " << eventValue;
  } else {
    stream << " " << eventName << " " << eventValue;
  }

  #if (PRINT_DAS_EXTRAS_AFTER_EVENT)
    if(printDasExtras) {
      stream << " " << logString;
    }
  #endif
      
  stream << std::endl;
      
  Log(logLevel, stream.str());
}
    
    
} // end namespace Util
} // end namespace Anki
