package ble

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kercre123/chipper/pkg/vector-bluetooth/rts"
)

// LogResponse is the unified response for log downloading
type LogResponse struct {
	Filename string `json:"filename,omitempty"`
}

// Marshal converts a LogResponse message to bytes
func (sr *LogResponse) Marshal() ([]byte, error) {
	return json.Marshal(sr)
}

// Unmarshal converts a LogResponse byte slice to a LogResponse
func (sr *LogResponse) Unmarshal(b []byte) error {
	return json.Unmarshal(b, sr)
}

// DownloadLogs is an appropriately named function.
func (v *VectorBLE) DownloadLogs() (*LogResponse, error) {
	if !v.state.getAuth() {
		return nil, errors.New(errNotAuthorized)
	}

	msg, err := rts.BuildLogRequestMessage(v.ble.Version())
	if err != nil {
		return nil, err
	}

	if err := v.ble.Send(msg); err != nil {
		return nil, err
	}

	b, err := v.watch()

	resp := LogResponse{}
	if err := resp.Unmarshal(b); err != nil {
		return nil, err
	}

	return &resp, err
}

func handleRtsLogResponse(v *VectorBLE, msg interface{}) (data []byte, cont bool, err error) {
	var sr *rts.RtsLogResponse

	switch v.ble.Version() {
	case rtsV2:
		t, ok := msg.(*rts.RtsConnection_2)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		sr = t.GetRtsLogResponse()

	case rtsV3:
		t, ok := msg.(*rts.RtsConnection_3)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		sr = t.GetRtsLogResponse()

	case rtsV4:
		t, ok := msg.(*rts.RtsConnection_4)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		sr = t.GetRtsLogResponse()

	case rtsV5:
		t, ok := msg.(*rts.RtsConnection_5)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		sr = t.GetRtsLogResponse()

	default:
		return handlerUnsupportedVersionError()
	}

	if sr.FileId != 0 {
		v.state.setFiledownload(
			filedownload{
				FileID: sr.FileId,
				//File:   f,
			},
		)
	}

	return nil, true, nil
}

func handleRtsFileDownload(v *VectorBLE, msg interface{}) (data []byte, cont bool, err error) {
	var sr *rts.RtsFileDownload

	switch v.ble.Version() {
	case rtsV2:
		t, ok := msg.(*rts.RtsConnection_2)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		sr = t.GetRtsFileDownload()

	case rtsV3:
		t, ok := msg.(*rts.RtsConnection_3)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		sr = t.GetRtsFileDownload()

	case rtsV4:
		t, ok := msg.(*rts.RtsConnection_4)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		sr = t.GetRtsFileDownload()

	case rtsV5:
		t, ok := msg.(*rts.RtsConnection_5)
		if !ok {
			return handlerUnsupportedTypeError()
		}
		sr = t.GetRtsFileDownload()

	default:
		return handlerUnsupportedVersionError()
	}

	v.sendLogStatus(
		&StatusCounter{
			PacketNumber: sr.PacketNumber,
			PacketTotal:  sr.PacketTotal,
		},
	)

	switch {
	case sr.FileId != v.state.filedownload.FileID:
		v.state.setFiledownload(filedownload{})
		return nil, false, errors.New("invalid file")

	case sr.PacketNumber < sr.PacketTotal:
		v.state.filedownload.Buffer = append(v.state.filedownload.Buffer, sr.FileChunk...)
		return nil, true, nil

	case sr.PacketNumber == sr.PacketTotal:
		v.state.filedownload.Buffer = append(v.state.filedownload.Buffer, sr.FileChunk...)

		fn, err := v.writeFile()
		if err != nil {
			return nil, false, errors.New("fatal error")
		}

		resp := LogResponse{
			Filename: fn,
		}

		b, err := resp.Marshal()
		if err != nil {
			return nil, false, errors.New("fatal error")
		}

		v.state.setFiledownload(filedownload{})

		return b, false, nil

	default:
		// something bad happened...
		return nil, false, errors.New("fatal error")
	}
}

func (v *VectorBLE) writeFile() (string, error) {
	now := time.Now().Format(time.RFC3339)
	filename := fmt.Sprintf("%s/%s.tar.bz2", v.logdir, now)

	output, err := os.OpenFile(
		filename,
		os.O_APPEND|os.O_RDWR|os.O_CREATE,
		0600,
	)
	if err != nil {
		return "", err
	}

	defer output.Close()

	buf := bytes.NewReader(v.state.filedownload.Buffer)
	if _, err := io.Copy(output, buf); err != nil {
		return "", err
	}

	return filename, nil
}
