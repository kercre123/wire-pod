package main

import "golang.org/x/time/rate"

// MultiLimiter defines a combination of rate limiters for more complex rate limiting.
// The max theoretical limit is the time of all limiters combined. But this is in the major spamming case.
// It is highly recommended to keep this logic simple.
type MultiLimiter struct {
	limiters []*rate.Limiter
}

// NewMultiLimiter creates a *MultiLimiter where the limiters are applied in the order provided.
func NewMultiLimiter(limiters ...*rate.Limiter) *MultiLimiter {
	ml := new(MultiLimiter)
	ml.limiters = limiters
	return ml
}

// Allow reports whether an event may happen at time.Now().
// Use this method if you intend to drop / skip events that exceed the rate limit.
func (ml *MultiLimiter) Allow() bool {
	for _, limiter := range ml.limiters {
		if !limiter.Allow() {
			return false
		}
	}
	return true
}
