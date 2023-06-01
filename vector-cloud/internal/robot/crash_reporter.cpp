//
// anki/robot/crash_reporter.cpp
//
// Wrapper methods so C++ library functions can be called from go.
// Calls are stubbed out unless ANKI_PLATFORM_VICOS is defined.
//

#include "crash_reporter.h"

#ifdef ANKI_PLATFORM_VICOS
#include "platform/victorCrashReports/victorCrashReporter.h"
#endif

extern "C"
{

void InstallCrashReporter(const char* proctag)
{
#ifdef ANKI_PLATFORM_VICOS
  Anki::Vector::InstallCrashReporter(proctag);
#endif
}

void UninstallCrashReporter()
{
#ifdef ANKI_PLATFORM_VICOS
  Anki::Vector::UninstallCrashReporter();
#endif
}

} // end extern "C"
