/**
 * File: util/logging/logtypes.h
 *
 * Description: Declaration of Anki::Util log types
 *
 * These declarations are shared by various implementations of Anki::Util
 * log facilities. They are broken out into a separate file so they
 * can be used by log facilities without bringing in all the overhead
 * of logging.h.
 *
 * Copyright: Anki, Inc. 2018
 *
 **/


#ifndef __util_logging_logtypes_h
#define __util_logging_logtypes_h

#include <string>
#include <vector>

namespace Anki {
namespace Util {

enum LogLevel {
  LOG_LEVEL_DEBUG = 0,
  LOG_LEVEL_INFO,
  LOG_LEVEL_EVENT,
  LOG_LEVEL_WARN,
  LOG_LEVEL_ERROR,

  _LOG_LEVEL_COUNT // control field. Do not use for regular logging
};

using KVPair = std::pair<const char *, const char *>;
using KVPairVector = std::vector<KVPair>;

// returns a descriptive string for the given log level / or level from string
const char * GetLogLevelString(LogLevel level);
LogLevel GetLogLevelValue(const std::string& levelStr);

} // end namespace Util
} // end namespace Anki

#endif // __util_logging_logtypes_h
