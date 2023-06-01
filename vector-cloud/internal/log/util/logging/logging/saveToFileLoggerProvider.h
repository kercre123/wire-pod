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
#ifndef __Util_Logging_SaveToFileLoggerProvider_H_
#define __Util_Logging_SaveToFileLoggerProvider_H_

#include "util/logging/iFormattedLoggerProvider.h"
#include "util/dispatchQueue/dispatchQueue.h"

#include <cstdlib>
#include <string>

namespace Anki {
namespace Util {
  
// Forward declarations
class RollingFileLogger;

class SaveToFileLoggerProvider : public IFormattedLoggerProvider {
public:
  static constexpr std::size_t kDefaultMaxFileSize = 1024 * 1024 * 20;
  
  SaveToFileLoggerProvider(Dispatch::Queue* queue, const std::string& baseDirectory, std::size_t maxFileSize = kDefaultMaxFileSize);
  virtual ~SaveToFileLoggerProvider();
  
  void Log(ILoggerProvider::LogLevel logLevel, const std::string& message) override;
  
  void Flush() override;
  
protected:
  
  // Don't want this to be copyable
  SaveToFileLoggerProvider( const SaveToFileLoggerProvider& ) = delete;
  SaveToFileLoggerProvider& operator=( const SaveToFileLoggerProvider& ) = delete;
  
  std::unique_ptr<RollingFileLogger>    _fileLogger;
};

} // end namespace Util
} // end namespace Anki


#endif //__Util_Logging_SaveToFileLoggerProvider_H_
