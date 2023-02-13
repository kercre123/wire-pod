package conn

import (
	"sync"

	"github.com/currantlabs/ble"
)

func (c *Connection) scan(d ble.Advertisement) {
	if d.Connectable() {
		c.scanresults.put(d)
	}
}

const vectorservice = "fee3"

type result struct {
	name string
	addr ble.Addr
}

type scan struct {
	results      sync.Map
	mutexCounter *mCounter
}

func newScan() *scan {
	m := scan{
		results:      sync.Map{},
		mutexCounter: newMutexCounter(),
	}
	return &m
}

func (m *scan) put(value ble.Advertisement) {
	// there's a race condition in the ble code triggered when value.Services()
	// is called, and it appears to be in all libraries based on currantlabs/ble.  sigh.
	for _, s := range value.Services() {
		if s.String() == vectorservice {
			// dedup
			r := m.getresults()
			for _, v := range r {
				if value.Address().String() == v.addr.String() {
					return
				}
			}
			m.results.Store(m.mutexCounter.getCount(), value)
		}
	}
}

func (m *scan) getresults() map[int]result {
	tm := map[int]result{}
	m.results.Range(
		func(key, value interface{}) bool {
			adv := value.(ble.Advertisement)
			tm[key.(int)] = result{
				name: adv.LocalName(),
				addr: adv.Address(),
			}
			return true
		},
	)
	return tm
}

func (m *scan) getresult(id int) ble.Addr {
	v, ok := m.results.Load(id)
	if !ok {
		return nil
	}
	return v.(ble.Advertisement).Address()
}

type mCounter struct {
	count int
	m     sync.Mutex
}

func newMutexCounter() *mCounter {
	return &mCounter{
		count: 1,
		m:     sync.Mutex{},
	}
}

func (m *mCounter) getCount() int {
	m.m.Lock()
	r := m.count
	m.count++
	m.m.Unlock()
	return r
}
