package vars

import (
	"encoding/json"
	"errors"
	"slices"
	"sync"
)

var (
	customIntentsMu     sync.RWMutex
	customIntents       []CustomIntent
	customIntentsExists bool
)

var ErrNoSuchIntent = errors.New("invalid intent number")

func GetCustomIntents() []CustomIntent {
	customIntentsMu.RLock()
	defer customIntentsMu.RUnlock()
	return slices.Clone(customIntents)
}

func CustomIntentsExist() bool {
	customIntentsMu.RLock()
	defer customIntentsMu.RUnlock()
	return customIntentsExists
}

func saveCustomIntentsLocked() error {
	b, err := json.Marshal(customIntents)
	if err != nil {
		return err
	}
	return WriteFileAtomic(CustomIntentsPath, b, 0600)
}

func AddCustomIntent(intent CustomIntent) error {
	customIntentsMu.Lock()
	defer customIntentsMu.Unlock()
	customIntents = append(customIntents, intent)
	customIntentsExists = true
	return saveCustomIntentsLocked()
}

func EditCustomIntent(number int, fn func(intent *CustomIntent) error) error {
	customIntentsMu.Lock()
	defer customIntentsMu.Unlock()
	if number < 1 || number > len(customIntents) {
		return ErrNoSuchIntent
	}
	edited := customIntents[number-1]
	if err := fn(&edited); err != nil {
		return err
	}
	customIntents[number-1] = edited
	return saveCustomIntentsLocked()
}

func RemoveCustomIntent(number int) error {
	customIntentsMu.Lock()
	defer customIntentsMu.Unlock()
	if number < 1 || number > len(customIntents) {
		return ErrNoSuchIntent
	}
	customIntents = slices.Delete(customIntents, number-1, number)
	return saveCustomIntentsLocked()
}
