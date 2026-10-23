package wyoming

import "encoding/binary"

// ToMono16 converts raw PCM to 16-bit little-endian mono at targetRate,
// which is what Vector's speaker stream expects. Supports 8/16/32-bit input
// with any channel count; channels are averaged, resampling is linear
// interpolation (good enough for speech).
func ToMono16(pcm []byte, format AudioFormat, targetRate int) []byte {
	if format.Width <= 0 || format.Channels <= 0 || format.Rate <= 0 {
		return nil
	}
	frameSize := format.Width * format.Channels
	frames := len(pcm) / frameSize
	mono := make([]float64, frames)
	for i := 0; i < frames; i++ {
		var sum float64
		for ch := 0; ch < format.Channels; ch++ {
			off := i*frameSize + ch*format.Width
			sum += sampleAt(pcm[off:off+format.Width], format.Width)
		}
		mono[i] = sum / float64(format.Channels)
	}

	out := mono
	if format.Rate != targetRate && frames > 1 {
		outLen := int(int64(frames) * int64(targetRate) / int64(format.Rate))
		out = make([]float64, outLen)
		step := float64(format.Rate) / float64(targetRate)
		for i := range out {
			pos := float64(i) * step
			j := int(pos)
			if j >= frames-1 {
				out[i] = mono[frames-1]
				continue
			}
			frac := pos - float64(j)
			out[i] = mono[j]*(1-frac) + mono[j+1]*frac
		}
	}

	result := make([]byte, len(out)*2)
	for i, v := range out {
		if v > 32767 {
			v = 32767
		} else if v < -32768 {
			v = -32768
		}
		binary.LittleEndian.PutUint16(result[i*2:], uint16(int16(v)))
	}
	return result
}

// sampleAt returns one sample scaled to the 16-bit range.
func sampleAt(b []byte, width int) float64 {
	switch width {
	case 1: // unsigned 8-bit
		return (float64(b[0]) - 128) * 256
	case 2:
		return float64(int16(binary.LittleEndian.Uint16(b)))
	case 4:
		return float64(int32(binary.LittleEndian.Uint32(b))) / 65536
	}
	return 0
}
