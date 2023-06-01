package util

import (
	"io"
	"sync"
	"time"
)

// DoOnce wraps a sync.Once object, taking a function at initialization time
// rather than at execution time; that way, if the function to be executed is
// known in advance for all cases, it can be stored in the object rather than
// repeated in arguments to Once.Do()
type DoOnce struct {
	once sync.Once
	todo func()
}

// NewDoOnce returns a new DoOnce object that will store the given function and
// execute it in calls to DoOnce.Do()
func NewDoOnce(todo func()) DoOnce {
	return DoOnce{sync.Once{}, todo}
}

// Do executes the function associated with this DoOnce instance
func (d *DoOnce) Do() {
	d.once.Do(d.todo)
}

// CanSelect is a helper for struct{} channels that checks if the channel can be
// pulled from in a select statement. It is recommended to use this only for channels
// whose purpose is signaling (i.e. closing the given channel when something is done,
// at which point it can always be selected, rather than transmitting actual struct{}
// values), since using this with actual struct{} values would cause a value to get
// pulled off the channel and potentially mess up synchronization.
func CanSelect(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

// TimeFuncMs returns the time, in milliseconds, required to run the given function
func TimeFuncMs(function func()) float64 {
	callStart := time.Now()
	function()
	return float64(time.Now().Sub(callStart).Nanoseconds()) / float64(time.Millisecond/time.Nanosecond)
}

type chanWriter struct {
	ch chan<- []byte
}

func (c chanWriter) Write(p []byte) (int, error) {
	c.ch <- p
	return len(p), nil
}

// NewChanWriter returns a wrapper around a []byte channel that
// turns it into an io.Writer
func NewChanWriter(ch chan<- []byte) io.Writer {
	return chanWriter{ch}
}

// SleepSelect is like calling time.Sleep() with an early exit if the given
// channel ch is closed. It can be used for situations such as a goroutine
// wanting to sleep while still responding quickly if it receives a signal
// from a channel. Returns true if sleep was ended early due to selecting
// from the channel, false if the sleep completed.
func SleepSelect(dur time.Duration, ch <-chan struct{}) bool {
	timer := time.NewTimer(dur)
	select {
	case <-timer.C:
		return false
	case <-ch:
		if !timer.Stop() {
			<-timer.C
		}
		return true
	}
}

// ErrorListener defines an interface that can be used as a common definition to inject error
// handlers into dependent modules
type ErrorListener interface {
	OnError(error)
}
