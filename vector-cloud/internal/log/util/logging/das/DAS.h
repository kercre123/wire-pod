/**
 * File: util/logging/DAS.h
 *
 * Description: DAS extensions for Util::Logging macros
 *
 * Copyright: Anki, Inc. 2018
 *
 **/


#ifndef __util_logging_DAS_h
#define __util_logging_DAS_h

#include <string>

//
// This file must be #included by each CPP file that uses DASMSG macros, to
// ensure that macros are expanded correctly by doxygen.  If this file
// is included through an intermediate header, code will compile correctly
// but doxygen will not expand macros as expected.
//
#if !defined(__INCLUDE_LEVEL__)
#error "This header may not be included by other headers. We rely on GCC __INCLUDE_LEVEL__ to enforce this restriction."
#error "If your compiler does not define __INCLUDE_LEVEL__, please add an appropriate check for your compiler."
#endif

#if (__INCLUDE_LEVEL__ > 1)
#error "This header must be included directly from your CPP file. This header may not be included by other headers."
#error "If this header is included by another header, doxygen macros will not expand correctly."
#endif

namespace Anki {
namespace Util {
namespace DAS {

//
// DAS v2 event fields
// https://ankiinc.atlassian.net/wiki/spaces/SAI/pages/221151429/DAS+Event+Fields+for+Victor+Robot
//
constexpr const char * SOURCE = "$source";
constexpr const char * EVENT = "$event";
constexpr const char * TS = "$ts";
constexpr const char * SEQ = "$seq";
constexpr const char * LEVEL = "$level";
constexpr const char * PROFILE = "$profile";
constexpr const char * ROBOT = "$robot";
constexpr const char * ROBOT_VERSION = "$robot_version";
constexpr const char * FEATURE_TYPE = "$feature_type";
constexpr const char * FEATURE_RUN = "$feature_run";
constexpr const char * STR1 = "$s1";
constexpr const char * STR2 = "$s2";
constexpr const char * STR3 = "$s3";
constexpr const char * STR4 = "$s4";
constexpr const char * INT1 = "$i1";
constexpr const char * INT2 = "$i2";
constexpr const char * INT3 = "$i3";
constexpr const char * INT4 = "$i4";

//
// DAS event marker
//
// This is the character chosen to precede a DAS log event on Victor.
// This is used as a hint that a log message should be parsed as an event.
// If this value changes, event readers and event writers must be updated.
//
constexpr const char EVENT_MARKER = '@';

//
// DAS field marker
//
// This is the character chosen to separate DAS log event fields on Victor.
// This character may not appear in event data.
// If this value changes, event readers and event writers must be updated.
//
constexpr const char FIELD_MARKER = '\x1F';

//
// DAS field count
// This is the number of fields used to represent a DAS log event on Victor.
// If this value changes, event readers and event writers must be updated.
//
constexpr const int FIELD_COUNT = 9;

//
// Return DAS uptime value, aka millisecs since boot
//
uint64_t UptimeMS();

//
// Return DAS encoded string, safe for use with JSON
//
std::string Escape(const char * str);
std::string Escape(const std::string & str);

} // end namespace DAS
} // end namespace Util
} // end namespace Anki

namespace Anki {
namespace Util {

//
// DAS message field
//
struct DasItem
{
  DasItem() = default;
  DasItem(const std::string & valueStr) { value = valueStr; }
  DasItem(int64_t valueInt) { value = std::to_string(valueInt); }

  inline const std::string & str() const { return value; }
  inline const char * c_str() const { return value.c_str(); }

  std::string value;
};

//
// DAS message struct
//
// Event name is required. Other fields are optional.
// Event structures should be declared with DASMSG.
// Event fields should be assigned with DASMSG_SET.
//
struct DasMsg
{
  DasMsg(const std::string & eventStr) { event = eventStr; }

  std::string event;
  DasItem s1;
  DasItem s2;
  DasItem s3;
  DasItem s4;
  DasItem i1;
  DasItem i2;
  DasItem i3;
  DasItem i4;
};

// Log an error event
__attribute__((__used__))
void sLogError(const DasMsg & dasMessage);

// Log a warning event
__attribute__((__used__))
void sLogWarning(const DasMsg & dasMessage);

// Log an info event
__attribute__((__used__))
void sLogInfo(const DasMsg & dasMessage);

// Log a debug event
__attribute__((__used__))
void sLogDebug(const DasMsg & dasMessage);

} // end namespace Util
} // end namespace Anki

//
// DAS feature start event
// This is the name of the event used to indicate start a new feature scope.
// If this value changes, or event parameters change, DASManager must be updated to match.
// This is declared as a macro, not a constexpr, so it can be expanded by doxygen.
//
#define DASMSG_FEATURE_START "behavior.feature.start"

//
// DAS BLE connection events
// These events are used to indicate start/stop of a BLE connection.
// If event name or event parameters change, DASManager must be updated to match.
// Names are defined as a macro, not a constexpr, so they can be expanded by doxygen.
//
#define DASMSG_BLE_CONN_ID_START "ble_conn_id.start"
#define DASMSG_BLE_CONN_ID_STOP "ble_conn_id.stop"

//
// DAS WIFI connection events
// These events are used to indicate start/stop of a Wi-Fi connection.
// If event name or event parameters change, DASManager must be updated to match.
// Names are defined as a macro, not a constexpr, so they can be expanded by doxygen.
//
#define DASMSG_WIFI_CONN_ID_START "wifi_conn_id.start"
#define DASMSG_WIFI_CONN_ID_STOP "wifi_conn_id.stop"

//
// DAS Profile ID events
// These events are used to indicate start/stop of association with an Anki profile ID.
// If event name or event parameters change, DASManager must be updated to match.
// Names are defined as a macro, not a constexpr, so they can be expanded by doxygen.
//
#define DASMSG_PROFILE_ID_START "profile_id.start"
#define DASMSG_PROFILE_ID_STOP "profile_id.stop"

//
// DAS Allow Upload
// This event is used to enable/disable DAS uploads.
// If event name or event parameters change, DASManager must be updated to match.
// Names are defined as a macro, not a constexpr, so they can be expanded by doxygen.
//
#define DASMSG_DAS_ALLOW_UPLOAD "das.allow_upload"

//
// DAS message macros
//
// These macros are used to ensure that developers provide some description of each event defined.
// Note these macros are expanded two ways!  In normal compilation, they are expanded into C++
// variable declarations and logging calls.  When processed by doxygen, they are are expanded
// into syntactically invalid C++ that contains magic directives to produce readable documentation.
//
// Overwriting fields that have already been set (e.g. calling DASMSG_SET(i1, ...) twice) is not allowed.
//
// If a string field can contain JSON syntax characters such as double quotes (") or backslash (\), it should be
// escaped with DASMSG_ESCAPE(). Most string fields are JSON-safe so we do not do this automatically.
//
#ifndef DOXYGEN

#define DASMSG(ezRef, eventName, documentation) { Anki::Util::DasMsg __DAS_msg(eventName);
#define DASMSG_SET(dasEntry, value, comment) __DAS_msg.dasEntry = Anki::Util::DasItem(value); \
                                             /* define an empty struct (to detect if this item has been set twice) */ \
                                             struct _already_set_##dasEntry {};
#define DASMSG_SEND()         Anki::Util::sLogInfo(__DAS_msg); }
#define DASMSG_SEND_WARNING() Anki::Util::sLogWarning(__DAS_msg); }
#define DASMSG_SEND_ERROR()   Anki::Util::sLogError(__DAS_msg); }
#define DASMSG_SEND_DEBUG()   Anki::Util::sLogDebug(__DAS_msg); }

#else

/*! \defgroup dasmsg DAS Messages
*/

class DasDoxMsg() {}

#define DASMSG(ezRef, eventName, documentation)  }}}}}}}}/** \ingroup dasmsg */ \
                                            /** \brief eventName */ \
                                            /** documentation */ \
                                            class ezRef(): public DasDoxMsg() { \
                                            public:
#define DASMSG_SET(dasEntry, value, comment) /** @param dasEntry comment \n*/
#define DASMSG_SEND };
#define DASMSG_SEND_WARNING };
#define DASMSG_SEND_ERROR };
#define DASMSG_SEND_DEBUG };

#endif

#define DASMSG_ESCAPE(str) Anki::Util::DAS::Escape(str)

#endif // __util_logging_DAS_h
