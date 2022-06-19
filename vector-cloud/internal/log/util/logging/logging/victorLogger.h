/**
* File: util/logging/victorLogger.h
*
* Description: Platform-independent wrapper for VictorLogger
*
* Copyright: Anki, inc. 2018
*
*/
#ifndef __util_logging_victorLogger_h
#define __util_logging_victorLogger_h

#if defined(VICOS) && VICOS
#include "victorLogger_vicos.h"
#else
#error "This class (VictorLogger) is not supported on this platform"
#endif

#endif //__util_logging_victorLogger_h
