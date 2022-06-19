/**
* File: tickTimeProvider
*
* Author: damjan
* Created: 4/21/2015
*
* Description: interfaces for logging class to obtain the current tick time
*
* Copyright: Anki, Inc. 2014
*
**/


#ifndef __Util_Logging_TickTimeProvider_H__
#define __Util_Logging_TickTimeProvider_H__


#import <cstdlib>

namespace Anki{
namespace Util {

class ITickTimeProvider {
public:
  virtual ~ITickTimeProvider() {};
  virtual const size_t GetTickCount() const = 0;
};

} // namespace Util
} // namespace Anki



#endif // __Util_Logging_TickTimeProvider_H__

