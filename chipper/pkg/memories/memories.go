// Package memories keeps a small durable list of facts per robot, so each robot can recall
// things in later conversations. Memories are stored as <data dir>/memories/<esn>.json and are
// never shared between robots.
package memories

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

const (
	MaxPerRobot = 50
	MaxTextLen  = 200
)

var (
	ErrEmpty   = errors.New("memory text is empty")
	ErrTooLong = errors.New("memory text must be 200 characters or fewer")
	ErrBadESN  = errors.New("invalid robot serial")
)

type Memory struct {
	ID      string    `json:"id"`
	Text    string    `json:"text"`
	Created time.Time `json:"created"`
}

var (
	mu    sync.Mutex
	cache = map[string][]Memory{}
	// DirOverride is for tests; empty means <data dir>/memories
	DirOverride string
	esnRe       = regexp.MustCompile(`^[a-z0-9]{1,32}$`)
)

func dir() string {
	if DirOverride != "" {
		return DirOverride
	}
	// vars.JdocsDir is <data dir>/jdocs
	return filepath.Join(filepath.Dir(vars.JdocsDir), "memories")
}

func normalizeESN(esn string) (string, error) {
	esn = strings.ToLower(strings.TrimSpace(esn))
	if !esnRe.MatchString(esn) {
		return "", ErrBadESN
	}
	return esn, nil
}

// CleanText trims and collapses whitespace (memories are single-line).
func CleanText(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

// Truncate shortens text to MaxTextLen characters.
func Truncate(text string) string {
	r := []rune(text)
	if len(r) > MaxTextLen {
		return strings.TrimSpace(string(r[:MaxTextLen]))
	}
	return text
}

// load must be called with mu held.
func load(esn string) ([]Memory, error) {
	if list, ok := cache[esn]; ok {
		return list, nil
	}
	b, err := os.ReadFile(filepath.Join(dir(), esn+".json"))
	if err != nil {
		if os.IsNotExist(err) {
			cache[esn] = []Memory{}
			return cache[esn], nil
		}
		return nil, err
	}
	var list []Memory
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, err
	}
	if list == nil {
		list = []Memory{}
	}
	cache[esn] = list
	return list, nil
}

// save writes atomically (temp file + rename). Must be called with mu held.
func save(esn string, list []Memory) error {
	if err := os.MkdirAll(dir(), 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir(), esn+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, filepath.Join(dir(), esn+".json")); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}

// List returns a copy of the robot's memories, oldest first.
func List(esn string) ([]Memory, error) {
	esn, err := normalizeESN(esn)
	if err != nil {
		return nil, err
	}
	mu.Lock()
	defer mu.Unlock()
	list, err := load(esn)
	if err != nil {
		return nil, err
	}
	return append([]Memory{}, list...), nil
}

// Add stores a memory. A case-insensitive duplicate is not added; the existing one is returned with added=false.
// When the robot already has MaxPerRobot memories the oldest is dropped.
func Add(esn, text string) (m Memory, added bool, err error) {
	esn, err = normalizeESN(esn)
	if err != nil {
		return
	}
	text = CleanText(text)
	if text == "" {
		err = ErrEmpty
		return
	}
	if utf8.RuneCountInString(text) > MaxTextLen {
		err = ErrTooLong
		return
	}
	mu.Lock()
	defer mu.Unlock()
	list, err := load(esn)
	if err != nil {
		return
	}
	for _, existing := range list {
		if strings.EqualFold(existing.Text, text) {
			return existing, false, nil
		}
	}
	idBytes := make([]byte, 6)
	rand.Read(idBytes)
	m = Memory{ID: hex.EncodeToString(idBytes), Text: text, Created: time.Now().UTC()}
	newList := append(append([]Memory{}, list...), m)
	if len(newList) > MaxPerRobot {
		newList = newList[len(newList)-MaxPerRobot:]
	}
	if err = save(esn, newList); err != nil {
		return
	}
	cache[esn] = newList
	return m, true, nil
}

// Delete removes one memory by id; found is false if there was none.
func Delete(esn, id string) (found bool, err error) {
	esn, err = normalizeESN(esn)
	if err != nil {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	list, err := load(esn)
	if err != nil {
		return
	}
	newList := []Memory{}
	for _, m := range list {
		if m.ID == id {
			found = true
			continue
		}
		newList = append(newList, m)
	}
	if !found {
		return
	}
	if err = save(esn, newList); err != nil {
		return false, err
	}
	cache[esn] = newList
	return true, nil
}

// Clear removes all of a robot's memories.
func Clear(esn string) error {
	esn, err := normalizeESN(esn)
	if err != nil {
		return err
	}
	mu.Lock()
	defer mu.Unlock()
	// don't bother reading a possibly-corrupt file: clearing replaces it
	if err := save(esn, []Memory{}); err != nil {
		return err
	}
	cache[esn] = []Memory{}
	return nil
}

// PromptBlock returns "Things you remember:" and a bullet list for the LLM system prompt, or "" if none.
func PromptBlock(esn string) string {
	list, err := List(esn)
	if err != nil || len(list) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("Things you remember:")
	for _, m := range list {
		sb.WriteString("\n- " + m.Text)
	}
	return sb.String()
}
