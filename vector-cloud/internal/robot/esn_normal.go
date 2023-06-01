// +build !vicos

package robot

import (
	"os/exec"
	"strings"
)

// ReadESN returns the ESN of a robot (in this case a fake one)
func ReadESN() (string, error) {
	return "00000000", nil
}

// OSVersion returns a string representation of the OS version, like:
// 4.15.0-33-generic.x86_64.GNU/Linux
func OSVersion() string {
	if osVersion != "" {
		return osVersion
	}
	runCmd := func(cmd string, params ...string) string {
		buf, err := exec.Command(cmd, params...).Output()
		if err != nil {
			return ""
		}
		return strings.Replace(strings.TrimSpace(string(buf)), " ", ".", -1)
	}

	propArgs := []string{"-oir"}

	for _, arg := range propArgs {
		if str := runCmd("uname", arg); str != "" {
			osVersion = str
			return osVersion
		}
	}
	return ""
}

func AnkiVersion() string {
	return ""
}

func VictorVersion() string {
	return ""
}
