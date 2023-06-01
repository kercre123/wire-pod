// Package time provides a wrapper for the Go time package that makes
// time based functions easier to test, while retaining the same api
// interface.
//
// eg.
//
//   var myfuncTime = testtime.New()
//
//   func myfunc() {
//     fmt.Println(myfuncTime.Now().Format(time.RFC3339))
//   }
//
// myfuncTime.Now() will pass directly through to time.Now for production code, but
// unit tests can futz with time by calling
//
//   init() {
//      myfuncTime = testtime.NewTestable()
//   }
//   func TestMyFunc(t *testing.T) {
//     myfuncTime = myfuncTime.(timefunc.TestableTime)
//     now := time.Date(2018, 9, 1, 12, 0, 0, 0, time.UTC)
//     myfuncTime.WithStaticNow(now, func() {
//       myfunc() // will alwys print 2018-09-01T12:00:00Z
//     })
//   }
package time

import (
	"sync/atomic"
	"time"
)

// Time defines some time package functions that are useful to  stub out
// during testing.
type Time interface {
	// Now implements time.Now
	Now() time.Time
}

// New returns a passthrough implementation of TimePkg.  The methods
// defined here literally call the time package functions; this should be
// the default for code that's not being manipulated by a unit test.
func New() Time {
	return new(passthru)
}

// NewTestable returns an implementation of TimePkg that allows for
// unit tests to safely manipulate time while the test is running.
func NewTestable() TestableTime {
	return newProtected()
}

// TestableTime is a superset of TimePkg, and includes methods for
// manipulating time returned by Now() etc.
//
// Methods defined within this interface are goroutine safe.
type TestableTime interface {
	Time

	// WithNowDelta changes the response of the Now call by the supplied delta.
	// duration while f is running.
	//
	// It restores the previous behviour when complete.  May be nested.
	WithNowDelta(timeDelta time.Duration, f func())

	// WithStaticNow changes the response of the Now call to return
	// a static time while f is running.
	//
	// It restores the previous behavior when complete.  May be nested.
	WithStaticNow(t time.Time, f func())

	// WithCustomNow changes the response of the Now call to run
	// the supplied tf function. oldNow will contain the previous value
	// of Now, which will usually be time.Now unless there are nested calls.
	//
	// It restores the previous behavior when complete.  May be nested.
	WithCustomNow(tf func(oldNow NowFunc) time.Time, f func())
}

type passthru struct{}

// Now passes through directly to time.Now
func (pt *passthru) Now() time.Time {
	return time.Now()
}

var _ TestableTime = new(protected)

type protected struct {
	nowFunc atomic.Value
}

func newProtected() *protected {
	p := new(protected)
	p.nowFunc.Store(time.Now)
	return p
}

func (p *protected) Now() time.Time {
	return p.nowFunc.Load().(func() time.Time)()
}

func (p *protected) WithNowDelta(timeDelta time.Duration, f func()) {
	p.WithCustomNow(func(oldNow NowFunc) time.Time {
		return oldNow().Add(timeDelta)
	}, f)
}

func (p *protected) WithStaticNow(t time.Time, f func()) {
	p.WithCustomNow(func(oldNow NowFunc) time.Time {
		return t
	}, f)
}

// NowFunc is the function signature for time.Now.
type NowFunc func() time.Time

func (p *protected) WithCustomNow(now func(oldNow NowFunc) time.Time, f func()) {
	org := p.nowFunc.Load().(func() time.Time)
	defer func(org func() time.Time) { p.nowFunc.Store(org) }(org)
	p.nowFunc.Store(func() time.Time {
		return now(org)
	})
	f()
}
