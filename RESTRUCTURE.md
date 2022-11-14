# restructuring wire-pod

-	The current voice processor implementation needs to be made more simple so features (such as different STT services, weather APIs, knowledge graph APIs) can be implemented easily.

## Areas of complexity
-	Multiple STT service implementations which need to be seperated for as wide as possible platform support
-	A request can be one of three types which each need to be handled differently

## Plan

-	Create speechrequest.go which contains a new `SpeechRequest` type and functions which deal with this new type.
	-	`SpeechRequest{Device: string, Session: string, Stream: interface{}, FirstReq []byte}`
	-	`reqToSpeechRequest` will convert any type of chipperpb req to a SpeechRequest
	-	`getNextAudioChunk` will return the next audio chunk in the stream

-	An sttHandler function should look something like this (not accurate to VOSK, just an example):
```
func voskSttHandler(req SpeechRequest) (transcribedText string, err error) {
	var transcribedText string
	vosk.Process(req.FirstReq)
	//req.FirstReq is a []byte of the first audio bytes in the stream
	for {
		// getNextAudioChunk should handle stream errors on its own
		isDone, err := vosk.Process(getNextAudioChunk)
		if err != nil {
			return "", err
		}
		if isDone {
			transcribedText, err = vosk.Flush()
			if err != nil {
				return "", err
			}
			break
		}
	}
	return transcribedText, nil
}
```
