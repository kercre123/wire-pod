/**
 * File: util/logging/logtypes.cpp
 *
 * Description: Implementation of Anki::Util log types
 *
 * Copyright: Anki, Inc. 2018
 *
 **/

#include "logtypes.h"
#include "logging.h"

namespace Anki {
namespace Util {

const char * GetLogLevelString(LogLevel level)
{
  switch(level)
  {
    case LOG_LEVEL_DEBUG: { return "Debug"; }
    case LOG_LEVEL_INFO : { return "Info"; }
    case LOG_LEVEL_EVENT: { return "Event"; }
    case LOG_LEVEL_WARN : { return "Warn"; }
    case LOG_LEVEL_ERROR: { return "Error"; }
    case _LOG_LEVEL_COUNT: { break; }
  };

  DEV_ASSERT_MSG(false, "GetLogLevelString.InvalidLogLevel", "%d is not a valid level", level);
  return "Invalid_Log_Level!";
}

LogLevel GetLogLevelValue(const std::string& levelStr)
{
  std::string levelLC = levelStr;
  std::transform(levelLC.begin(), levelLC.end(), levelLC.begin(), ::tolower);
  if ( levelLC == "debug" ) {
    return LOG_LEVEL_DEBUG;
  } else if ( levelLC == "info" ) {
    return LOG_LEVEL_INFO;
  } else if ( levelLC == "event" ) {
    return LOG_LEVEL_EVENT;
  } else if ( levelLC == "warn" ) {
    return LOG_LEVEL_WARN;
  } else if ( levelLC == "error" ) {
    return LOG_LEVEL_ERROR;
  }

  DEV_ASSERT_MSG(false, "GetLogLevelValue.InvalidLogLevel", "'%s' is not a valid level", levelStr.c_str());
  return _LOG_LEVEL_COUNT;
}

} // end namespace Util
} // end namespace Anki
