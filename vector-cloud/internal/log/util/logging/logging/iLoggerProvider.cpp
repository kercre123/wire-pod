/**
* File: iLoggerProvider.cpp
*
* Author: raul
* Created: 06/30/16
*
* Description: interface for anki log
*
* Copyright: Anki, Inc. 2014
*
**/
#include "iLoggerProvider.h"

#include "logging.h"

namespace Anki {
namespace Util {

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void ILoggerProvider::SetFilter(const std::shared_ptr<const IChannelFilter>& infoFilter)
{
  _infoFilter = infoFilter;
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void ILoggerProvider::PrintChanneledLogI(const char* channel,
      const char* eventName,
      const std::vector<std::pair<const char*, const char*>>& keyValues,
      const char* eventValue)
{
  // if no filter is set or the channel is enabled
  if ( !_infoFilter || _infoFilter->IsChannelEnabled(channel) )
  {
    // pass to subclass
    PrintLogI(channel, eventName, keyValues, eventValue);
  }
}

void ILoggerProvider::PrintChanneledLogD(const char* channel,
      const char* eventName,
      const std::vector<std::pair<const char*, const char*>>& keyValues,
      const char* eventValue)
{
  // if no filter is set or the channel is enabled
  if ( !_infoFilter || _infoFilter->IsChannelEnabled(channel) )
  {
    // pass to subclass
    PrintLogD(channel, eventName, keyValues, eventValue);
  }
}

} // namespace Util
} // namespace Anki
