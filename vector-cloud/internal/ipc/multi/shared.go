package multi

func getBufferForMessage(dest string, buf []byte) []byte {
	sendbuf := append([]byte(dest+"\x00"), buf...)
	return sendbuf
}
