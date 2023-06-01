/**
 * File: MultiFormattedLoggerProvider
 *
 * Author: Trevor Dasch
 * Created: 2/9/16
 *
 * Description:
 * MultiFormattedplexes multiple formatted 
 * logger providers so they can share one allocated
 * formatted string.
 *
 * Copyright: Anki, inc. 2016
 *
 */
#ifndef __Util_Logging_MultiFormattedLoggerProvider_H_
#define __Util_Logging_MultiFormattedLoggerProvider_H_
#include "util/logging/iFormattedLoggerProvider.h"
#include "util/logging/multiLoggerProvider.h"

namespace Anki {
  namespace Util {
    
    class MultiFormattedLoggerProvider : public MultiLoggerProvider {
      
    public:
      inline MultiFormattedLoggerProvider(const std::vector<IFormattedLoggerProvider*>& inVec) : MultiLoggerProvider({}) {
        for(IFormattedLoggerProvider* provider : inVec ) {
          _providers.emplace_back(provider);
        }
      }
      
      MultiFormattedLoggerProvider(const MultiFormattedLoggerProvider&) = delete;
      MultiFormattedLoggerProvider& operator=(const MultiFormattedLoggerProvider&) = delete;
      
      inline IFormattedLoggerProvider* GetProvider(int index) {
        return dynamic_cast<IFormattedLoggerProvider*>(_providers[index]);
      }
      
      inline void Log(ILoggerProvider::LogLevel logLevel, const std::string& logMessage) {
        for(ILoggerProvider* provider : _providers ) {
          IFormattedLoggerProvider* formatted_provider = dynamic_cast<IFormattedLoggerProvider*>(provider);
          formatted_provider->Log(logLevel, logMessage);
        }
      }
      
      inline void SetMinLogLevel(Anki::Util::ILoggerProvider::LogLevel log_level) {
        for(ILoggerProvider* provider : _providers ) {
          IFormattedLoggerProvider* formatted_provider = dynamic_cast<IFormattedLoggerProvider*>(provider);
          formatted_provider->SetMinLogLevel(log_level);
        }
      }
      
    };
    
  } // end namespace Util
} // end namespace Anki


#endif //__Util_Logging_MultiFormattedLoggerProvider_H_
