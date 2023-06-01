package testconn

import "errors"

type response struct {
	obj interface{}
	err error
}

type TestConn struct {
	respChan    chan response
	AudioSends  [][]byte
	ErrorOnSend bool
}

func NewTestConn() *TestConn {
	ret := new(TestConn)
	ret.respChan = make(chan response)
	ret.AudioSends = [][]byte{}
	return ret
}

func (c *TestConn) Close() error {
	close(c.respChan)
	return nil
}

func (c *TestConn) CloseSend() error {
	return nil
}

func (c *TestConn) SendAudio(buf []byte) error {
	if c.ErrorOnSend {
		return errors.New("oh no!")
	}
	c.AudioSends = append(c.AudioSends, buf)
	return nil
}

func (c *TestConn) WaitForResponse() (interface{}, error) {
	resp := <-c.respChan
	return resp.obj, resp.err
}

func (c *TestConn) TriggerResponse(obj interface{}, err error) {
	c.respChan <- response{obj, err}
}
