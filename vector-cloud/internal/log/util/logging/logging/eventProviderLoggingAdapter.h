/**
 * File: eventProviderLoggingAdapter.h
 *
 * Author: Brad Neuman
 * Created: 2018-08-20
 *
 * Description: Adapter to log DAS event info over another ILoggerProvider (e.g. for webots)
 *
 * Copyright: Anki, Inc. 2018
 *
 **/

#ifndef __Util_Logging_EventProviderLoggingAdapter_H__
#define __Util_Logging_EventProviderLoggingAdapter_H__

#include "util/logging/iLoggerProvider.h"
#include "util/logging/iEventProvider.h"

#include <mutex>

namespace Anki {
namespace Util {

class EventProviderLoggingAdapter : public IEventProvider
{
public:
  EventProviderLoggingAdapter( ILoggerProvider* logger );
  virtual ~EventProviderLoggingAdapter();

  // Implements IEventProvider
  virtual void SetGlobal(const char * key, const char * value) override;
  virtual void GetGlobals(std::map<std::string, std::string> & globals) override;
  virtual void LogEvent(LogLevel level, const DasMsg & dasMsg) override;

private:

  ILoggerProvider* _logger;

  std::mutex _mutex;
  std::map<std::string, std::string> _eventGlobals;


};


} // end namespace Util
} // end namespace Anki


#endif
