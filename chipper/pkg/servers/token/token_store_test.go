package tokenserver

import (
	"fmt"
	"sync"
	"testing"
)

// A robot sends two ReadDocs at once on first connect; both match the same
// TokenHashStore entry and remove it. That must not panic or drop other entries.
func TestPrimaryStoreConcurrentRemove(t *testing.T) {
	TokenHashStore = nil
	bot := [3]string{"192.168.1.10", "guid-a", "hash-a"}
	other := [3]string{"192.168.1.11", "guid-b", "hash-b"}
	AddToPrimaryStore(bot)
	AddToPrimaryStore(other)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for _, pair := range PrimaryStoreEntries() {
				if pair[0] == bot[0] {
					RemoveFromPrimaryStore(pair)
				}
			}
		}()
		go func(i int) {
			defer wg.Done()
			AddToPrimaryStore([3]string{fmt.Sprintf("10.0.0.%d", i), "g", "h"})
		}(i)
	}
	wg.Wait()

	entries := PrimaryStoreEntries()
	if len(entries) != 51 {
		t.Fatalf("expected 51 entries, got %d", len(entries))
	}
	for _, e := range entries {
		if e == bot {
			t.Fatal("matched entry was not removed")
		}
	}
	found := false
	for _, e := range entries {
		if e == other {
			found = true
		}
	}
	if !found {
		t.Fatal("another robot's entry was removed")
	}
}
