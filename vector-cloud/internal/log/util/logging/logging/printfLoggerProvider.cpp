/**
* File: printfLoggerProvider
*
* Author: damjan stulic
* Created: 4/25/15
*
* Description: 
*
* Copyright: Anki, inc. 2015
*
*/

#include "util/logging/printfLoggerProvider.h"

#include "util/color/color.h"

namespace Anki {
namespace Util {


PrintfLoggerProvider::PrintfLoggerProvider()
: _minToStderrLevel(1000)
{
  SetMinLogLevel(LOG_LEVEL_INFO);
}

PrintfLoggerProvider::PrintfLoggerProvider(ILoggerProvider::LogLevel minToStderrLogLevel,
                                           bool colorizeStderrOutput)
: _minToStderrLevel(minToStderrLogLevel)
, _colorizeStderrOutput(colorizeStderrOutput)
{
  SetMinLogLevel(LOG_LEVEL_INFO);
}

void PrintfLoggerProvider::Log(ILoggerProvider::LogLevel logLevel, const std::string& message)
{
  if (message.empty()) {
    return;
  }
  
  const bool outputToStderr = (logLevel >= _minToStderrLevel);
  
  FILE* outStream = outputToStderr ? stderr : stdout;
  
  const bool colorizeOutput = outputToStderr && _colorizeStderrOutput;
  
  if (colorizeOutput) {
    // Prepend/append color escape codes. Need to move the newline to the end if it exists (due to Webots behavior)
    fprintf(outStream,
            ANSI_COLOR_RED "%s" ANSI_COLOR_RESET "\n",
            (message.back() == '\n') ? message.substr(0, message.length() - 1).c_str() : message.c_str());
  } else {
    fprintf(outStream, "%s", message.c_str());
  }
}

void PrintfLoggerProvider::Flush()
{
  fflush(stderr);
  fflush(stdout);
}

} // end namespace Util
} // end namespace Anki
