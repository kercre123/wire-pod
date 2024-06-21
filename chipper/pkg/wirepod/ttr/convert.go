package wirepod_ttr

// take out every third sample, return 1024 byte chunks
func downsampleAudio(input []byte) [][]byte {
	numSamples := len(input) / 2
	output := make([]byte, 0, (2*numSamples*2)/3)
	for i := 0; i < numSamples; i += 3 {
		if i+2 < numSamples {
			output = append(output, input[2*i], input[2*i+1])
			output = append(output, input[2*i+2], input[2*i+3])
		}
	}
	var chunks [][]byte
	for i := 0; i < len(output); i += 1024 {
		end := i + 1024
		if end > len(output) {
			end = len(output)
		}
		chunks = append(chunks, output[i:end])
	}
	return chunks
}
