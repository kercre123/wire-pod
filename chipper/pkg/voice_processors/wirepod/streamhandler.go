package wirepod

import (
	"encoding/binary"
	"fmt"

	opus "github.com/digital-dream-labs/opus-go/opus"
)

func split(buf []byte) [][]byte {
	var chunk [][]byte
	for len(buf) >= 320 {
		chunk = append(chunk, buf[:320])
		buf = buf[320:]
	}
	return chunk
}

func bytesToIntVAD(stream opus.OggStream, data []byte, die bool, isOpus bool) [][]byte {
	// detect if data is pcm or opus
	if die {
		return nil
	}
	if isOpus {
		// opus
		n, err := stream.Decode(data)
		if err != nil {
			fmt.Println(err)
		}
		byteArray := split(n)
		return byteArray
	} else {
		// pcm
		byteArray := split(data)
		return byteArray
	}
}

func bytesToSamples(buf []byte) []int16 {
	samples := make([]int16, len(buf)/2)
	for i := 0; i < len(buf)/2; i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(buf[i*2:]))
	}
	return samples
}

func bytesToIntHound(stream opus.OggStream, data []byte, die bool, isOpus bool) []byte {
	// detect if data is pcm or opus
	if die {
		return nil
	}
	return data
}
