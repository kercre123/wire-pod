// +build vicos

package main

import (
	"github.com/digital-dream-labs/vector-cloud/internal/cloudproc"
	"github.com/digital-dream-labs/vector-cloud/internal/robot"
	"github.com/digital-dream-labs/vector-cloud/internal/voice"
)

func init() {
	checkDataFunc = checkCloudDataFiles
	platformOpts = append(platformOpts, cloudproc.WithVoiceOptions(voice.WithRequireToken()))
}

func checkCloudDataFiles() error {
	esn, err := robot.ReadESN()
	if err != nil {
		return err
	}

	return robot.CheckFactoryCloudFiles(robot.DefaultCloudDir, esn)
}
