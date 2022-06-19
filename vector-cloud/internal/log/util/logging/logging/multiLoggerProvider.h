/**
 * File: MultiLoggerProvider
 *
 * Author: Trevor Dasch
 * Created: 2/9/16
 *
 * Description:
 * Multiplexes multiple loggerProviders
 *
 * Copyright: Anki, inc. 2016
 *
 */
#ifndef __Util_Logging_MultiLoggerProvider_H_
#define __Util_Logging_MultiLoggerProvider_H_
#include "util/logging/iLoggerProvider.h"

namespace Anki {
  namespace Util {
    
    class MultiLoggerProvider : public ILoggerProvider {
      
    public:
      inline MultiLoggerProvider(const std::vector<ILoggerProvider*>& inVec) {
        for(ILoggerProvider* provider : inVec ) {
          _providers.emplace_back(provider);
        }
      }
      
      inline ~MultiLoggerProvider() {
        for(ILoggerProvider* provider : _providers) {
          delete provider;
        }
        _providers.clear();
      }
      
      MultiLoggerProvider(const MultiLoggerProvider&) = delete;
      MultiLoggerProvider& operator=(const MultiLoggerProvider&) = delete;
      
      inline void PrintEvent(const char* eventName,
                             const std::vector<std::pair<const char*, const char*>>& keyValues,
                             const char* eventValue) override {
        for(ILoggerProvider* provider : _providers) {
          provider->PrintEvent(eventName, keyValues, eventValue);
        }
      };
      inline void PrintLogE(const char* eventName,
                            const std::vector<std::pair<const char*, const char*>>& keyValues,
                            const char* eventValue) override {
        for(ILoggerProvider* provider : _providers) {
          provider->PrintLogE(eventName, keyValues, eventValue);
        }
      }
      inline void PrintLogW(const char* eventName,
                            const std::vector<std::pair<const char*, const char*>>& keyValues,
                            const char* eventValue) override {
        for(ILoggerProvider* provider : _providers) {
          provider->PrintLogW(eventName, keyValues, eventValue);
        }
      };
      inline void PrintLogI(const char* channel,
                            const char* eventName,
                            const std::vector<std::pair<const char*, const char*>>& keyValues,
                            const char* eventValue) override {
        for(ILoggerProvider* provider : _providers) {
          provider->PrintChanneledLogI(channel, eventName, keyValues, eventValue);
        }
      };
      inline void PrintLogD(const char* channel,
                            const char* eventName,
                            const std::vector<std::pair<const char*, const char*>>& keyValues,
                            const char* eventValue) override {
        for(ILoggerProvider* provider : _providers) {
          provider->PrintChanneledLogD(channel, eventName, keyValues, eventValue);
        }
      }
      
      inline void Flush() override {
        for (ILoggerProvider* provider : _providers) {
          provider->Flush();
        }
      }
      
      inline ILoggerProvider* GetProvider(int index) {
        return _providers[index];
      }
      
      inline const std::vector<ILoggerProvider*>& GetProviders() {
        return _providers;
      }
      
    protected:
      std::vector<ILoggerProvider*> _providers;
      
      
    };
    
  } // end namespace Util
} // end namespace Anki


#endif //__Util_Logging_MultiLoggerProvider_H_
