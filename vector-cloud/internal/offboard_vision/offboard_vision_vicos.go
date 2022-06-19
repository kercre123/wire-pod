// +build !shipping,vicos

package offboard_vision

import (
	"github.com/digital-dream-labs/vector-cloud/internal/log"
	"github.com/digital-dream-labs/vector-cloud/internal/robot"
)

func init() {
	if esn, err := robot.ReadESN(); err != nil {
		log.Println("Couldn't read robot ESN:", err)
	} else {
		deviceID = esn
	}
}
