/**
 * File: ChannelFilter
 *
 * Author: Zhetao Wang
 * Created: 05/19/2015
 *
 * Description: Class that provides a channel filter (for logs) that can load from json and provides console vars
 * to enable/disable log channels at runtime.
 *
 * Copyright: Anki, inc. 2015
 *
 */

#ifndef __Util_Logging_ChannelFilter_H__
#define __Util_Logging_ChannelFilter_H__

#include "iChannelFilter.h"
#include "util/console/consoleInterface.h"
#include "util/helpers/noncopyable.h"
#include "json/json.h"

#include <string>
#include <unordered_map>

namespace Anki {
namespace Util {

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
// ChannelVar: console var
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
struct ChannelVar : public ConsoleVar<bool> {
  ChannelVar(const std::string& name, bool defaultEnable, bool unregisterInDestructor)
  : ConsoleVar<bool>(enable, name.c_str(), "Channels", unregisterInDestructor)
  , enable(defaultEnable) {
    // We must re-set the default value here, after the call to ConsoleVar constructor,
    // because that constructor uses a reference to "enable", but enable isn't set until
    // after that constructor call.  VIC-13609
    this->_defaultValue = enable;
  }

  virtual void ToggleValue() override {
    ConsoleVar<bool>::ToggleValue();

    // report back to channel filter if required
  }

  bool enable; // variable where the value is stored
};

// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
// ChannelFilter
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
class ChannelFilter : public IChannelFilter, Anki::Util::noncopyable {
public:
  ChannelFilter() : _initialized(false){}
  ~ChannelFilter();
  
  // initialize with an optional json configuration (can be empty)
  void Initialize(const Json::Value& config = Json::Value());
  inline bool IsInitialized() const{ return _initialized; }
  
  void EnableChannel(const std::string& channelName);
  void DisableChannel(const std::string& channelName);

  // IChannelFilter API
  virtual bool IsChannelEnabled(const std::string& channelName) const override;
  
private:
  std::unordered_map<std::string, ChannelVar*> _channelEnableList;
  bool _initialized;
};

} // namespace Util
} // namespace Anki

#endif /* defined(__Util_Logging_ChannelFilter_H__) */
