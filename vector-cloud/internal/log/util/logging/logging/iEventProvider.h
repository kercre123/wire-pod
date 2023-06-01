/**
* File: iEventProvider
*
* Author: damjan
* Created: 4/21/2015
*
* Description: interface for anki BI Events
*
* Copyright: Anki, Inc. 2014
*
**/

#ifndef __Util_Logging_IEventProvider_H__
#define __Util_Logging_IEventProvider_H__

#include "logtypes.h"

#include <map>
#include <string>

// Forward declarations
namespace Anki {
  namespace Util {
    struct DasMsg;
  }
}

namespace Anki {
namespace Util {

class IEventProvider {
public:

  // Log an error event
  inline void LogError(const DasMsg & dasMsg) {
    LogEvent(LOG_LEVEL_ERROR, dasMsg);
  }

  // Log a warning event
  inline void LogWarning(const DasMsg & dasMsg) {
    LogEvent(LOG_LEVEL_WARN, dasMsg);
  }

  // Log an info event
  inline void LogInfo(const DasMsg & dasMsg) {
    LogEvent(LOG_LEVEL_INFO, dasMsg);
  }

  // Log a debug event
  inline void LogDebug(const DasMsg & dasMsg) {
    LogEvent(LOG_LEVEL_DEBUG, dasMsg);
  }

  // sets global properties.
  // all future logs+events will have the updated globals attached to them
  virtual void SetGlobal(const char* key, const char* value) = 0;

  // Get Globals
  virtual void GetGlobals(std::map<std::string, std::string>& dasGlobals) = 0;

  virtual void EnableNetwork(int reason) {}
  virtual void DisableNetwork(int reason) {}

protected:

  // Log an event at given level
  // To be implemented by each event provider
  virtual void LogEvent(LogLevel level, const DasMsg & dasMsg) = 0;

};

} // namespace Util
} // namespace Anki

#endif // __Util_Logging_IEventProvider_H__
