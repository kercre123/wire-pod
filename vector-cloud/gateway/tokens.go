package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	cloud_clad "github.com/digital-dream-labs/vector-cloud/internal/clad/cloud"

	"github.com/digital-dream-labs/vector-cloud/internal/ipc"
	"github.com/digital-dream-labs/vector-cloud/internal/log"
	"github.com/digital-dream-labs/vector-cloud/internal/robot"
	"github.com/digital-dream-labs/vector-cloud/internal/token"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const (
	jdocDomainSocket = "jdocs_server"
	jdocSocketSuffix = "gateway_client"
	tokensFile       = "/data/vic-gateway/token-hashes.json"
)

// ClientToken holds the tuple of the client token hash and the
// user-visible client name (e.g. Adam's iPhone)
type ClientToken struct {
	Hash       string `json:"hash"`
	ClientName string `json:"client_name"`
	AppId      string `json:"app_id"`
	IssuedAt   string `json:"issued_at"`
}

// ClientTokenManager holds all the client token tuples for a given
// userid+robot, along with a handle to the Jdocs service document
// that stores them.
// Note: comes from the ClientTokenDocument definition
type ClientTokenManager struct {
	ClientTokens      []ClientToken      `json:"client_tokens"`
	jdocIPC           IpcManager         `json:"-"`
	checkValid        chan struct{}      `json:"-"`
	notifyValid       chan struct{}      `json:"-"`
	updateNowChan     chan chan struct{} `json:"-"`
	recentTokenIndex  int                `json:"-"`
	lastUpdatedTokens time.Time          `json:"-"`
	forceClearFile    bool               `json:"-"`
	limiter           *MultiLimiter      `json:"-"`
}

func (ctm *ClientTokenManager) Init() error {
	ctm.forceClearFile = false
	ctm.lastUpdatedTokens = time.Now().Add(-24 * time.Hour) // older than our startup time
	// Limit the updates of the AppTokens with the following logic:
	// - At the maximum, allow 1 update per minute
	// - After 3 requests (within 15 minutes) allow 1 update per 15 minutes
	// - After 6 requests (within an hour) allow 1 update per hour
	// The limiters will refill over time, and everything will be back to the initial state
	// after 6 hours of inactivity have passed.
	//
	// In the case that a user is connected for over 6 hours (which shouldn't be the norm),
	// it's fine to not keep checking because the valid tokens should not be take out from
	// under a user too often.
	//
	// Due to the ordering of MultiLimiter, it will acquire the shorter tokens, and only grab
	// acquire the longer tokens if it was able to grab a shorter one.
	ctm.limiter = NewMultiLimiter(
		rate.NewLimiter(rate.Every(time.Minute), 1),
		rate.NewLimiter(rate.Every(15*time.Minute), 3),
		rate.NewLimiter(rate.Every(time.Hour), 6),
	)
	ctm.checkValid = make(chan struct{})
	ctm.notifyValid = make(chan struct{})
	ctm.updateNowChan = make(chan chan struct{})
	ctm.jdocIPC.Connect(ipc.GetSocketPath(jdocDomainSocket), jdocSocketSuffix)
	err := ctm.readTokensFile()
	if err != nil {
		return ctm.UpdateTokens()
	}
	return nil
}

func (ctm *ClientTokenManager) Close() error {
	return ctm.jdocIPC.Close()
}

func (ctm *ClientTokenManager) readTokensFile() error {
	clientTokens, err := ioutil.ReadFile(tokensFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(clientTokens, ctm)
}

func (ctm *ClientTokenManager) writeTokensFile(data []byte) error {
	return ioutil.WriteFile(tokensFile, data, 0600)
}

func (ctm *ClientTokenManager) CheckToken(clientToken string) (string, error) {
	ctm.checkValid <- struct{}{}
	<-ctm.notifyValid
	if len(ctm.ClientTokens) == 0 {
		return "", grpc.Errorf(codes.Unauthenticated, "no valid tokens")
	}
	recentToken := ctm.ClientTokens[ctm.recentTokenIndex]
	err := token.CompareHashAndToken(recentToken.Hash, clientToken)
	if err == nil {
		return recentToken.ClientName, nil
	}
	for idx, validToken := range ctm.ClientTokens {
		if idx == ctm.recentTokenIndex || len(validToken.Hash) == 0 {
			continue
		}
		err = token.CompareHashAndToken(validToken.Hash, clientToken)
		if err == nil {
			ctm.recentTokenIndex = idx
			return validToken.ClientName, nil
		}
	}
	return "", grpc.Errorf(codes.Unauthenticated, "invalid token")
}

// DecodeTokenJdoc will update existing valid tokens, from a jdoc received from the server
func (ctm *ClientTokenManager) DecodeTokenJdoc(jdoc []byte) error {
	ctm.recentTokenIndex = 0
	ctm.lastUpdatedTokens = time.Now()
	err := json.Unmarshal(jdoc, ctm)
	if err != nil {
		log.Printf("Unmarshal tokens failed. Invalidating. %s\n", err.Error())
		ctm.ClientTokens = []ClientToken{}
	} else {
		log.Println("Updated valid tokens")
	}
	return err
}

// UpdateTokens polls the server for new tokens, and will update as necessary
func (ctm *ClientTokenManager) UpdateTokens() error {
	if ctm.forceClearFile {
		err := os.Remove(tokensFile)
		if err == nil {
			ctm.forceClearFile = false
		}
	}
	id, esn, err := ctm.getIDs()
	if err != nil {
		return err
	}
	resp, err := ctm.sendBlock(cloud_clad.NewDocRequestWithRead(&cloud_clad.ReadRequest{
		Account: id,
		Thing:   fmt.Sprintf("vic:%s", esn),
		Items: []cloud_clad.ReadItem{
			cloud_clad.ReadItem{
				DocName:      "vic.AppTokens",
				MyDocVersion: 0,
			},
		},
	}))
	if err != nil {
		return err
	}
	read := resp.GetRead()
	if read == nil {
		return fmt.Errorf("error while trying to read jdocs: %#v", resp)
	}
	if len(read.Items) == 0 {
		return errors.New("no jdoc in read response")
	}
	data := []byte(read.Items[0].Doc.JsonDoc)
	err = ctm.DecodeTokenJdoc(data)
	if err != nil {
		return nil
	}
	err = ctm.writeTokensFile(data)
	if err != nil {
		return nil
	}
	return nil
}

func (ctm *ClientTokenManager) getIDs() (string, string, error) {
	resp, err := ctm.sendBlock(cloud_clad.NewDocRequestWithUser(&cloud_clad.Void{}))
	if err != nil {
		return "", "", err
	}
	user := resp.GetUser()
	if user == nil {
		return "", "", fmt.Errorf("Unable to get robot's user id:  %#v", resp)
	}
	esn, err := robot.ReadESN()
	if err != nil {
		return "", "", err
	}

	return user.UserId, esn, nil
}

func (ctm *ClientTokenManager) sendBlock(request *cloud_clad.DocRequest) (*cloud_clad.DocResponse, error) {
	// Write the request
	var err error
	var buf bytes.Buffer
	if request == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "Unable to parse request")
	}

	// The domain socket to vic-cloud does not have size in front of messages

	if err = request.Pack(&buf); err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}
	log.Printf("%T.sendBlock: writing DocRequest message to JDoc Manager\n", *ctm)
	// TODO: use channel to a jdoc read/write manager goroutine
	_, err = ctm.jdocIPC.conn.Write(buf.Bytes())
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}
	// Read the response
	msgBuffer := ctm.jdocIPC.conn.ReadBlock()
	if msgBuffer == nil {
		log.Errorf("%T.sendBlock: engine socket returned empty message\n", *ctm)
		return nil, grpc.Errorf(codes.Internal, "engine socket returned empty message")
	}
	var recvBuf bytes.Buffer
	recvBuf.Write(msgBuffer)
	msg := &cloud_clad.DocResponse{}
	if err := msg.Unpack(&recvBuf); err != nil {
		log.Errorf("%T.sendBlock: Unpack response error = %#v\n", *ctm, err)
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}
	return msg, nil
}

func (ctm *ClientTokenManager) ForceUpdate(response chan struct{}) {
	ctm.forceClearFile = true
	ctm.recentTokenIndex = 0
	ctm.ClientTokens = []ClientToken{}
	ctm.updateNowChan <- response
}

func (ctm *ClientTokenManager) updateListener() {
	for range ctm.checkValid {
		if time.Since(ctm.lastUpdatedTokens) > time.Hour && ctm.limiter.Allow() {
			ctm.updateNowChan <- ctm.notifyValid
		} else {
			ctm.notifyValid <- struct{}{}
		}
	}
}

func (ctm *ClientTokenManager) StartUpdateListener() {
	go ctm.updateListener()
	for {
		select {
		case response := <-ctm.updateNowChan:
			err := ctm.UpdateTokens()
			if err != nil {
				log.Printf("Unable to update tokens: %s", err.Error())
			}
			if response != nil {
				response <- struct{}{}
			}
		}
	}
}
