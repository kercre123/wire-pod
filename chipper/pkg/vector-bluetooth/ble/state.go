package ble

import "sync"

type state struct {
	nonceResponse []byte
	authorized    bool
	clientGUID    string
	filedownload  filedownload
	mutex         sync.RWMutex
}

func (s *state) setNonce(arg []byte) {
	s.mutex.Lock()
	s.nonceResponse = arg
	s.mutex.Unlock()
}

func (s *state) getNonce() []byte {
	s.mutex.RLock()
	r := s.nonceResponse
	s.mutex.RUnlock()
	return r
}

func (s *state) setAuth(arg bool) {
	s.mutex.Lock()
	s.authorized = arg
	s.mutex.Unlock()
}

func (s *state) getAuth() bool {
	s.mutex.RLock()
	r := s.authorized
	s.mutex.RUnlock()
	return r
}

func (s *state) setClientGUID(arg string) {
	s.mutex.Lock()
	s.clientGUID = arg
	s.mutex.Unlock()
}

func (s *state) getClientGUID() string {
	s.mutex.RLock()
	r := s.clientGUID
	s.mutex.RUnlock()
	return r
}

func (s *state) setFiledownload(arg filedownload) {
	s.mutex.Lock()
	s.filedownload = arg
	s.mutex.Unlock()
}

func (s *state) getFiledownload() filedownload {
	s.mutex.RLock()
	r := s.filedownload
	s.mutex.RUnlock()
	return r
}
