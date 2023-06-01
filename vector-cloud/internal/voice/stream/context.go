package stream

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/digital-dream-labs/vector-cloud/internal/clad/cloud"

	"github.com/digital-dream-labs/vector-cloud/internal/log"
	"github.com/digital-dream-labs/vector-cloud/internal/util"

	"github.com/digital-dream-labs/api-clients/chipper"
)

func (strm *Streamer) sendAudio(samples []byte) error {
	var err error
	sendTime := util.TimeFuncMs(func() {
		err = strm.conn.SendAudio(samples)
	})
	if err != nil {
		log.Println("Cloud error:", err)
		strm.respOnce.Do(func() {
			strm.receiver.OnError(errorReason(err), err)
		})
		return err
	}
	logVerbose("Sent", len(samples), "bytes to Chipper (call took", int(sendTime), "ms)")
	return nil
}

// bufferRoutine uses channels to ensure that the main routine can always add bytes to
// the current buffer without stalling if the GRPC streaming routine is blocked on
// something and not ready for more audio yet
func (strm *Streamer) bufferRoutine(streamSize int) {
	defer close(strm.byteChan)
	defer close(strm.audioStream)
	audioBuf := make([]byte, 0, streamSize*2)
	// function to enable/disable streaming case depending on whether we have enough bytes
	// to send audio
	streamData := func() (chan<- []byte, []byte) {
		if len(audioBuf) >= streamSize {
			return strm.audioStream, audioBuf[:streamSize]
		}
		return nil, nil
	}
	for {
		// depending on who's ready, either add more bytes to our buffer from main routine
		// or send audio to upload routine (assuming our validation functions allow it)
		streamChan, streamBuf := streamData()
		select {
		case streamChan <- streamBuf:
			audioBuf = audioBuf[streamSize:]
		case buf := <-strm.byteChan:
			audioBuf = append(audioBuf, buf...)
		case <-strm.ctx.Done():
			return
		}
	}
}

func (strm *Streamer) addBytes(buf []byte) {
	strm.byteChan <- buf
}

func (strm *Streamer) testRoutine(streamSize int) {
	// start go routine to shove random bytes up to the server. This is more conservative,
	// as these will have a worse compression ratio
	buf := make([]byte, streamSize)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	opts := strm.opts.checkOpts
	numChunks := int(opts.TotalAudioMs / opts.AudioPerRequestMs)
	for i := 0; i < numChunks; i++ {
		n, err := r.Read(buf)
		if n != streamSize || err != nil {
			log.Println("Error reading random bytes for connection check")
		}
		strm.addBytes(buf)
		if util.SleepSelect(time.Duration(opts.AudioPerRequestMs)*time.Millisecond, strm.ctx.Done()) {
			return
		}
	}
	return
}

// responseRoutine should be started after streaming begins, and will wait for a response
// to send back to the main routine on the given channels
func (strm *Streamer) responseRoutine() {
	resp, err := strm.conn.WaitForResponse()
	strm.respOnce.Do(func() {
		if strm.closed {
			if err != nil {
				log.Println("Ignoring error on closed context")
			} else {
				log.Println("Ignoring response on closed context")
			}
			return
		}
		if err != nil {
			log.Println("CCE error:", err)
			strm.receiver.OnError(errorReason(err), err)
			return
		}
		if verbose {
			log.Println("Intent response ->", resp)
		} else {
			log.Println("Intent response ->", fmt.Sprintf("%T", resp))
		}
		switch r := resp.(type) {
		case *chipper.IntentGraphResponse:
			sendIntentGraphResponse(r, strm.receiver)
		case *chipper.KnowledgeGraphResponse:
			sendKGResponse(r, strm.receiver, false)
		case *chipper.ConnectionCheckResponse:
			sendConnectionCheckResponse(r, strm.receiver, strm.opts.checkOpts)
		default:
			log.Println("Unexpected response type:", fmt.Sprintf("%T", resp))
		}
	})
}

func (strm *Streamer) cancelResponse() {
	done := strm.ctx.Done()
	if done == nil {
		return
	}
	<-strm.ctx.Done()
	// if we pulled from the Done channel because of a call to strm.cancel(), the closed bool will be set
	if strm.closed {
		return
	}
	strm.respOnce.Do(func() {
		strm.receiver.OnError(cloud.ErrorType_Timeout, strm.ctx.Err())
	})
}

func sendIntentResponse(resp *chipper.IntentResult, receiver Receiver) {
	metadata := fmt.Sprintf("text: %s  confidence: %f  handler: %s",
		resp.QueryText, resp.IntentConfidence, resp.Service)
	var buf bytes.Buffer
	if resp.Parameters != nil && len(resp.Parameters) > 0 {
		encoder := json.NewEncoder(&buf)
		if err := encoder.Encode(resp.Parameters); err != nil {
			log.Println("JSON encode error:", err)
			receiver.OnError(cloud.ErrorType_Json, err)
			return
		}
	}

	receiver.OnIntent(&cloud.IntentResult{Intent: resp.Action, Parameters: buf.String(), Metadata: metadata})
}

func sendKGResponse(resp *chipper.KnowledgeGraphResponse, receiver Receiver, bypass bool) {
	var buf bytes.Buffer
	params := map[string]string{
		"answer":      resp.SpokenText,
		"answer_type": resp.CommandType,
		"query_text":  resp.QueryText,
	}
	for i, d := range resp.DomainsUsed {
		params[fmt.Sprintf("domains.%d", i)] = d
	}
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(params); err != nil {
		log.Println("JSON encode error:", err)
		receiver.OnError(cloud.ErrorType_Json, err)
		return
	}

	// Bypass if its an intent graph response
	intentName := "intent_knowledge_response_extend"
	if bypass {
		intentName = "intent_knowledge_response_extend_bypass"
	}

	receiver.OnIntent(&cloud.IntentResult{
		Intent:     intentName,
		Parameters: buf.String(),
		Metadata:   "",
	})
}

func sendIntentGraphResponse(resp *chipper.IntentGraphResponse, receiver Receiver) {
	if chipper.IsIntent(*resp) {
		sendIntentResponse(chipper.ConvertToIntentResp(resp), receiver)
	} else {
		sendKGResponse(chipper.ConvertToKnowledgeGraphResp(resp), receiver, true)
	}
}

func sendConnectionCheckResponse(resp *chipper.ConnectionCheckResponse, receiver Receiver, opts *chipper.ConnectOpts) {
	toSend := &cloud.ConnectionResult{
		Status:          resp.Status,
		NumPackets:      uint8(resp.FramesReceived),
		ExpectedPackets: uint8(opts.TotalAudioMs / opts.AudioPerRequestMs),
	}
	if resp.Status == "Success" {
		toSend.Code = cloud.ConnectionCode_Available
	} else {
		toSend.Code = cloud.ConnectionCode_Bandwidth
	}
	receiver.OnConnectionResult(toSend)
}

func errorReason(err error) cloud.ErrorType {
	if err == context.DeadlineExceeded {
		return cloud.ErrorType_Timeout
	}
	return cloud.ErrorType_Server
}

var verbose bool

func logVerbose(a ...interface{}) {
	if !verbose {
		return
	}
	log.Println(a...)
}
