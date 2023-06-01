/**
 * File: callstack
 *
 * Author: raul
 * Created: 06/18/2016
 *
 * Description: Functions related to debugging the callstack.
 *
 * Copyright: Anki, Inc. 2016
 *
 **/
#include "callstack.h"
#include "util/logging/logging.h"
#include "util/string/stringHelpers.h"

#include <cstdlib>

#include <string>
#include <sstream>
#include <vector>
#include <assert.h>

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
// appropriate headers per platform
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
#if defined(__MACH__) && defined(__APPLE__)
#define USE_BACKTRACE
#define USE_MACOS_DEMANGLE
#elif defined(VICOS)
#define USE_BACKTRACE
#define USE_VICOS_DEMANGLE
#elif defined(ANDROID) || defined(LINUX)

#else
#error "Unsupported platform"
#endif

#ifdef USE_BACKTRACE
#include <execinfo.h>
#include <cxxabi.h>
#endif

namespace Anki {
namespace Util {

#ifdef USE_MACOS_DEMANGLE
static std::string Demangle(const std::string& backtraceFrame) {
  std::vector<std::string> splitFrame = SplitString(backtraceFrame, ' ');

  // -4 because the valid statuses from __cxa_demangle() is 0, -1, -2, -3. If we got -4 after
  // calling __cxa_demangle() something very wrong happened.
  int status = -4;

  // Make sure that when the backtrace frame string is split
  // by spaces there is at least 3 items in there.
  if (splitFrame.size() > 3) {
    // A backtraceFrame looks like:
    // "0   webotsCtrlGameEngine                0x000000010cc84894 _ZN4Anki4Util14sDumpCallstackERKNSt3__112basic_stringIcNS1_11char_traitsIcEENS1_9allocatorIcEEEE + 52"
    // The mangled symbol is the third element from the back, when split by ' '
    //
    // Use a tempPtr here in case __cxa_demangle cannot demangle the given string and returns a nullptr.
    char* tempPtr = abi::__cxa_demangle(splitFrame.end()[-3].c_str(), 0, 0, &status);

    switch (status) {
      case 0:
        assert(tempPtr != nullptr);
        // Everything is fine, proceed to change the mangled name into the demangled version.
        splitFrame.end()[-3] = tempPtr;
        break;

      case -2:
      {
        // Demangle didn't work, don't change the name.
        PRINT_NAMED_DEBUG("Callstack.Demangle",
                          "%s is not a valid name under the C++ ABI mangling rules.",
                          splitFrame.end()[-3].c_str());
        break;
      }

      default:
      {
        // Demangle didn't work, don't change the name.
        PRINT_NAMED_WARNING("Callstack.Demangle",
                            "Couldn't demangle the symbol, __cxa_demangle returned status = %i", status);
        break;
      }
    }
    free(tempPtr);
  } else {
    PRINT_NAMED_WARNING("Callstack.Demangle",
                        "Something is wrong with the format of the backtrace frame. It should look "
                        "something like "
                        "'0   webotsCtrlGameEngine                0x000000010cc84894 _ZN4Anki4Util14sDumpCallstackERKNSt3__112basic_stringIcNS1_11char_traitsIcEENS1_9allocatorIcEEEE + 52'");
  }

  std::ostringstream os;
  for (const std::string& part : splitFrame) {
    os << part << " ";
  }

  return os.str();
}
#endif  // USE_MACOS_DEMANGLE

#ifdef USE_VICOS_DEMANGLE
static std::string Demangle(const char * frame)
{
  if (frame == nullptr) {
    // Don't crash
    return "NULL";
  }

  //
  // VicOS stack frames look like these lines:
  //   /anki/lib/libankiutil.so(_ZN4Anki4Util14sDumpCallstackERKNSt3__112basic_stringIcNS1_11char_traitsIcEENS1_9allocatorIcEEEE+0x23) [0xb03d4920]
  //   /anki/bin/vic-anim(+0x3dcc2) [0x7f592cc2]
  // or generically
  //  obj(name+offset) [addr]
  //
  // The mangled name (if any) appears between left paren and plus.
  // If we can't find a mangled name, return the frame unchanged.
  //
  std::string s = frame;

  auto pos1 = s.find('(');
  if (pos1 == std::string::npos) {
    // Can't find left paren
    return s;
  }

  auto pos2 = s.find('+', pos1+1);
  if (pos2 == std::string::npos) {
    // Can't find plus
    return s;
  }

  // Get the mangled name
  std::string name = s.substr(pos1+1, pos2-pos1-1);
  if (name.empty()) {
    // No name provided
    return s;
  }

  // Perform the demangle
  int status = 0;
  char * temp = abi::__cxa_demangle(name.c_str(), 0, 0, &status);
  if (temp == nullptr) {
    // Unable to demangle
    return s;
  }

  // Replace mangled name with demangled name
  s.replace(pos1+1, pos2-pos1-1, temp);
  free(temp);

  return s;
}
#endif

void sDumpCallstack(const std::string& name)
{
  #ifdef USE_BACKTRACE
  {
    void* callstack[128];
    int frames = backtrace(callstack, 128);
    char** strs = backtrace_symbols(callstack, frames);
    for (int i = 0; i < frames; ++i) {
      PRINT_CH_INFO("Unfiltered", name.c_str(), "%s", Demangle(strs[i]).c_str());
    }
    free(strs);
  }
  #else

  //TODO implement android

  #endif
}

} // namespace Util
} // namespace Anki
