/**
 * File: util/logging/logging.h
 *
 * Author: damjan
 * Created: 4/3/2014
 *
 * Description: logging functions.
 * structure of the function names is: s<Level><style>
 *   levels are: Event, Error, Warning, Info, Debug
 *   style: f - takes (...) , v - takes va_list
 *   functions are spelled out, instead of stacked (sErrorF -> calls sErrorV -> calls sError)
 *   to improve on the stack space. If you think about improving on this, please consider macros to re-use code.
 *   If you however feel that we should stack them into one set of functions that use LogLevel as a param, think about
 *   need to translate Ank::Util::LogLevel to DasLogLevel and then to ios/android LogLevel.
 *
 * Copyright: Anki, Inc. 2014
 *
 **/


#ifndef __Util_Logging_Logging_H_
#define __Util_Logging_Logging_H_

#include "util/global/globalDefinitions.h"
#include "util/logging/callstack.h"
#include "util/logging/logtypes.h"

#include "platform/anki-trace/tracing.h"

#include <string>
#include <vector>
#include <stdio.h>

#ifndef ALLOW_DEBUG_LOGGING
#define ALLOW_DEBUG_LOGGING ANKI_DEVELOPER_CODE
#endif

#if !defined(ANKI_BREADCRUMBS)
#define ANKI_BREADCRUMBS 0
#endif

#define MAX_LOG_STRING_LEN 1024

namespace Anki {
namespace Util {

#if ANKI_BREADCRUMBS
bool DropBreadcrumb(bool result, const char* file, int line);
#else
static inline bool DropBreadcrumb(bool result, const char*, int) { return result; }
#endif

class ITickTimeProvider;
class ILoggerProvider;
class ChannelFilter;
class IEventProvider;

std::string HexDump(const void *value, const size_t len, char delimiter);

extern ITickTimeProvider* gTickTimeProvider;
extern ILoggerProvider* gLoggerProvider;
extern IEventProvider* gEventProvider;

// Accessors for global error flag for unit testing
void sSetErrG();            // Sets errG to true
void sUnSetErrG();          // Sets errG to false
bool sGetErrG();            // Gets the value of errG
// Only useful if access to the global error flag involves some locking mechanism
void sPushErrG(bool value); // Sets errG to value and locks access by other threads until PopErrG is called
void sPopErrG();            // Restores errG to its value before calling sPushErrG and enables access to errG by other threads

// Global flag to control break-on-error behavior
extern bool _errBreakOnError;

// If true, access to the global error flag uses a mutex device. Changing this value while logging could
// lead to a mutex lock, so don't.
extern bool _lockErrG;

//
// "Event level" logging is no longer a thing. Do not use it.
// Messages intended for DAS should use the explicit DASMSG interface declared by util/logging/DAS.h.
//
__attribute__((__deprecated__))
__attribute__((__used__))
void sEventF(const char* name, const KVPairVector & keyvals, const char* format, ...) __attribute__((format(printf,3,4)));

__attribute__((__deprecated__))
__attribute__((__used__))
void sEventV(const char* name, const KVPairVector & keyvals, const char* format, va_list args) __attribute__((format(printf,3,0)));

__attribute__((__used__))
void sEvent(const char* name, const KVPairVector & keyvals, const char* strval);

__attribute__((__used__))
void sErrorF(const char* name, const KVPairVector & keyvals, const char* format, ...) __attribute__((format(printf,3,4)));

__attribute__((__used__))
void sErrorV(const char* name, const KVPairVector & keyvals, const char* format, va_list args) __attribute__((format(printf,3,0)));

__attribute__((__used__))
void sError(const char* name, const KVPairVector & keyvals, const char* strval);

__attribute__((__used__))
void sWarningF(const char* name, const KVPairVector & keyvals, const char* format, ...) __attribute__((format(printf,3,4)));

__attribute__((__used__))
void sWarningV(const char* name, const KVPairVector & keyvals, const char* format, va_list args) __attribute__((format(printf,3,0)));

__attribute__((__used__))
void sWarning(const char* name, const KVPairVector & keyvals, const char* strval);

__attribute__((__used__))
void sInfoF(const char* name, const KVPairVector & keyvals, const char* format, ...) __attribute__((format(printf,3,4)));

__attribute__((__used__))
void sInfoV(const char* name, const KVPairVector & keyvals, const char* format, va_list args) __attribute__((format(printf,3,0)));

__attribute__((__used__))
void sInfo(const char* name, const KVPairVector & keyvals, const char* strval);

__attribute__((__used__))
void sChanneledInfoF(const char* channel, const char* name, const KVPairVector & keyvals, const char* format, ...) __attribute__((format(printf,4,5)));

__attribute__((__used__))
void sChanneledInfoV(const char* channel, const char* name, const KVPairVector & keyvals, const char* format, va_list args) __attribute__((format(printf,4,0)));

__attribute__((__used__))
void sChanneledInfo(const char* channel, const char* name, const KVPairVector & keyvals, const char* strval);

__attribute__((__used__))
void sChanneledDebugF(const char* channel, const char* name, const KVPairVector & keyvals, const char* format, ...) __attribute__((format(printf,4,5)));

__attribute__((__used__))
void sChanneledDebugV(const char* channel, const char* name, const KVPairVector & keyvals, const char* format, va_list args) __attribute__((format(printf,4,0)));

__attribute__((__used__))
void sChanneledDebug(const char* channel, const char* name, const KVPairVector & keyvals, const char* strval);

// Helper for use with ANKI_VERIFY macro. Always returns true.
bool sVerifySucceededReturnTrue(const char* file, int line);

// Helper for use with ANKI_VERIFY macro. Always returns false.
__attribute__((__used__))
bool sVerifyFailedReturnFalse(const char* file, int line, const char* name, const char* format, ...) __attribute__((format(printf,4,5)));


void sSetGlobal(const char* key, const char* value);

//
// Anki::Util::sLogFlush()
// Perform synchronous flush of log data to underlying storage.
// This calls blocks until log data has been flushed.
//
void sLogFlush();

//
// Anki::Util::sDebugBreak()
// Break to debugger (if possible), then return to caller.
// If break to debugger is not supported, this function provides
// a convenient hook for developers to set a breakpoint by hand.
// This function is enabled for build configurations with ANKI_DEVELOPER_CODE=1.
// This function is a no-op for build configurations with ANKI_DEVELOPER_CODE=0.
//
void sDebugBreak();

// Anki::Util::sDebugBreakOnError()
// Calls sDebugBreak() in error situations if the configuration
// allows it. This is a separate function rather than a macro
// so that its behavior isn't affected by different configurations
// at different levels of the project (i.e. DriveEngine has a
// different setting than the OverDrive app or something like that)
//
void sDebugBreakOnError();

//
// Anki::Util::sAbort()
// Dump core (if possible) and terminate process.
// Does not flush buffers.
// Does not invoke exit handlers.
// Never returns to caller.
//
__attribute__((noreturn)) void sAbort();

} // namespace Util
} // namespace Anki

//
// Logging with names.
//
#define PRINT_NAMED_ERROR(name, format, ...) do { \
  if(ANKITRACE_ENABLED) { \
    char PRINT_LOG_VAR_logString[MAX_LOG_STRING_LEN];                   \
    snprintf(PRINT_LOG_VAR_logString, MAX_LOG_STRING_LEN, format, ##__VA_ARGS__); \
    tracelog(TRACE_ERR, "%s %s", name, PRINT_LOG_VAR_logString);        \
    ::Anki::Util::sErrorF(name, {}, "%s", PRINT_LOG_VAR_logString);     \
  } else { \
    ::Anki::Util::sErrorF(name, {}, format, ##__VA_ARGS__); \
  } \
  ::Anki::Util::DropBreadcrumb(false, __FILE__, __LINE__);       \
  ::Anki::Util::sSetErrG(); \
  if (::Anki::Util::_errBreakOnError) { \
    ::Anki::Util::sDebugBreakOnError(); \
  } \
} while(0)

#define PRINT_NAMED_WARNING(name, format, ...) do { \
  if(ANKITRACE_ENABLED) { \
    char PRINT_LOG_VAR_logString[MAX_LOG_STRING_LEN];\
    snprintf(PRINT_LOG_VAR_logString, MAX_LOG_STRING_LEN, format, ##__VA_ARGS__);\
    tracelog(TRACE_WARNING, "%s %s", name, PRINT_LOG_VAR_logString);\
    ::Anki::Util::sWarningF(name, {}, "%s", PRINT_LOG_VAR_logString);\
  } else { \
    ::Anki::Util::sWarningF(name, {}, format, ##__VA_ARGS__);\
  } \
  ::Anki::Util::DropBreadcrumb(false, __FILE__, __LINE__); \
} while(0)

#define PRINT_NAMED_INFO(name, format, ...) do { \
  if(ANKITRACE_ENABLED) { \
    char PRINT_LOG_VAR_logString[MAX_LOG_STRING_LEN];                   \
    snprintf(PRINT_LOG_VAR_logString, MAX_LOG_STRING_LEN, format, ##__VA_ARGS__); \
    tracelog(TRACE_INFO, "%s %s", name, PRINT_LOG_VAR_logString);       \
    ::Anki::Util::sChanneledInfoF(name, name, {}, "%s", PRINT_LOG_VAR_logString); \
  } else { \
    ::Anki::Util::sChanneledInfoF(name, name, {}, format, ##__VA_ARGS__); \
  } \
  ::Anki::Util::DropBreadcrumb(false, __FILE__, __LINE__);              \
} while(0)

#if ALLOW_DEBUG_LOGGING
#define PRINT_NAMED_DEBUG(name, format, ...) do { \
  if(ANKITRACE_ENABLED) { \
    char PRINT_LOG_VAR_logString[MAX_LOG_STRING_LEN]; \
    snprintf(PRINT_LOG_VAR_logString, MAX_LOG_STRING_LEN, format, ##__VA_ARGS__); \
    tracelog(TRACE_DEBUG, "%s %s", name, PRINT_LOG_VAR_logString); \
    ::Anki::Util::sChanneledDebugF(name, name, {}, "%s", PRINT_LOG_VAR_logString); \
  } else { \
    ::Anki::Util::sChanneledDebugF(name, name, {}, format, ##__VA_ARGS__); \
  } \
  ::Anki::Util::DropBreadcrumb(false, __FILE__, __LINE__);              \
} while(0)
#else
#define PRINT_NAMED_DEBUG(name, format, ...) do { \
  if(ANKITRACE_ENABLED) { \
    char PRINT_LOG_VAR_logString[MAX_LOG_STRING_LEN]; \
    snprintf(PRINT_LOG_VAR_logString, MAX_LOG_STRING_LEN, format, ##__VA_ARGS__); \
    tracelog(TRACE_DEBUG, "%s %s", name, PRINT_LOG_VAR_logString); \
  } \
  ::Anki::Util::DropBreadcrumb(false, __FILE__, __LINE__);  \
} while(0)
#endif

//
// ANKI_VERIFY(expr, name, format, args...)
// Helper macro for simple error checks / assertions.
// Similar to DEV_ASSERT (below) but enabled in both debug and release (and shipping) builds.
//
// If the conditional expression (expr) is true, ANKI_VERIFY returns true.
// If the conditional expression (expr) is false, ANKI_VERIFY logs an error message, dumps the callstack,
// and returns false.
//
// The conditional expression (expr) will be evaluated only once. Arguments required to produce the formatted
// string for the log will only be evaluated if expr==false.
//
// Example 1:
// Use
//   if (ANKI_VERIFY(x == y, "VerifyXY", "%p != %p", x, y)) {
//     /* do stuff */
//   }
// in place of
//   if (x == y) {
//     /* do stuff */
//   } else {
//     PRINT_NAMED_ERROR("VerifyXY", "%p != p", x, y);
//     sDumpCallstack();
//   }
//
// Example 2:
// Use
//   if (!ANKI_VERIFY(x == y, "VerifyXY", "%p != %p", x, y)) {
//     return FAIL;
//   }
// in place of
//   if (x != y) {
//     PRINT_NAMED_ERROR("VerifyXY", "%p != %p", x, y);
//     sDumpCallstack();
//     return FAIL;
//   }
//
// Note that "&& false" is used to inform static analysis that the "verify failed" branch always returns false.
// This prevents analyzers from generating bogus warnings caused by impossible code paths.
//
#define ANKI_VERIFY(expr, name, format, ...) \
 (expr ? ::Anki::Util::sVerifySucceededReturnTrue(__FILE__, __LINE__) : (::Anki::Util::sVerifyFailedReturnFalse(__FILE__, __LINE__, name, "VERIFY(%s): " format, #expr, ##__VA_ARGS__) && false))

//
// Logging with channels.
//
#define PRINT_CH_INFO(channel, name, format, ...) do { \
  if(ANKITRACE_ENABLED) { \
    char PRINT_LOG_VAR_logString[MAX_LOG_STRING_LEN];                   \
    snprintf(PRINT_LOG_VAR_logString, MAX_LOG_STRING_LEN, format, ##__VA_ARGS__); \
    tracelog(TRACE_INFO, "%s %s %s", channel, name, PRINT_LOG_VAR_logString); \
    ::Anki::Util::sChanneledInfoF(channel, name, {}, "%s", PRINT_LOG_VAR_logString); \
  } else { \
    ::Anki::Util::sChanneledInfoF(channel, name, {}, format, ##__VA_ARGS__); \
  } \
  ::Anki::Util::DropBreadcrumb(false, __FILE__, __LINE__);  \
} while(0)

#if ALLOW_DEBUG_LOGGING
#define PRINT_CH_DEBUG(channel, name, format, ...) do { \
  if(ANKITRACE_ENABLED) {                                               \
    char PRINT_LOG_VAR_logString[MAX_LOG_STRING_LEN];                   \
    snprintf(PRINT_LOG_VAR_logString, MAX_LOG_STRING_LEN, format, ##__VA_ARGS__); \
    tracelog(TRACE_DEBUG, "%s %s %s", channel, name, PRINT_LOG_VAR_logString); \
    ::Anki::Util::sChanneledDebugF(channel, name, {}, "%s", PRINT_LOG_VAR_logString); \
  } else { \
    ::Anki::Util::sChanneledDebugF(channel, name, {}, format, ##__VA_ARGS__); \
  } \
  ::Anki::Util::DropBreadcrumb(false, __FILE__, __LINE__); \
} while(0)
#else
#define PRINT_CH_DEBUG(channel, name, format, ...) do { \
  if(ANKITRACE_ENABLED) {                                               \
    char PRINT_LOG_VAR_logString[MAX_LOG_STRING_LEN];                   \
    snprintf(PRINT_LOG_VAR_logString, MAX_LOG_STRING_LEN, format, ##__VA_ARGS__); \
    tracelog(TRACE_DEBUG, "%s %s %s", channel, name, PRINT_LOG_VAR_logString); \
  } \
  ::Anki::Util::DropBreadcrumb(false, __FILE__, __LINE__);   \
} while(0)
#endif

//
// Periodic logging with channels.
//

// Helper used by debug/info versions below
#define PRINT_PERIODIC_CH_HELPER(func, period, channel, name, format, ...) \
{ static u16 cnt = period;                                                 \
  if (++cnt >= period) {                                                   \
    ::Anki::Util::DropBreadcrumb(false, __FILE__, __LINE__);                 \
    ::Anki::Util::func(channel, name, {}, format, ##__VA_ARGS__);          \
    cnt = 0;                                                               \
  }                                                                        \
}

// Actually use these in your code (not the helper above)
#define PRINT_PERIODIC_CH_INFO(period, channel, name, format, ...)  \
  if(ANKITRACE_ENABLED) {                                               \
    char PRINT_LOG_VAR_logString[MAX_LOG_STRING_LEN];                     \
    snprintf(PRINT_LOG_VAR_logString, MAX_LOG_STRING_LEN, format, ##__VA_ARGS__); \
    tracelog(TRACE_INFO, "%d %s %s %s", period, channel, name, PRINT_LOG_VAR_logString); \
    PRINT_PERIODIC_CH_HELPER(sChanneledInfoF, period, channel, name, "%s", PRINT_LOG_VAR_logString) \
  } else { \
    PRINT_PERIODIC_CH_HELPER(sChanneledInfoF, period, channel, name, format, ##__VA_ARGS__) \
  }

#define PRINT_PERIODIC_CH_DEBUG(period, channel, name, format, ...) \
  if(ANKITRACE_ENABLED) {                                               \
    char PRINT_LOG_VAR_logString[MAX_LOG_STRING_LEN];                                   \
    snprintf(PRINT_LOG_VAR_logString, MAX_LOG_STRING_LEN, format, ##__VA_ARGS__); \
    tracelog(TRACE_DEBUG, "%d %s %s %s", period, channel, name, PRINT_LOG_VAR_logString); \
    PRINT_PERIODIC_CH_HELPER(sChanneledDebugF, period, channel, name, "%s", PRINT_LOG_VAR_logString) \
  } else { \
      PRINT_PERIODIC_CH_HELPER(sChanneledDebugF, period, channel, name, format, ##__VA_ARGS__) \
  }

// Streams
#define PRINT_STREAM_ERROR(name, args) do{         \
      std::stringstream ss; ss<<args;                   \
      tracelog(TRACE_ERR, "%s %s", name, ss.str().c_str());    \
      ::Anki::Util::DropBreadcrumb(false, __FILE__, __LINE__); \
      ::Anki::Util::sError(name, {}, ss.str().c_str()); \
    } while(0)

#define PRINT_STREAM_WARNING(name, args) do{       \
      std::stringstream ss; ss<<args;                   \
      tracelog(TRACE_WARNING, "%s %s", name, ss.str().c_str());\
      ::Anki::Util::DropBreadcrumb(false, __FILE__, __LINE__); \
      ::Anki::Util::sWarning(name, {}, ss.str().c_str()); \
    } while(0)

#define PRINT_STREAM_INFO(name, args) do{          \
      std::stringstream ss; ss<<args;                   \
      tracelog(TRACE_INFO, "%s %s", name, ss.str().c_str()); \
      ::Anki::Util::DropBreadcrumb(false, __FILE__, __LINE__); \
      ::Anki::Util::sChanneledInfo(name, name, {}, ss.str().c_str()); \
    } while(0)

#if ALLOW_DEBUG_LOGGING
#define PRINT_STREAM_DEBUG(name, args) do {         \
      std::stringstream ss; ss<<args;                   \
      tracelog(TRACE_DEBUG, "%s %s", name, ss.str().c_str());   \
      ::Anki::Util::DropBreadcrumb(false, __FILE__, __LINE__); \
      ::Anki::Util::sChanneledDebug(name, name, {}, ss.str().c_str()); \
    } while(0)
#else
#define PRINT_STREAM_DEBUG(eventName, args)
#endif

// Auto streams
#if ALLOW_DEBUG_LOGGING
#define PRINT_AUTOSTREAM_DEBUG(args) do { \
    char nameBuf[256]; GENERATE_EVENT_NAME(nameBuf, sizeof(nameBuf)); \
    PRINT_STREAM_DEBUG(nameBuf, args); } while(0)
#else
#define PRINT_AUTOSTREAM_DEBUG(args)
#endif

#define SHORT_FILE ( strrchr(__FILE__, '/') ? strrchr(__FILE__, '/') + 1 : __FILE__ )
#define GENERATE_EVENT_NAME(nameBuf, nameBufLen) { snprintf(nameBuf, nameBufLen, "%s.%s.%d", (SHORT_FILE), __FUNCTION__, __LINE__); }


// Anki assert definition
#if defined(NDEBUG) || defined(UNIT_TEST)
#define DEBUG_ABORT ((void)0)
#else
#define DEBUG_ABORT __builtin_trap()
#endif

#define ASSERT_NAMED(expr, name) do {                       \
  ::Anki::Util::DropBreadcrumb(expr ? true : false, __FILE__, __LINE__); \
  if (!(expr)) {                                            \
    PRINT_NAMED_ERROR(name, "Assertion Failed: %s", #expr); \
    Anki::Util::sDumpCallstack("AssertCallstack");          \
    Anki::Util::sLogFlush();                                \
    DEBUG_ABORT;                                            \
  }                                                         \
} while(0)

#define ASSERT_NAMED_AND_RETURN_FALSE_IF_FAIL(exp, name) do { \
  ::Anki::Util::DropBreadcrumb(false, __FILE__, __LINE__);    \
  if(!(exp)) {                                              \
    PRINT_NAMED_ERROR(name, "Assertion Failed: %s", #exp);  \
    Anki::Util::sDumpCallstack("AssertCallstack");          \
    Anki::Util::sLogFlush();                                \
    DEBUG_ABORT;                                            \
    return false;                                           \
  }                                                         \
}while(0)


#define ASSERT_NAMED_EVENT(expr, name, format, ...) do {                      \
  ::Anki::Util::DropBreadcrumb(expr ? true : false, __FILE__, __LINE__);        \
  if (!(expr)) {                                                              \
    PRINT_NAMED_ERROR(name, "ASSERT ( %s ): " format, #expr, ##__VA_ARGS__);  \
    Anki::Util::sDumpCallstack("AssertCallstack");                            \
    Anki::Util::sLogFlush();                                                  \
    DEBUG_ABORT;                                                              \
  }                                                                           \
} while(0)


#define ASSERT_NAMED_EVENT_AND_RETURN_FALSE_IF_FAIL(exp, name, format, ...) do { \
  ::Anki::Util::DropBreadcrumb(exp ? true : false, __FILE__, __LINE__);         \
  if(!(exp)) {                                                                \
    PRINT_NAMED_ERROR(name, "ASSERT ( %s ): " format, #exp, ##__VA_ARGS__);   \
    Anki::Util::sDumpCallstack("AssertCallstack");                            \
    Anki::Util::sLogFlush();                                                  \
    DEBUG_ABORT;                                                              \
    return false;                                                             \
  }                                                                           \
}while(0)


//
// Developer assertions are compiled for debug builds ONLY.
// Developer assertions are discarded for release and shipping builds.
//
// Code blocks that are only used for developer assertions should be guarded with #if ANKI_DEV_ASSERT_ENABLED.
// Variables that are only used for developer assertions should be guarded with DEV_ASSERT_ONLY.

#ifndef ANKI_DEV_ASSERT_ENABLED
  #define ANKI_DEV_ASSERT_ENABLED ANKI_DEVELOPER_CODE
#endif

#if ANKI_DEV_ASSERT_ENABLED

#define DEV_ASSERT_MSG(expr, name, format, ...) do { \
  if (!(expr)) { \
    PRINT_NAMED_ERROR(name, "ASSERT(%s): " format, #expr, ##__VA_ARGS__); \
    Anki::Util::sDumpCallstack("ASSERT"); \
    Anki::Util::sLogFlush(); \
    Anki::Util::sAbort(); \
  } \
} while (0)

#define DEV_ASSERT_ONLY(expr) expr

#else

//
// Code within "if false" will be analyzed by compiler, so variables are counted as "used",
// but the entire block will be discarded by the optimizer because it can't be executed.
//
#define DEV_ASSERT_MSG(expr, name, format, ...) do { \
  if (false) { \
    if (!(expr)) { \
      PRINT_NAMED_ERROR(name, "ASSERT(%s): " format, #expr, ##__VA_ARGS__); \
    } \
  } \
} while (0)

#define DEV_ASSERT_ONLY(expr)

#endif

#define DEV_ASSERT(expr, name) DEV_ASSERT_MSG(expr, name, "Assertion failed")

//
// DAS events are structured messages for use with backend analytics.
// Event name and data fields are determined by the analytics team.
//
#define DPHYS "$phys"
#define DDATA "$data"
#define DGROUP "$group"
#define DGAME "$game"
#define DCONNECTSESSION "$session_id"

//
// Compact version of PRINT_NAMED_ERROR & friends
// LOG_INFO and LOG_DEBUG assume you declared something like
//  #define LOG_CHANNEL "Name"
// near the top of your cpp file.
//
#define LOG_ERROR(name, fmt, ...)   PRINT_NAMED_ERROR(name, fmt, ##__VA_ARGS__)
#define LOG_WARNING(name, fmt, ...) PRINT_NAMED_WARNING(name, fmt, ##__VA_ARGS__)
#define LOG_INFO(name, fmt, ...)    PRINT_CH_INFO(LOG_CHANNEL, name, fmt, ##__VA_ARGS__)
#define LOG_DEBUG(name, fmt, ...)   PRINT_CH_DEBUG(LOG_CHANNEL, name, fmt, ##__VA_ARGS__)

#define LOG_PERIODIC_INFO(period, name, fmt, ...) \
  PRINT_PERIODIC_CH_INFO(period, LOG_CHANNEL, name, fmt, ##__VA_ARGS__)
#define LOG_PERIODIC_DEBUG(period, name, fmt, ...) \
  PRINT_PERIODIC_CH_DEBUG(period, LOG_CHANNEL, name, fmt, ##__VA_ARGS__)

//
// Compact version of PRINT_CH_INFO & PRINT_CH_DEBUG macros.
// These macros can be used in header files or templates
// where "#define LOG_CHANNEL" is not appropriate.
//
#define LOG_CH_INFO(ch, name, fmt, ...)  PRINT_CH_INFO(ch, name, fmt, ##__VA_ARGS__)
#define LOG_CH_DEBUG(ch, name, fmt, ...) PRINT_CH_DEBUG(ch, name, fmt, ##__VA_ARGS__)

#endif // __Util_Logging_Logging_H_
