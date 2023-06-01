package stream_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/digital-dream-labs/vector-cloud/internal/clad/cloud"
	"github.com/digital-dream-labs/vector-cloud/internal/log"
	"github.com/digital-dream-labs/vector-cloud/internal/voice/stream"
	"github.com/digital-dream-labs/vector-cloud/internal/voice/stream/testconn"

	"github.com/digital-dream-labs/api-clients/chipper"
	"github.com/stretchr/testify/assert"
)

type testReceiver struct {
	err    chan *stream.CloudError
	open   chan string
	intent chan *cloud.IntentResult
	result chan *cloud.ConnectionResult
}

func (r *testReceiver) OnError(kind cloud.ErrorType, err error) {
	r.err <- cloudErr(kind, err)
}

func (r *testReceiver) OnStreamOpen(str string) {
	r.open <- str
}

func (r *testReceiver) OnIntent(intent *cloud.IntentResult) {
	r.intent <- intent
}

func (r *testReceiver) OnConnectionResult(result *cloud.ConnectionResult) {
	r.result <- result
}

func (r *testReceiver) Close() {
	close(r.err)
	close(r.open)
	close(r.intent)
	close(r.result)
}

func (r *testReceiver) CouldPull(shouldLog ...bool) bool {
	l := false
	if len(shouldLog) > 0 {
		l = shouldLog[0]
	}
	maybeLog := func(args ...interface{}) {
		if !l {
			return
		}
		log.Println(args...)
	}
	select {
	case <-r.err:
		maybeLog("can pull from error channel")
	case <-r.open:
		maybeLog("can pull from stream open channel")
	case <-r.intent:
		maybeLog("can pull from intent channel")
	case <-r.result:
		maybeLog("can pull from connection result channel")
	default:
		return false
	}
	return true
}

// returns a ConnectFn that gives the streamer a TestConn, and will block until the given
// trigger (last return value) is called
func connector() (*testconn.TestConn, stream.ConnectFunc, func()) {
	conn := testconn.NewTestConn()
	ctx, cancel := context.WithCancel(context.Background())
	fn := func(context.Context) (stream.Conn, *stream.CloudError) {
		<-ctx.Done()
		return conn, nil
	}
	return conn, fn, cancel
}

func newReceiver() *testReceiver {
	return &testReceiver{
		make(chan *stream.CloudError),
		make(chan string),
		make(chan *cloud.IntentResult),
		make(chan *cloud.ConnectionResult)}
}

var receiver = newReceiver()

func TestConnectTimeout(t *testing.T) {
	// make connect func block longer than stream's timeout
	connctx, conncancel := context.WithCancel(context.Background())
	fn := func(context.Context) (stream.Conn, *stream.CloudError) {
		<-connctx.Done()
		return nil, nil
	}

	// make new context for stream client
	streamctx, streamcancel := context.WithCancel(context.Background())
	_ = stream.NewStreamer(streamctx, receiver, 1, stream.WithConnectFunc(fn))

	streamcancel()
	time.Sleep(5 * time.Millisecond)

	err := <-receiver.err
	assert.Equal(t, cloud.ErrorType_Timeout, err.Kind)
	assert.False(t, receiver.CouldPull(true))

	conncancel()
	time.Sleep(5 * time.Millisecond)
	assert.False(t, receiver.CouldPull(true))
}

func cloudErr(kind cloud.ErrorType, err error) *stream.CloudError {
	return &stream.CloudError{Kind: kind, Err: err}
}

func TestCloseInConnect(t *testing.T) {
	// make connect func block until stream is closed
	fn := func(ctx context.Context) (stream.Conn, *stream.CloudError) {
		<-ctx.Done()
		return nil, cloudErr(cloud.ErrorType_Timeout, errors.New("interrupted"))
	}

	strm := stream.NewStreamer(context.Background(), receiver, 1, stream.WithConnectFunc(fn))
	time.Sleep(5 * time.Millisecond)
	strm.Close()
	time.Sleep(5 * time.Millisecond)

	err := <-receiver.err
	assert.Equal(t, cloud.ErrorType_Timeout, err.Kind)
	assert.False(t, receiver.CouldPull(true))
}

func TestAudioBuffering(t *testing.T) {
	conn, fn, trigger := connector()

	const streamSize = 100
	strm := stream.NewStreamer(context.Background(), receiver, streamSize, stream.WithConnectFunc(fn))
	defer strm.Close()
	time.Sleep(5 * time.Millisecond)

	const sends = 80
	for i := 0; i < sends; i++ {
		if i%2 == 0 {
			strm.AddBytes(make([]byte, streamSize/2))
		} else {
			strm.AddSamples(make([]int16, streamSize/4))
		}
	}

	// before finishing connection, no audio should be sent
	assert.Empty(t, conn.AudioSends)

	trigger()
	time.Sleep(5 * time.Millisecond)
	// now we should have received 1 audio send on the server side per 2 sends (of half a chunk size)
	// into the stream
	assert.Equal(t, sends/2, len(conn.AudioSends))
	for i := 0; i < len(conn.AudioSends); i++ {
		assert.Equal(t, streamSize, len(conn.AudioSends[i]))
	}
}

func TestSendAfterResponse(t *testing.T) {
	conn, fn, trigger := connector()

	strm := stream.NewStreamer(context.Background(), receiver, 100, stream.WithConnectFunc(fn))
	defer strm.Close()
	trigger()

	const sends = 100
	for i := 0; i < sends; i++ {
		strm.AddBytes(make([]byte, 50))
		if i == 50 {
			conn.TriggerResponse(&chipper.IntentResult{}, nil)
		}
	}

	res := <-receiver.intent
	assert.NotNil(t, res)
	// Allow time to make sure the sending has a chance to finish
	time.Sleep(time.Second)
	assert.True(t, len(conn.AudioSends) == 50)
}

func TestConnectFailResponse(t *testing.T) {
	vals := []cloud.ErrorType{cloud.ErrorType_Server,
		cloud.ErrorType_Timeout,
		cloud.ErrorType_Json,
		cloud.ErrorType_InvalidConfig,
		cloud.ErrorType_Connecting,
		cloud.ErrorType_NewStream,
		cloud.ErrorType_Token,
		cloud.ErrorType_TLS,
		cloud.ErrorType_Connectivity}

	for _, val := range vals {
		fn := func(context.Context) (stream.Conn, *stream.CloudError) {
			return nil, cloudErr(val, errors.New(fmt.Sprint(val)))
		}

		strm := stream.NewStreamer(context.Background(), receiver, 1, stream.WithConnectFunc(fn))
		err := <-receiver.err
		assert.Equal(t, val, err.Kind)
		time.Sleep(5 * time.Millisecond)
		assert.False(t, receiver.CouldPull(true))
		strm.Close()
	}
}

func TestAudioSendFail(t *testing.T) {
	conn, fn, trigger := connector()
	conn.ErrorOnSend = true

	strm := stream.NewStreamer(context.Background(), receiver, 100, stream.WithConnectFunc(fn))
	trigger()

	const sends = 5
	for i := 0; i < sends; i++ {
		strm.AddBytes(make([]byte, 50))
	}
	err := <-receiver.err
	assert.Equal(t, cloud.ErrorType_Server, err.Kind)
	time.Sleep(5 * time.Millisecond)
	assert.False(t, receiver.CouldPull(true))
	strm.Close()
}

func TestConnectionCheck(t *testing.T) {
	const sends = 10
	const reqMs = 5

	resps := []*chipper.ConnectionCheckResponse{
		&chipper.ConnectionCheckResponse{Status: "Success", FramesReceived: sends},
		&chipper.ConnectionCheckResponse{},
	}
	expected := []cloud.ConnectionResult{
		cloud.ConnectionResult{Code: cloud.ConnectionCode_Available, Status: "Success", NumPackets: sends, ExpectedPackets: sends},
		cloud.ConnectionResult{Code: cloud.ConnectionCode_Bandwidth, NumPackets: 0, ExpectedPackets: sends},
	}
	assert.Equal(t, len(resps), len(expected))

	opts := chipper.ConnectOpts{
		TotalAudioMs:      reqMs * sends,
		AudioPerRequestMs: reqMs}

	for i := range resps {
		conn, fn, trigger := connector()

		strm := stream.NewStreamer(context.Background(), receiver, 100, stream.WithConnectFunc(fn),
			stream.WithConnectionCheckOptions(opts))
		trigger()

		go func() {
			for len(conn.AudioSends) < sends {
				time.Sleep(reqMs * time.Millisecond)
			}
			conn.TriggerResponse(resps[i], nil)
		}()

		res := <-receiver.result
		assert.Equal(t, expected[i], *res)
		assert.Equal(t, len(conn.AudioSends), int(opts.TotalAudioMs/opts.AudioPerRequestMs))
		time.Sleep(5 * time.Millisecond)
		assert.False(t, receiver.CouldPull(true))
		strm.Close()
	}
}

func TestKnowledgeGraph(t *testing.T) {
	conn, fn, trigger := connector()

	strm := stream.NewStreamer(context.Background(), receiver, 100, stream.WithConnectFunc(fn),
		stream.WithKnowledgeGraphOptions(chipper.KGOpts{}))
	trigger()

	const sends = 100
	for i := 0; i < sends; i++ {
		strm.AddBytes(make([]byte, 50))
	}
	conn.TriggerResponse(&chipper.KnowledgeGraphResponse{QueryText: "this is a question",
		SpokenText:  "this is an answer",
		CommandType: "cmdtype"}, nil)

	res := <-receiver.intent
	assert.IsType(t, cloud.IntentResult{}, *res)
	assert.Equal(t, len(conn.AudioSends), sends/2)
	time.Sleep(5 * time.Millisecond)
	assert.False(t, receiver.CouldPull(true))
	strm.Close()
}

func TestTimeoutOption(t *testing.T) {
	_, fn, trigger := connector()

	strm := stream.NewStreamer(context.Background(), receiver, 100, stream.WithConnectFunc(fn),
		stream.WithIntentOptions(chipper.IntentOpts{StreamOpts: chipper.StreamOpts{Timeout: 5 * time.Millisecond}},
			cloud.StreamType_Normal))
	trigger()

	time.Sleep(15 * time.Millisecond)

	err := <-receiver.err
	assert.Equal(t, cloud.ErrorType_Timeout, err.Kind)
	time.Sleep(5 * time.Millisecond)
	assert.False(t, receiver.CouldPull(true))
	strm.Close()
}
