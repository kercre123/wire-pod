// +build vicos

package util

import (
	"github.com/digital-dream-labs/vector-cloud/internal/robot"
)

func init() {
	if opt := robot.OSUserAgent(); opt != nil {
		platformOpts = append(platformOpts, opt)
	}
}
