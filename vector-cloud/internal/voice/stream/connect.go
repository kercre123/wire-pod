package stream

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"time"

	"github.com/digital-dream-labs/vector-cloud/internal/clad/cloud"

	"github.com/digital-dream-labs/vector-cloud/internal/config"
	"github.com/digital-dream-labs/vector-cloud/internal/log"
	"github.com/digital-dream-labs/vector-cloud/internal/robot"
	"github.com/digital-dream-labs/vector-cloud/internal/util"

	"github.com/gwatts/rootcerts"

	"github.com/digital-dream-labs/api-clients/chipper"
	"github.com/google/uuid"
	"google.golang.org/grpc/credentials"
)

const (
	HeadRequestTimeout = 8 * time.Second
)

func (strm *Streamer) newChipperConn(ctx context.Context) (Conn, *CloudError) {
	if strm.opts.checkOpts != nil {
		// for connection check, first try the connection check URL with no tls
		esn, _ := robot.ReadESN()
		ankiver := robot.AnkiVersion()
		victorver := robot.VictorVersion()
		suffix := "?emresn=" + esn + "&ankiversion=" + ankiver + "&victorversion=" + victorver
		otaURL := "http://" + config.Env.Check + suffix
		req, err := http.NewRequest("HEAD", otaURL, nil)
		if err != nil {
			log.Println("Error creating CDN server http head request:", err)
			return nil, &CloudError{cloud.ErrorType_Connectivity, err}
		}

		agent := "Victor-CCHECK/" + ankiver
		req.Header.Set("User-Agent", agent)

		if resp, err := http.DefaultClient.Do(req.WithContext(ctx)); err != nil {
			log.Println("Error requesting head of CDN server:", err)
			return nil, &CloudError{cloud.ErrorType_Connectivity, err}
		} else {
			resp.Body.Close()
			log.Println("Successfully dialed CDN")
		}

		// for connection check, next try a simple https connection to our connection check endpoint
		httpsClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: rootcerts.ServerCertPool(),
				},
			},
		}

		otaURL = "https://" + config.Env.Check + suffix
		req, err = http.NewRequest("HEAD", otaURL, nil)
		if err != nil {
			log.Println("Error creating CDN server https head request:", err)
			return nil, &CloudError{cloud.ErrorType_TLS, err}
		}

		req.Header.Set("User-Agent", agent)

		if resp, err := httpsClient.Do(req.WithContext(ctx)); err != nil {
			log.Println("Error requesting head of CDN server over https:", err)
			strm.receiver.OnError(cloud.ErrorType_TLS, err)
			return nil, &CloudError{cloud.ErrorType_TLS, err}
		} else {
			resp.Body.Close()
			log.Println("Successfully dialed CDN over https")
		}
	}

	var creds credentials.PerRPCCredentials
	var err error
	var tokenTime float64
	if strm.opts.tokener != nil {
		tokenTime = util.TimeFuncMs(func() {
			creds, err = strm.opts.tokener.Credentials()
		})
	}
	if strm.opts.requireToken {
		if creds == nil && err == nil {
			err = errors.New("token required, got empty string")
		}
		if err != nil {
			return nil, &CloudError{cloud.ErrorType_Token, err}
		}
	}

	sessionID := uuid.New().String()[:16]
	var c chipperConn
	var cerr *CloudError
	connectTime := util.TimeFuncMs(func() {
		c.conn, c.stream, cerr = strm.openChipperStream(ctx, creds, sessionID)
	})
	if cerr != nil {
		log.Println("Error creating Chipper:", cerr.Err)
		return nil, cerr
	}

	// signal to engine that we got a connection; the _absence_ of this will
	// be used to detect server timeout errors
	strm.receiver.OnStreamOpen(sessionID)

	logVerbose("Received hotword event", strm.opts.mode, "created session", sessionID, "in",
		int(connectTime), "ms (token", int(tokenTime), "ms)")
	return &c, nil
}

func (strm *Streamer) openChipperStream(ctx context.Context, creds credentials.PerRPCCredentials,
	sessionID string) (*chipper.Conn, chipper.Stream, *CloudError) {

	opts := platformOpts
	if grpcOpts := util.CommonGRPC(); grpcOpts != nil {
		opts = append(opts, chipper.WithGrpcOptions(grpcOpts...))
	}
	if creds != nil {
		opts = append(opts, chipper.WithCredentials(creds))
	}
	opts = append(opts, chipper.WithSessionID(sessionID))
	opts = append(opts, chipper.WithFirmwareVersion(robot.OSVersion()))
	opts = append(opts, chipper.WithBootID(robot.BootID()))
	conn, err := chipper.NewConn(ctx, strm.opts.url, "", opts...)
	if err != nil {
		log.Println("Error getting chipper connection:", err)
		return nil, nil, &CloudError{cloud.ErrorType_Connecting, err}
	}
	var stream chipper.Stream

	switch {
	case strm.opts.checkOpts != nil:
		stream, err = conn.NewConnectionStream(ctx, *strm.opts.checkOpts)
	case strm.opts.intentGraphOpts != nil:
		stream, err = conn.NewIntentGraphStream(ctx, *strm.opts.intentGraphOpts)
	case strm.opts.kgOpts != nil:
		stream, err = conn.NewKGStream(ctx, *strm.opts.kgOpts)
	default:
		err = errors.New("fatal error: all stream option types are nil")
	}

	if err != nil {
		return nil, nil, &CloudError{cloud.ErrorType_NewStream, err}
	}
	return conn, stream, nil
}
