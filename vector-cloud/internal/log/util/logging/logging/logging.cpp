/**
 * File: util/logging/logging.cpp
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
 *   If you however feel that we should stack them into one set of function that uses LogLevel as a param, think about need
 *   to translate Ank::Util::LogLevel to DasLogLevel and then to ios/android LogLevel.
 *
 * Copyright: Anki, Inc. 2014
 *
 **/

#include "util/logging/logging.h"
#include "util/logging/iTickTimeProvider.h"
#include "util/logging/iLoggerProvider.h"
#include "util/logging/channelFilter.h"
#include "util/logging/iEventProvider.h"
#include "util/helpers/ankiDefines.h"

#include <cstdlib>
#include <iostream>
#include <iomanip>
#include <sstream>
#include <signal.h>

#include <sys/time.h>

namespace Anki {
namespace Util {

std::string HexDump(const void *value, const size_t len, char delimiter)
{
  const unsigned char *bytes = (const unsigned char *) value;
  size_t bufferLen = len * 3;
  char *str = (char *) malloc(sizeof (char) * bufferLen);
  memset(str, 0, bufferLen);

  const char *hex = "0123456789ABCDEF";
  char *s = str;

  for (size_t i = 0; i < len - 1; ++i) {
    *s++ = hex[(*bytes >> 4)&0xF];
    *s++ = hex[(*bytes++)&0xF];
    *s++ = delimiter;
  }
  *s++ = hex[(*bytes >> 4)&0xF];
  *s++ = hex[(*bytes++)&0xF];
  *s++ = 0;

  std::string cppString(str);
  free(str);
  return cppString;
}

ITickTimeProvider * gTickTimeProvider = nullptr;
ILoggerProvider * gLoggerProvider = nullptr;
IEventProvider * gEventProvider = nullptr;

// Has an error been reported?
bool _errG = false;

// Do we break on any error?
bool _errBreakOnError = true;

// If true, access to _errG uses a mutex device
bool _lockErrG = false;

// Cached _errG during sPushErrG and sPopErrG
std::vector<bool> sOldErrG;

std::recursive_mutex sErrGMutex;

const size_t kMaxStringBufferSize = 1024;

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
// helpers
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
namespace {
using KVV = std::vector<std::pair<const char*, const char*>>;

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

void AddTickCount(std::ostringstream& oss)
{
  if (gTickTimeProvider != nullptr) {
    oss << "(tc";
    oss << std::right << std::setw(4) << std::setfill('0') << gTickTimeProvider->GetTickCount();
    oss << ") ";
  }
}

std::string PrependTickCount(const char * logString)
{
  if (gTickTimeProvider != nullptr) {
    std::ostringstream oss;
    AddTickCount(oss);
    oss << logString;
    return oss.str();
  }
  return logString;
}

void LogError(const char* name, const KVV& keyvals, const char* logString)
{
  if (nullptr == gLoggerProvider) {
    return;
  }

  gLoggerProvider->PrintLogE(name, keyvals, PrependTickCount(logString).c_str());
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void LogWarning(const char* name, const KVV& keyvals, const char* logString)
{
  if (gLoggerProvider == nullptr) {
    return;
  }

  gLoggerProvider->PrintLogW(name, keyvals, PrependTickCount(logString).c_str());
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void LogChanneledInfo(const char* channel, const char* name, const KVV& keyvals, const char* logString)
{
  if (nullptr == gLoggerProvider) {
    return;
  }

  // set tick count and channel name if available
  if (gTickTimeProvider != nullptr) {
    std::ostringstream finalLogStr;
    AddTickCount(finalLogStr);
    finalLogStr << logString;
    gLoggerProvider->PrintChanneledLogI(channel, name, keyvals, finalLogStr.str().c_str());
  } else {
    gLoggerProvider->PrintChanneledLogI(channel, name, keyvals, logString);
  }
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void LogChannelDebug(const char* channel, const char* name, const KVV& keyvals, const char* logString)
{
  if (nullptr == gLoggerProvider) {
    return;
  }

  gLoggerProvider->PrintChanneledLogD(channel, name, keyvals, PrependTickCount(logString).c_str());
}

}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void sEventF(const char* name, const KVV& keyvals, const char* format, ...)
{
  if (nullptr == gLoggerProvider) {
    return;
  }
  // event is BI event, and the data is specifically formatted to be read on the backend.
  // we should not modify tis data under any circumstance. Hence, no tick timer here
  va_list args;
  char logString[kMaxStringBufferSize];
  va_start(args, format);
  vsnprintf(logString, kMaxStringBufferSize, format, args);
  va_end(args);
  gLoggerProvider->PrintEvent(name, keyvals, logString);
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void sEventV(const char* name, const KVV& keyvals, const char* format, va_list args)
{
  if (nullptr == gLoggerProvider) {
    return;
  }
  // event is BI event, and the data is specifically formatted to be read on the backend.
  // we should not modify tis data under any circumstance. Hence, no tick timer here
  char logString[kMaxStringBufferSize];
  vsnprintf(logString, kMaxStringBufferSize, format, args);
  gLoggerProvider->PrintEvent(name, keyvals, logString);
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void sEvent(const char* name, const KVV& keyvals, const char* strval)
{
  if (nullptr == gLoggerProvider) {
    return;
  }
  gLoggerProvider->PrintEvent(name, keyvals, strval);
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void sErrorF(const char* name, const KVV& keyvals, const char* format, ...)
{
  if (nullptr == gLoggerProvider) {
    return;
  }

  // parse string
  va_list args;
  char logString[kMaxStringBufferSize];
  va_start(args, format);
  vsnprintf(logString, kMaxStringBufferSize, format, args);
  va_end(args);

  // log it
  LogError(name, keyvals, logString);
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void sErrorV(const char* name, const KVV& keyvals, const char* format, va_list args)
{
  if (nullptr == gLoggerProvider) {
    return;
  }

  // parse string
  char logString[kMaxStringBufferSize];
  vsnprintf(logString, kMaxStringBufferSize, format, args);

  // log it
  LogError(name, keyvals, logString);
}



// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void sError(const char* name, const KVV& keyvals, const char* strval)
{
  if (nullptr == gLoggerProvider) {
    return;
  }

  // log it
  LogError(name, keyvals, strval);
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void sWarningF(const char* name, const KVV& keyvals, const char* format, ...)
{
  if (nullptr == gLoggerProvider) {
    return;
  }

  // parse string
  va_list args;
  char logString[kMaxStringBufferSize];
  va_start(args, format);
  vsnprintf(logString, kMaxStringBufferSize, format, args);
  va_end(args);

  // log it
  LogWarning(name, keyvals, logString);
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void sWarningV(const char* name, const KVV& keyvals, const char* format, va_list args)
{
  if (nullptr == gLoggerProvider) {
    return;
  }

  // parse string
  char logString[kMaxStringBufferSize];
  vsnprintf(logString, kMaxStringBufferSize, format, args);

  // log it
  LogWarning(name, keyvals, logString);
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void sWarning(const char* name, const KVV& keyvals, const char* strval)
{
  if (nullptr == gLoggerProvider) {
    return;
  }

  // log it
  LogWarning(name, keyvals, strval);
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void sInfoF(const char* name, const KVV& keyvals, const char* format, ...)
{
  if (nullptr == gLoggerProvider) {
    return;
  }

  va_list args;
  va_start(args, format);
  sInfoV(name, keyvals, format, args);
  va_end(args);
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void sInfoV(const char* name, const KVV& keyvals, const char* format, va_list args)
{
  if (nullptr == gLoggerProvider) {
    return;
  }

  // format string
  char logString[kMaxStringBufferSize];
  vsnprintf(logString, kMaxStringBufferSize, format, args);

  // log it
  sInfo(name, keyvals, logString);
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void sInfo(const char* name, const KVV& keyvals, const char* strval)
{
  if (nullptr == gLoggerProvider) {
    return;
  }

  // log it
  LogChanneledInfo(name, name, keyvals, strval);
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

void sChanneledInfoF(const char* channel, const char* name, const KVV& keyvals, const char* format, ...)
{
  va_list args;
  va_start(args, format);
  sChanneledInfoV(channel, name, keyvals, format, args);
  va_end(args);
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void sChanneledInfoV(const char* channel, const char* name, const KVV& keyvals, const char* format, va_list args)
{
  if (nullptr == gLoggerProvider) {
    return;
  }

  // parse string
  char logString[kMaxStringBufferSize];
  vsnprintf(logString, kMaxStringBufferSize, format, args);

  // log it
  LogChanneledInfo(channel, name, keyvals, logString);
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void sChanneledInfo(const char* channel, const char* name, const KVV& keyvals, const char* strval)
{
  if (nullptr == gLoggerProvider) {
    return;
  }

  // log it
  LogChanneledInfo(channel, name, keyvals, strval);
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void sChanneledDebugF(const char* channel, const char* name, const KVV& keyvals, const char* format, ...)
{
  #if ALLOW_DEBUG_LOGGING
  if (nullptr == gLoggerProvider) {
    return;
  }

  // parse string
  va_list args;
  char logString[kMaxStringBufferSize];
  va_start(args, format);
  vsnprintf(logString, kMaxStringBufferSize, format, args);
  va_end(args);

  // log it
  LogChannelDebug(channel, name, keyvals, logString);
  #endif
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void sChanneledDebugV(const char* channel, const char* name, const KVV& keyvals, const char* format, va_list args)
{
  if (nullptr == gLoggerProvider) {
    return;
  }

  // parse string
  char logString[kMaxStringBufferSize];
  vsnprintf(logString, kMaxStringBufferSize, format, args);

  // log it
  LogChannelDebug(channel, name, keyvals, logString);
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void sChanneledDebug(const char* channel, const char* name, const KVV& keyvals, const char* strval)
{
  if (nullptr == gLoggerProvider) {
    return;
  }

  // log it
  LogChannelDebug(channel, name, keyvals, strval);
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
bool sVerifySucceededReturnTrue(const char* file, int line)
{
  Anki::Util::DropBreadcrumb(true, file, line);
  return true;
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
bool sVerifyFailedReturnFalse(const char* file, int line, const char* name, const char* format, ...)
{
  Anki::Util::DropBreadcrumb(false, file, line);

  va_list args;
  va_start(args, format);
  sErrorV(name, {}, format, args);
  va_end(args);
  sSetErrG();
  sDumpCallstack("VERIFY");
  sLogFlush();
  if (_errBreakOnError) {
    sDebugBreak();
  }
  return false;
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void sLogFlush()
{
  if (nullptr == gLoggerProvider) {
    return;
  }
  gLoggerProvider->Flush();
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
void sLogError(const DasMsg& dasMessage)
{
  if (nullptr != gEventProvider) {
    gEventProvider->LogError(dasMessage);
  }
}

void sLogWarning(const DasMsg& dasMessage)
{
  if (nullptr != gEventProvider) {
    gEventProvider->LogWarning(dasMessage);
  }
}

void sLogInfo(const DasMsg& dasMessage)
{
  if (nullptr != gEventProvider) {
    gEventProvider->LogInfo(dasMessage);
  }
}

void sLogDebug(const DasMsg& dasMessage)
{
  if (nullptr != gEventProvider) {
    gEventProvider->LogDebug(dasMessage);
  }
}


void sSetGlobal(const char* key, const char* value)
{
  if (nullptr == gEventProvider) {
    return;
  }
  gEventProvider->SetGlobal(key, value);
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

void sDebugBreak()
{

#if ANKI_DEVELOPER_CODE

#if defined(ANKI_PLATFORM_IOS)

  // iOS device - break to supervisor process
  // This works on a debug build, but causes an access exception (EXC_BAD_ACCESS)
  // in a release build.
  asm volatile ("svc #0");

#elif defined(ANKI_PLATFORM_OSX)

  // MacOS X - break to supervisor process
  // This works for debug or release, but causes SIGTRAP if there is no supervisor.
  // http://stackoverflow.com/questions/37299/xcode-equivalent-of-asm-int-3-debugbreak-halt
  // asm volatile ("int $3");

  // Interrupt thread with no-op signal.  This causes debugger breakpoint inside pthread_kill.
  pthread_kill(pthread_self(), SIGCONT);

#else

  // Android, Windows, linux TBD
  // Send no-op signal to cause debugger break
  pthread_kill(pthread_self(), SIGCONT);

#endif // TARGET_OS

#endif // ANKI_DEVELOPER_CODE

}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

#ifndef ALLOW_BREAK_ON_ERROR
#define ALLOW_BREAK_ON_ERROR 1
#endif

void sDebugBreakOnError()
{
  #if ALLOW_BREAK_ON_ERROR
  sDebugBreak();
  #endif
}

#undef ALLOW_BREAK_ON_ERROR

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

void sAbort()
{
  LogError("Util.Logging.Abort", {}, "Application abort");

  // Add breakpoint here to inspect application state */
  abort();

}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

void sSetErrG()
{
  // locking here is to block access during a call to sPushErrG/sPopErrG
  if (_lockErrG) {
    sErrGMutex.lock();
  }
  _errG = true;
  if (_lockErrG) {
    sErrGMutex.unlock();
  }
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

void sUnSetErrG()
{
  // locking here is to block access during a call to sPushErrG/sPopErrG
  if (_lockErrG) {
    sErrGMutex.lock();
  }
  _errG = false;
  if (_lockErrG) {
    sErrGMutex.unlock();
  }
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

bool sGetErrG()
{
  // locking here is to block access during a call to sPushErrG/sPopErrG
  if (_lockErrG) {
    sErrGMutex.lock();
  }
  const bool errG = _errG;
  if (_lockErrG) {
    sErrGMutex.unlock();
  }
  return errG;
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

void sPushErrG(bool value)
{
  if (_lockErrG) {
    sErrGMutex.lock();
  }
  sOldErrG.push_back( _errG );
  _errG = value;
}

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

void sPopErrG()
{
  DEV_ASSERT( !sOldErrG.empty(), "sPopErrG.PushWasntCalled" );
  _errG = sOldErrG.back();
  sOldErrG.pop_back();

  if (_lockErrG) {
    sErrGMutex.unlock();
  }
}

#if ANKI_BREADCRUMBS
bool DropBreadcrumb(bool result, const char* file, int line)
{
  static const int MAX_THREADS = 32; // max threads per process
  static const int BUFFER_SIZE = 16; // number of entries for file/line
  static const int LOOP_DEPTH = 3; // amount of history to check for dupe file/line

  // single statically allocated buffer for each process shared between threads

  static const char* files[BUFFER_SIZE * MAX_THREADS] = {0};
  static int lines[BUFFER_SIZE * MAX_THREADS] = {0};
  static int counts[BUFFER_SIZE * MAX_THREADS] = {0};
  static struct timeval time[BUFFER_SIZE * MAX_THREADS];

  // thread local storage, store a baseptr into statically allocated buffers above, plus
  // running round-robin offset

  // offset - 1 is the last written entry
  // offset +/- 0 is the oldest
  // offset + 1 is the next oldest

  static __thread int base = -1;
  static __thread int offset = 0;
  static __thread bool crashed = false;

  static std::atomic<int> alloc(0);

  if (base == -1) {
    // in release, keep wrapping around the internal buffer, corrupts some state but doesn't crash
    // assert in debug
    base = alloc++;
    base %= MAX_THREADS;
    base *= BUFFER_SIZE;
  }

  if (line == -1 && !crashed) {
    static std::mutex mutex;
    std::lock_guard<std::mutex> lock(mutex);
    pthread_t tid = pthread_self();

    printf("breadcrumbs for thread %p (not a stack trace)...\n", (void*)tid);
    const int oldestOffset = ((offset + 0) + BUFFER_SIZE) % BUFFER_SIZE;
    for(int i = 0; i < BUFFER_SIZE; ++i) {
      const int currentOffset = ((offset + i) + BUFFER_SIZE) % BUFFER_SIZE;
      if (files[currentOffset]) {
          const int64_t delta_sec = time[base + currentOffset].tv_sec - time[base + oldestOffset].tv_sec;
          const int64_t delta_usec = (delta_sec * 1000000) + (int64_t)(time[base + currentOffset].tv_usec - time[base + oldestOffset].tv_usec);
          printf("%d)  %s:%d cnt %d %lld usec\n", i, files[base + currentOffset], lines[base + currentOffset], counts[base + currentOffset], delta_usec);
      }
    }

    crashed = true;
  }

  if (!crashed) {
    bool loop = false;

    for(int i = 1; i <= LOOP_DEPTH; ++i) {
      // offset is one past the last entry
      const int prevOffset = ((offset - i) + BUFFER_SIZE) % BUFFER_SIZE;
      if (files[base + prevOffset] == file && lines[base + prevOffset] == line) {
        ++counts[base + prevOffset];
        loop = true;
        break;
      }
    }

    if (!loop) {
      // not in a loop
      files[base + offset] = file;
      lines[base + offset] = line;
      counts[base + offset] = 0;
      gettimeofday(&time[base + offset], NULL);

      offset = (offset + 1) % BUFFER_SIZE;
    }
  }

  return result;
}
#endif

} // namespace Util
} // namespace Anki
