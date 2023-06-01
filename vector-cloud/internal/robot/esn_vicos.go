// +build vicos

package robot

import (
	"os/exec"
	"strings"
)

var esn string

// ReadESN returns the ESN of a robot by reading it from
// the robot's filesystem
func ReadESN() (string, error) {
	if esn != "" {
		return esn, nil
	}

	buf, err := exec.Command("/bin/emr-cat", "e").Output()
	if err != nil {
		return "", err
	}
	esn = strings.TrimSpace(string(buf))
	return esn, nil
}

// OSVersion returns a string representation of the OS version, like:
// v0.10.1252d_os0.10.1252d-79470cd-201806271633
func OSVersion() string {
	if osVersion != "" {
		return osVersion
	}
	runCmd := func(cmd string, params ...string) string {
		buf, err := exec.Command(cmd, params...).Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(buf))
	}

	propArgs := []string{"ro.build.fingerprint", "ro.revision",
		"ro.anki.os_build_comment", "ro.build.version.release"}

	for _, arg := range propArgs {
		if str := runCmd("getprop", arg); str != "" {
			osVersion = str
			return osVersion
		}
	}
	return ""
}

var ankiVersion string

func AnkiVersion() string {
	if ankiVersion != "" {
		return ankiVersion
	}

	buf, err := exec.Command("getprop", "ro.anki.version").Output()
	if err != nil {
		return ""
	}
	ankiVersion = strings.TrimSpace(string(buf))
	return ankiVersion
}

var victorVersion string

func VictorVersion() string {
	if victorVersion != "" {
		return victorVersion
	}

	buf, err := exec.Command("getprop", "ro.anki.victor.version").Output()
	if err != nil {
		return ""
	}
	victorVersion = strings.TrimSpace(string(buf))
	return victorVersion
}
