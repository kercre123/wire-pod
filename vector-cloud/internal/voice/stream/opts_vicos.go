// +build vicos

package stream

import (
	"github.com/digital-dream-labs/vector-cloud/internal/log"
	"github.com/digital-dream-labs/vector-cloud/internal/robot"

	"github.com/digital-dream-labs/api-clients/chipper"
)

func init() {
	if esn, err := robot.ReadESN(); err != nil {
		log.Println("Couldn't read robot ESN:", err)
	} else {
		platformOpts = append(platformOpts, chipper.WithDeviceID(esn))
	}
}
