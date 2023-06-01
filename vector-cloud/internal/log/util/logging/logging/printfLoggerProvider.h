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
#ifndef __Util_Logging_PrintfLoggerProvider_H_
#define __Util_Logging_PrintfLoggerProvider_H_
#include "util/logging/iFormattedLoggerProvider.h"

namespace Anki {
namespace Util {

class PrintfLoggerProvider : public IFormattedLoggerProvider {

public:
  PrintfLoggerProvider();
  PrintfLoggerProvider(ILoggerProvider::LogLevel minToStderrLogLevel,
                       bool colorizeStderrOutput = false);
  
  void Log(ILoggerProvider::LogLevel logLevel, const std::string& message) override;
  
  // Set the minimum level required to print to stderr.
  // Everything above prints to stdout.
  void SetMinToStderrLevel(int level) { _minToStderrLevel = level; };
  
  void SetColorizeStderrOutput(bool b = true) { _colorizeStderrOutput = b; };

  void Flush() override;
  
private:
  
  int _minToStderrLevel;
  
  // Use ANSI escape color codes to colorize printf output to stderr
  bool _colorizeStderrOutput = false;
};

} // end namespace Util
} // end namespace Anki


#endif //__Util_Logging_PrintfLoggerProvider_H_
