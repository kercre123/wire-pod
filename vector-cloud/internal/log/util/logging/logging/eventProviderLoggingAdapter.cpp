/**
 * File: eventProviderLoggingAdapter.cpp
 *
 * Author: Brad Neuman
 * Created: 2018-08-20
 *
 * Description: Adapter to log DAS event info over another ILoggerProvider (e.g. for webots)
 *
 * Copyright: Anki, Inc. 2018
 *
 **/

#include "util/logging/eventProviderLoggingAdapter.h"
#include "util/logging/DAS.h"

#include <assert.h>
#include <sstream>
#include <string>


namespace Anki {
namespace Util {

EventProviderLoggingAdapter::EventProviderLoggingAdapter( ILoggerProvider* logger )
  : _logger(logger)
{
}

EventProviderLoggingAdapter::~EventProviderLoggingAdapter()
{
}

void EventProviderLoggingAdapter::SetGlobal(const char * key, const char * value)
{
  // Key may not be null
  assert(key != nullptr);

  std::lock_guard<std::mutex> lock(_mutex);

  if (value == nullptr) {
    if( _logger ) {
      _logger->PrintEvent("DAS_GLOBALS.Clear", {}, key);
    }
    
    _eventGlobals.erase(key);
  } else {
    if( _logger ) {
      std::stringstream ss;
      ss << key << "=" << value;      
      _logger->PrintEvent("DAS_GLOBALS.Set", {}, ss.str().c_str());
    }
    _eventGlobals.emplace(std::pair<std::string,std::string>{key, value});
  }
}

void EventProviderLoggingAdapter::GetGlobals(std::map<std::string, std::string> & globals)
{
  std::lock_guard<std::mutex> lock(_mutex);
  globals = _eventGlobals;
}


#define DAS_KV_HELPER(kv, msg, item) {                  \
  if( !msg.item.value.empty() ) {                       \
    kv.emplace_back( #item, msg.item.value.c_str() );   \
  }                                                     \
}

void EventProviderLoggingAdapter::LogEvent(LogLevel level, const DasMsg & dasMsg)
{
  if( _logger == nullptr ) {
    return;
  }

  // TODO:(bn) option for whether or not to include globals?

  std::vector<std::pair<const char*, const char*>> keyValues;
  DAS_KV_HELPER(keyValues, dasMsg, s1);
  DAS_KV_HELPER(keyValues, dasMsg, s2);
  DAS_KV_HELPER(keyValues, dasMsg, s3);
  DAS_KV_HELPER(keyValues, dasMsg, s4);
  DAS_KV_HELPER(keyValues, dasMsg, i1);
  DAS_KV_HELPER(keyValues, dasMsg, i2);
  DAS_KV_HELPER(keyValues, dasMsg, i3);
  DAS_KV_HELPER(keyValues, dasMsg, i4);

  const std::string & uptime_ms = std::to_string(Anki::Util::DAS::UptimeMS());
  keyValues.emplace_back("uptime_ms", uptime_ms.c_str());

  switch(level) {
    case LOG_LEVEL_DEBUG:
    case LOG_LEVEL_INFO:
    case LOG_LEVEL_EVENT:
      _logger->PrintEvent("DASMSG", keyValues, dasMsg.event.c_str());
      break;

    case LOG_LEVEL_WARN:
      _logger->PrintLogW("DASMSG", keyValues, dasMsg.event.c_str());
      break;

    case LOG_LEVEL_ERROR:
      _logger->PrintLogE("DASMSG", keyValues, dasMsg.event.c_str());
      break;
      
    case _LOG_LEVEL_COUNT:
      assert(level != _LOG_LEVEL_COUNT);
      break;
  }
}


} // end namespace Util
} // end namespace Anki
