package processreqs

import (
	"fmt"
	"io"
	"time"

	sr "github.com/kercre123/chipper/pkg/wirepod/speechrequest"
	"github.com/soundhound/houndify-sdk-go"
)

func StreamAudio(sreq sr.SpeechRequest, client houndify.Client, fname, uid string) string {
	// houndify won't accept the decoded bytes for whatever reason... my guess is that the slowness is due to that
	// the server seems to decode the OPUS bytes on its own
	var err error
	rp, wp := io.Pipe()
	req := houndify.VoiceRequest{
		AudioStream: rp,
		UserID:      sreq.Device,
		RequestID:   sreq.Session,
	}
	done := make(chan bool)
	var chunk []byte
	var chunkNum int = 6
	var tempChunk []byte
	var streamNum int = 0
	var hasDoneFirstReq bool = false
	go func(wp *io.PipeWriter) {
		defer wp.Close()

		for {
			select {
			case <-done:
				return
			default:
				chunk = []byte{}
				streamNum = 0
				if hasDoneFirstReq {
					chunkNum = 4
				}
				for streamNum < chunkNum {
					tempChunk = []byte{}
					sreq, tempChunk, err = sr.GetNextStreamChunkNoD(sreq)
					chunk = append(chunk, tempChunk...)
					streamNum = streamNum + 1
					hasDoneFirstReq = true
				}

				// At the EOF, the buffer will still have bytes read into it, have to write
				// those out before breaking the loop
				if err != nil {
					fmt.Println("End of stream")
					return
				}

				// Write the amount of bytes that were read in
				wp.Write(chunk)
				time.Sleep(time.Duration(1) * time.Second)
			}
		}
	}(wp)

	// listen for partial transcript responses
	partialTranscripts := make(chan houndify.PartialTranscript)
	go func() {
		for partial := range partialTranscripts {
			if *partial.SafeToStopAudio {
				fmt.Println("Safe to stop audio recieved")
				done <- true
				return
			}
		}
	}()

	serverResponse, err := client.VoiceSearch(req, partialTranscripts)
	if err != nil {
		fmt.Println(err)
		fmt.Println(serverResponse)
	}
	return serverResponse
}
