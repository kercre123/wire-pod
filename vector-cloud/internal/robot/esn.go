package robot

import (
	"io/ioutil"

	"github.com/digital-dream-labs/vector-cloud/internal/log"

	"google.golang.org/grpc"
)

var osVersion string

// OSUserAgent returns a grpc.DialOption that will set a user agent with the
// string: "Victor/<os_version>", if the OS version can be obtained. Otherwise,
// nil is returned.
func OSUserAgent() grpc.DialOption {
	ver := OSVersion()
	if ver == "" {
		return nil
	}
	return grpc.WithUserAgent("Victor/" + ver)
}

var bootID string

// BootID returns the unique ID generated on each robot bootup
func BootID() string {
	if bootID == "" {
		if buf, err := ioutil.ReadFile("/proc/sys/kernel/random/boot_id"); err != nil {
			log.IfVicos("Error reading boot ID:", err)
		} else {
			bootID = string(buf)
		}
	}
	return bootID
}
