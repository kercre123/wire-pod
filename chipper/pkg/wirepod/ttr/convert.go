package wirepod_ttr

import (
	"bytes"
	"encoding/binary"

	"github.com/zaf/resample"
)

// take out every third sample, return 1024 byte chunks
func downsampleAudio(input []byte) [][]byte {
	newBytes := new(bytes.Buffer)
	dec, _ := resample.New(newBytes, 24000, 16000, 1, resample.I16, resample.HighQ)
	dec.Write(input)
	var audioChunks [][]byte
	decodedBytes := newBytes.Bytes()
	iVolBytes, _ := IncreaseVolume(decodedBytes, 5)
	for len(iVolBytes) >= 1024 {
		audioChunks = append(audioChunks, iVolBytes[:1024])
		iVolBytes = iVolBytes[1024:]
	}
	return audioChunks
}

func IncreaseVolume(audioBytes []byte, gain float64) ([]byte, error) {
	adjustedBytes := make([]byte, len(audioBytes))
	var sample int16

	for i := 0; i < len(audioBytes); i += 2 {
		// handle peaking
		sample = int16(binary.LittleEndian.Uint16(audioBytes[i : i+2]))
		adjustedSample := float64(sample) * gain
		if adjustedSample < -32768 {
			adjustedSample = -32768
		} else if adjustedSample > 32767 {
			adjustedSample = 32767
		}
		binary.LittleEndian.PutUint16(adjustedBytes[i:i+2], uint16(int16(adjustedSample)))
	}

	return adjustedBytes, nil
}
