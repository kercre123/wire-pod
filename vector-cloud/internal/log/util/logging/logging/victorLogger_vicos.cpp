/**
* File: util/logging/victorLogger_vicos.cpp
*
* Description: Implements ILoggerProvider for Victor
*
* Copyright: Anki, inc. 2018
*
*/

#include "util/logging/victorLogger.h"
#include "util/logging/DAS.h"
#include "util/global/globalDefinitions.h"

#include <android/log.h>
#include <assert.h>
#include <vector>

namespace Anki {
namespace Util {

VictorLogger::VictorLogger(const std::string& tag) :
  _tag(tag)
{

}

void VictorLogger::Log(android_LogPriority prio,
  const char * channel,
  const char * name,
  const KVPairVector & keyvals,
  const char * strval)
{
  // Channel, name, strval may not be null
  assert(channel != nullptr);
  assert(name != nullptr);
  assert(strval != nullptr);

  __android_log_print(prio, _tag.c_str(), "[@%s] %s: %s", channel, name, strval);
}

void VictorLogger::Log(android_LogPriority prio,
  const char * name,
  const KVPairVector & keyvals,
  const char * strval)
{
  // Name, strval may not be null
  assert(name != nullptr);
  assert(strval != nullptr);

  __android_log_print(prio, _tag.c_str(), "%s: %s", name, strval);
}

void VictorLogger::LogEvent(android_LogPriority prio,
  const char * name,
  const KVPairVector & keyvals)
{
  // Name may not be null
  assert(name != nullptr);

  //
  // Marshal values for each DAS v2 event fields that must be provided.
  // Note some fields will be provided by the log record itself,
  // while others will be provided by the aggregator.
  //
  //const char * SOURCE = ""; (represented by android log tag)
  //const char * TS = ""; (represented by android log timestamp)
  //const char * LEVEL = ""; (represented by android log level)
  //const char * ROBOT = ""; (provided by aggregator)
  //const char * ROBOT_VERSION = ""; (provided by aggregator)
  //const char * SEQ = seq; (provided by aggregator)
  //const char * PROFILE_ID = ""; (provided by aggregator)
  //const char * FEATURE_TYPE = ""; (provided by aggregator)
  //const char * FEATURE_RUN_ID = ""; (provided by aggregator)
  //const char * EVENT = name; (provided by caller)
  const char * str1 = "";
  const char * str2 = "";
  const char * str3 = "";
  const char * str4 = "";
  const char * int1 = "";
  const char * int2 = "";
  const char * int3 = "";
  const char * int4 = "";

  //
  // Iterate key-value pairs to see if they define additional event fields.
  // Note that key names must be an EXACT MATCH for values declared by Anki::Util::DAS namespace.
  //
  // TO DO: Replace with hash or enum lookup?
  //
  for (const auto & kv : keyvals) {
    // Key, value may not be null
    assert(kv.first != nullptr);
    assert(kv.second != nullptr);
    const char * key = kv.first;
    if (strcmp(key, DAS::STR1) == 0) {
      str1 = kv.second;
    } else if (strcmp(key, DAS::STR2) == 0) {
      str2 = kv.second;
    } else if (strcmp(key, DAS::STR3) == 0) {
      str3 = kv.second;
    } else if (strcmp(key, DAS::STR4) == 0) {
      str4 = kv.second;
    } else if (strcmp(key, DAS::INT1) == 0) {
      int1 = kv.second;
    } else if (strcmp(key, DAS::INT2) == 0) {
      int2 = kv.second;
    } else if (strcmp(key, DAS::INT3) == 0) {
      int3 = kv.second;
    } else if (strcmp(key, DAS::INT4) == 0) {
      int4 = kv.second;
    } else {
      // Any unknown keys are ignored
    }
  }

  // Format fields into a compact CSV format.
  // Leading @ serves as a hint that this row is in compact CSV format.
  //
  // We use a fixed format string for performance, but we need to update the format string
  // if event format ever changes.
  static_assert(Anki::Util::DAS::EVENT_MARKER == '@', "DAS event marker does not match declarations");
  static_assert(Anki::Util::DAS::FIELD_MARKER == '\x1F', "DAS field marker does not match declarations");
  static_assert(Anki::Util::DAS::FIELD_COUNT == 9, "DAS field count does not match declarations");

  const auto uptime_ms = DAS::UptimeMS();

  __android_log_print(prio, _tag.c_str(), "@%s\x1F%s\x1F%s\x1F%s\x1F%s\x1F%s\x1F%s\x1F%s\x1F%s\x1F%llu",
                      name, str1, str2, str3, str4, int1, int2, int3, int4, uptime_ms);

}

void VictorLogger::LogError(android_LogPriority prio,
  const char * name,
  const KVPairVector & keyvals,
  const char * strval)
{
#if ANKI_REPORT_ERRORS_TO_DAS
  // Normally errors just go to the local log system, but with this enabled
  // we will send them up to the DAS server.
  KVPairVector kv;
  kv.emplace_back(DAS::STR1, name);
#if ANKI_REPORT_ERRORS_WITH_STRVAL_TO_DAS
  // Move strval to s2 value for DAS event.  This can be useful for
  // debugging the error.  It is expected that this will NOT be used
  // in shipping builds.
  std::string escaped = Anki::Util::DAS::Escape(strval);
  kv.emplace_back(DAS::STR2, escaped.c_str());
  // Append any existing keyvals to the end of kv.  If keyvals has a value for s1 or s2,
  // they will supersede what is in kv when processed by LogEvent
#endif // ANKI_REPORT_ERRORS_WITH_STRVAL_TO_DAS
  kv.insert(std::end(kv), std::begin(keyvals), std::end(keyvals));
  LogEvent(prio, "log.error", kv);
#else
  Log(prio, name, keyvals, strval);
#endif // ANKI_REPORT_ERRORS_TO_DAS
}

//
// Map Anki log level to android log priority
//
inline android_LogPriority GetLogPrio(LogLevel level)
{
  android_LogPriority prio = ANDROID_LOG_UNKNOWN;
  switch (level)
  {
    case LOG_LEVEL_ERROR:
      prio = ANDROID_LOG_ERROR;
      break;
    case LOG_LEVEL_WARN:
      prio = ANDROID_LOG_WARN;
      break;
    case LOG_LEVEL_EVENT:
    case LOG_LEVEL_INFO:
      prio = ANDROID_LOG_INFO;
      break;
    case LOG_LEVEL_DEBUG:
      prio = ANDROID_LOG_DEBUG;
      break;
    case _LOG_LEVEL_COUNT:
      break;
  }
  DEV_ASSERT(prio != ANDROID_LOG_UNKNOWN, "VictorLogger.GetLogPrio.UnknownLogLevel");
  return prio;
}

void VictorLogger::LogEvent(LogLevel level, const DasMsg & dasMsg)
{
  //
  // Marshal values for each DAS v2 event fields that must be provided.
  // Note some fields will be provided by the log record itself,
  // while others will be provided by the aggregator.
  //
  //const char * SOURCE = ""; (represented by android log tag)
  //const char * TS = ""; (represented by android log timestamp)
  //const char * LEVEL = ""; (represented by android log level)
  //const char * ROBOT = ""; (provided by aggregator)
  //const char * ROBOT_VERSION = ""; (provided by aggregator)
  //const char * SEQ = seq; (provided by aggregator)
  //const char * PROFILE_ID = ""; (provided by aggregator)
  //const char * FEATURE_TYPE = ""; (provided by aggregator)
  //const char * FEATURE_RUN_ID = ""; (provided by aggregator)
  const auto prio = GetLogPrio(level);

  // Format fields into a compact CSV format.
  // Leading @ serves as a hint that this row is in compact CSV format.
  //
  // We use a fixed format string for performance, but we need to update the format string
  // if event format ever changes.
  static_assert(Anki::Util::DAS::EVENT_MARKER == '@', "DAS event marker does not match declarations");
  static_assert(Anki::Util::DAS::FIELD_MARKER == '\x1F', "DAS field marker does not match declarations");
  static_assert(Anki::Util::DAS::FIELD_COUNT == 9, "DAS field count does not match declarations");

  const auto uptime_ms = Anki::Util::DAS::UptimeMS();

  __android_log_print(prio, _tag.c_str(), "@%s\x1F%s\x1F%s\x1F%s\x1F%s\x1F%s\x1F%s\x1F%s\x1F%s\x1F%llu",
                      dasMsg.event.c_str(), dasMsg.s1.c_str(), dasMsg.s2.c_str(), dasMsg.s3.c_str(), dasMsg.s4.c_str(),
                      dasMsg.i1.c_str(), dasMsg.i2.c_str(), dasMsg.i3.c_str(), dasMsg.i4.c_str(), uptime_ms);

}

void VictorLogger::SetGlobal(const char * key, const char * value)
{
  // Key may not be null
  assert(key != nullptr);

  std::lock_guard<std::mutex> lock(_mutex);

  if (value == nullptr) {
    _globals.erase(key);
  } else {
    _globals.emplace(std::pair<std::string,std::string>{key, value});
  }
}

void VictorLogger::GetGlobals(std::map<std::string, std::string> & globals)
{
  std::lock_guard<std::mutex> lock(_mutex);
  globals = _globals;
}

} // end namespace Util
} // end namespace Anki
