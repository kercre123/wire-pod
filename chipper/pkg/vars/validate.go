package vars

import (
	"fmt"
	"regexp"
	"strings"
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
