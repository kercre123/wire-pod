package conn

import "sync"

type lockState struct {
	enabled bool
	m       sync.RWMutex
}

func newLockState() *lockState {
	return &lockState{
		enabled: false,
		m:       sync.RWMutex{},
	}
}

func (l *lockState) Enabled() bool {
	l.m.RLock()
	r := l.enabled
	l.m.RUnlock()
	return r
}

func (l *lockState) Enable() {
	l.m.Lock()
	l.enabled = true
	l.m.Unlock()
}

func (l *lockState) Disable() {
	l.m.Lock()
	l.enabled = false
	l.m.Unlock()
}
