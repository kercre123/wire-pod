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
#ifndef __Util_Logging_IFormattedLoggerProvider_H_
#define __Util_Logging_IFormattedLoggerProvider_H_
#include "util/logging/iLoggerProvider.h"
#include "json/json-forwards.h"
#include <string>

namespace Anki {
  namespace Util {
    
    class IFormattedLoggerProvider : public ILoggerProvider {
      
    public:
      IFormattedLoggerProvider();
      
      inline void PrintLogE(const char* eventName,
                            const std::vector<std::pair<const char*, const char*>>& keyValues,
                            const char* eventValue) override {
        if ( !IsLogLevelEnabled(LOG_LEVEL_ERROR) ) { return; }
        FormatAndLog(LOG_LEVEL_ERROR, eventName, keyValues, eventValue);
      }
      inline void PrintLogW(const char* eventName,
                            const std::vector<std::pair<const char*, const char*>>& keyValues,
                            const char* eventValue) override {
        if (!IsLogLevelEnabled(LOG_LEVEL_WARN)) {return;}
        FormatAndLog(LOG_LEVEL_WARN, eventName, keyValues, eventValue);
      };
      inline void PrintEvent(const char* eventName,
                             const std::vector<std::pair<const char*, const char*>>& keyValues,
                             const char* eventValue) override {
        if (!IsLogLevelEnabled(LOG_LEVEL_EVENT)) {return;}
        FormatAndLog(LOG_LEVEL_EVENT, eventName, keyValues, eventValue);
      };
      inline void PrintLogI(const char* channelName,
                            const char* eventName,
                            const std::vector<std::pair<const char*, const char*>>& keyValues,
                            const char* eventValue) override {
        if (!IsLogLevelEnabled(LOG_LEVEL_INFO)) {return;}
        FormatAndLogChanneled(LOG_LEVEL_INFO, channelName, eventName, keyValues, eventValue);
      };
      inline void PrintLogD(const char* channelName,
                            const char* eventName,
                            const std::vector<std::pair<const char*, const char*>>& keyValues,
                            const char* eventValue) override {
        if (!IsLogLevelEnabled(LOG_LEVEL_DEBUG)) {return;}
        FormatAndLogChanneled(LOG_LEVEL_DEBUG, channelName, eventName, keyValues, eventValue);
      }
      
      // sets the minimum log level that is enabled. Levels above this one will also be enabled
      void SetMinLogLevel(LogLevel logLevel);
      // sets whether one specific log level is enabled
      void SetLogLevelEnabled(LogLevel logLevel, bool enabled);
      
      // reads which levels are enabled from json file
      void ParseLogLevelSettings(const Json::Value& config);
      
      // This has to be public for MultiFormattedLoggerProvider to work.
      virtual void Log(ILoggerProvider::LogLevel logLevel, const std::string& message) = 0;

    private:
    
      // returns true if the given log level is enabled, false otherwise
      bool IsLogLevelEnabled(LogLevel logLevel) const;
    
      void FormatAndLog(ILoggerProvider::LogLevel logLevel, const char* eventName,
                  const std::vector<std::pair<const char*, const char*>>& keyValues,
                  const char* eventValue);
      void FormatAndLogChanneled(ILoggerProvider::LogLevel logLevel, const char* channel,
                  const char* eventName,
                  const std::vector<std::pair<const char*, const char*>>& keyValues,
                  const char* eventValue);
      
      // whether specific log levels are enabled
      std::vector<bool> _logLevelEnabledFlags;
    };
    
  } // end namespace Util
} // end namespace Anki


#endif //__Util_Logging_IFormattedLoggerProvider_H_
