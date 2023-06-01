/**
 * File: iChannelFilter
 *
 * Author: raul
 * Created: 06/30/16
 *
 * Description: Interface for channel filters that we can apply to loggers.
 *
 * Copyright: Anki, inc. 2015
 *
 */

#ifndef __Util_Logging_iChannelFilter_H__
#define __Util_Logging_iChannelFilter_H__

#include <string>

namespace Anki {
namespace Util {

class IChannelFilter
{
public:

  // destructor
  virtual ~IChannelFilter() {}

  // returns true if the given channel is enabled, false otherwise
  virtual bool IsChannelEnabled(const std::string& channelName) const  = 0;
};

} // namespace Util
} // namespace Anki

#endif
