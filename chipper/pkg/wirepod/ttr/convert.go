package wirepod_ttr

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/zaf/resample"
)

func bytesToInt16s(data []byte) []int16 {
	int16s := make([]int16, len(data)/2)
	for i := range int16s {
		int16s[i] = int16(binary.LittleEndian.Uint16(data[i*2 : i*2+2]))
	}
	return int16s
}

func int16sToBytes(data []int16) []byte {
	bytes := make([]byte, len(data)*2)
	for i, val := range data {
		binary.LittleEndian.PutUint16(bytes[i*2:], uint16(val))
	}
	return bytes
}

// entry-point (go, not defaulgt)
func downsample24kTo16k(input []byte) [][]byte {
	iVolBytes := increaseVolume(input, 6)
	outBytes := downsample24kTo16kLinear(iVolBytes)
	var audioChunks [][]byte
	// the "s" sounds are harsh. put through filter
	//filteredBytes := lowPassFilter(outBytes, 4000, 16000)
	for len(outBytes) >= 1024 {
		audioChunks = append(audioChunks, outBytes[:1024])
		outBytes = outBytes[1024:]
	}

	return audioChunks
}

func increaseVolume(data []byte, factor float64) []byte {
	int16s := bytesToInt16s(data)

	for i := range int16s {
		scaled := float64(int16s[i]) * factor
		if scaled > math.MaxInt16 {
			int16s[i] = math.MaxInt16
		} else if scaled < math.MinInt16 {
			int16s[i] = math.MinInt16
		} else {
			int16s[i] = int16(scaled)
		}
	}

	return int16sToBytes(int16s)
}

// this is copied
func lowPassFilter(data []byte, cutoffFreq float64, sampleRate int) []byte {
	int16s := bytesToInt16s(data)
	filtered := make([]int16, len(int16s))
	rc := 1.0 / (2 * 3.1416 * cutoffFreq)
	dt := 1.0 / float64(sampleRate)
	alpha := dt / (rc + dt)
	filtered[0] = int16s[0]
	for i := 1; i < len(int16s); i++ {
		current := alpha*float64(int16s[i]) + (1-alpha)*float64(filtered[i-1])
		filtered[i] = int16(current)
	}

	return int16sToBytes(filtered)
}

// technically copied too
func downsample24kTo16kLinear(input []byte) []byte {
	int16s := bytesToInt16s(input)
	outputLength := (len(int16s) * 2) / 3
	output := make([]int16, outputLength)

	j := 0
	for i := 0; i < len(int16s)-2; i += 3 {
		first := (2*int32(int16s[i]) + int32(int16s[i+1])) / 3
		second := (int32(int16s[i+1]) + 2*int32(int16s[i+2])) / 3
		output[j] = int16(first)
		output[j+1] = int16(second)
		j += 2
	}

	return int16sToBytes(output)
}

// entry-point (soxr)
func downsampleAudioSoxr(input []byte) [][]byte {
	newBytes := new(bytes.Buffer)
	dec, _ := resample.New(newBytes, 24000, 16000, 1, resample.I16, resample.VeryHighQ)
	dec.Write(input)
	var audioChunks [][]byte
	decodedBytes := newBytes.Bytes()
	filteredBytes := lowPassFilter(decodedBytes, 4000, 16000)
	iVolBytes := increaseVolume(filteredBytes, 5)
	for len(iVolBytes) >= 1024 {
		audioChunks = append(audioChunks, iVolBytes[:1024])
		iVolBytes = iVolBytes[1024:]
	}
	return audioChunks
}
