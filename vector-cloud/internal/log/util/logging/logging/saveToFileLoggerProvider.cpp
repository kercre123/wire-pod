/**
* File: saveToFileLoggerProvider
*
* Author: Lee Crippen
* Created: 3/29/2016
*
* Description: 
*
* Copyright: Anki, inc. 2016
*
*/

#include "util/logging/saveToFileLoggerProvider.h"
#include "util/logging/rollingFileLogger.h"

namespace Anki {
namespace Util {
  
  
SaveToFileLoggerProvider::SaveToFileLoggerProvider(Dispatch::Queue* queue, const std::string& baseDirectory, std::size_t maxFileSize)
: _fileLogger(new RollingFileLogger(queue, baseDirectory, RollingFileLogger::kDefaultFileExtension, maxFileSize))
{
  
}
  
SaveToFileLoggerProvider::~SaveToFileLoggerProvider() = default;
  
void SaveToFileLoggerProvider::Log(ILoggerProvider::LogLevel logLevel, const std::string& message)
{
  _fileLogger->Write(message);
}
  
void SaveToFileLoggerProvider::Flush()
{
  _fileLogger->Flush();
}
  
} // end namespace Util
} // end namespace Anki
