/**
* File: androidLogPrintLogger_vicos.h
*
* Description: Implements ILoggerProvider for __android_log_print()
*
* Copyright: Anki, inc. 2017
*
*/
#ifndef __Util_Logging_AndroidLogPrintLogger_H_
#define __Util_Logging_AndroidLogPrintLogger_H_

#include "util/logging/iLoggerProvider.h"

#include <string>

namespace Anki {
namespace Util {

class AndroidLogPrintLogger : public ILoggerProvider {
public:

  AndroidLogPrintLogger(const std::string& tag = "anki");
  
  // Implements ILoggerProvider
  virtual void PrintEvent(const char * eventName,
    const std::vector<std::pair<const char * , const char *>>& keyValues,
    const char * eventValue)
  {
      Log(LogLevel::LOG_LEVEL_EVENT, eventName, keyValues, eventValue);
  }

  virtual void PrintLogE(const char * eventName,
    const std::vector<std::pair<const char *, const char *>>& keyValues,
    const char * eventValue)
  {
      Log(LogLevel::LOG_LEVEL_ERROR, eventName, keyValues, eventValue);
  }

  virtual void PrintLogW(const char* eventName,
    const std::vector<std::pair<const char *, const char *>>& keyValues,
    const char * eventValue)
  {
      Log(LogLevel::LOG_LEVEL_WARN, eventName, keyValues, eventValue);
  }

  virtual void PrintLogI(const char * channel,
    const char * eventName,
    const std::vector<std::pair<const char *, const char *>>& keyValues,
    const char * eventValue)
  {
      Log(LogLevel::LOG_LEVEL_INFO, channel, eventName, keyValues, eventValue);
  }

  virtual void PrintLogD(const char * channel,
    const char * eventName,
    const std::vector<std::pair<const char *, const char *>>& keyValues,
    const char * eventValue)
  {
      Log(LogLevel::LOG_LEVEL_DEBUG, channel, eventName, keyValues, eventValue);
  }
  
  // Implements IFormattedLoggerProvider
  // virtual void Log(ILoggerProvider::LogLevel logLevel, const std::string& message) override;

private:
  std::string _tag;

  void Log(LogLevel level,
    const char * eventName, 
    const std::vector<std::pair<const char *, const char *>>& keyValues,
    const char * eventValue);

  void Log(LogLevel level, 
    const char * channel, 
    const char * eventName, 
    const std::vector<std::pair<const char *, const char *>>& keyValues,
    const char * eventValue);

}; // class AndroidLogPrintLogger

} // end namespace Util
} // end namespace Anki

#endif //__Util_Logging_AndroidLogPrintLogger_H_
