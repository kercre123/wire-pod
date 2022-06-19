package log

/*
#include "das.h"
#include <stdint.h>
#include <time.h>

#if defined(CLOCK_BOOTTIME)
#define CLOCK CLOCK_BOOTTIME
#else
#define CLOCK CLOCK_MONOTONIC
#endif

uint64_t dasUptimeMS()
{
  struct timespec ts;
  if (clock_gettime(CLOCK, &ts) == 0) {
    return (ts.tv_sec * 1000) + (ts.tv_nsec / 1000000);
  }
  return 0;
}

*/
import "C"
import (
	"strconv"
	"strings"
)

var dasEventMarker = C.GoStringN(&C.dasEventMarker, 1)
var dasFieldMarker = C.GoStringN(&C.dasFieldMarker, 1)
var dasFieldCount = int(C.dasFieldCount)

// DasFields contains the data that can be associated with an event. An event can store...
// - up to 4 strings in the Strings member
// - up to 4 integers in the Ints member
// There is no requirement that these be filled, and each event can define the meaning of these
// data fields as it wishes.
type DasFields struct {
	Strings [4]string
	Ints    [4]string
}

// SetStrings fills the Strings member of this DasFields object with the supplied
// strings. Only up to 4 strings are copied.
func (f *DasFields) SetStrings(strs ...string) *DasFields {
	for i := range strs {
		if i >= len(f.Strings) {
			break
		}
		f.Strings[i] = strs[i]
	}
	return f
}

// SetInts fills the Ints member of this DasFields object with the supplied
// integers. Only up to 4 ints are copied.
func (f *DasFields) SetInts(ints ...int) *DasFields {
	for i := range ints {
		if i >= len(f.Ints) {
			break
		}
		f.Ints[i] = strconv.FormatInt(int64(ints[i]), 10)
	}
	return f
}

// DasEmpty logs a DAS event with only the given event name, and no extra data.
func DasEmpty(event string) {
	Das(event, &DasFields{})
}

// Das logs a DAS event with the given event name and extra data fields.
func Das(event string, fields *DasFields) {
	strs := strings.Join(fields.Strings[:], dasFieldMarker)
	ints := strings.Join(fields.Ints[:], dasFieldMarker)
	uptimeMS := strconv.FormatUint(uint64(C.dasUptimeMS()), 10)
	allFields := strings.Join([]string{event, strs, ints, uptimeMS}, dasFieldMarker)
	Printf("%s%s", dasEventMarker, allFields)
}
