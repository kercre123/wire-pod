package vars

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var esnRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}$`)
var robotNameRegex = regexp.MustCompile(`^[A-Za-z]+-[A-Za-z0-9]+$`)

func ValidESN(esn string) bool {
	return esnRegex.MatchString(esn)
}

func ValidRobotName(name string) bool {
	return robotNameRegex.MatchString(name)
}

func ESNFromThing(thing string) (string, error) {
	esn := strings.TrimPrefix(thing, "vic:")
	if esn == thing || !ValidESN(esn) {
		return "", fmt.Errorf("invalid thing: %q", thing)
	}
	return esn, nil
}

const MaxDisplayNameLen = 32

// ValidDisplayName checks a user-chosen display name: 1-32 characters, no control characters.
func ValidDisplayName(name string) bool {
	if name == "" || utf8.RuneCountInString(name) > MaxDisplayNameLen {
		return false
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}