/**
 * File: util/logging/DAS.cpp
 *
 * Description: DAS extensions for Util::Logging macros
 *
 * Copyright: Anki, Inc. 2018
 *
 **/

#include "DAS.h"
#include "json/json.h"
#include <time.h>

namespace Anki {
namespace Util {
namespace DAS {

//
// If boot clock is available, use it, else fall back to monotonic clock
//
#if defined(CLOCK_BOOTTIME)
#define CLOCK CLOCK_BOOTTIME
#else
#define CLOCK CLOCK_MONOTONIC
#endif

uint64_t UptimeMS()
{
  struct timespec ts{0};
  clock_gettime(CLOCK, &ts);
  return (ts.tv_sec * 1000) + (ts.tv_nsec/1000000);
}

std::string Escape(const char * str)
{
  // Use JsonCpp to quote string, then remove surrounding quotes
  const std::string & quotedString = Json::valueToQuotedString(str ? str : "");
  return quotedString.substr(1, quotedString.size()-2);
}

std::string Escape(const std::string & str)
{
  return Escape(str.c_str());
}


} // end namespace DAS
} // end namespace Util
} // end namespace Anki
