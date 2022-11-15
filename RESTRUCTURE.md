# restructuring wire-pod

-	The current voice processor implementation needs to be made more simple so features (such as different STT services, weather APIs, knowledge graph APIs) can be implemented easily.

## Areas of complexity
-	Multiple STT service implementations which need to be seperated for as wide as possible platform support
-	A request can be one of three types which each need to be handled differently

## Plan

-	Create speechrequest.go which contains a new `SpeechRequest` type and functions which deal with this new type.
	-	The functions will deal with opus to pcm conversion on their own
	-	`func reqToSpeechRequest(req interface{}) SpeechRequest` will convert any type of chipperpb req to a SpeechRequest
	-	`func getNextAudioChunk(req SpeechRequest) (SpeechRequest, []byte, err)` will return the next audio chunk in the stream
		-	req must be set equal to the output of this function
-	Each STT service should only need just one go file and minimal modification to server.go
-	Maybe work on weather too, seperate weatherAPI and openweathermap services into different functions and create a standard for future functions
-	Make knowledgegraph its own entity, remove from sttHandler functions
-	Make VAD a function
-	An sttHandler function should look something like this (not accurate to VOSK at all, just an example):
```
func voskSttHandler(req SpeechRequest) (string, error) {
	var transcribedText string
	vosk.Process(req.FirstReq)
	//req.FirstReq is a []byte of the first audio bytes in the stream
	for {
		var chunk []byte
		req, chunk, err = getNextAudioChunk(req)
		if err != nil {
			return "", err
		isDone, _ := vosk.Process(chunk)
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

## TODO

-	Overhaul weather, create functon standard
-	Make logging more consistent
-	More code comments
-	Create documentation and more examples for the function standards
-	Implement in wire-prod-pod