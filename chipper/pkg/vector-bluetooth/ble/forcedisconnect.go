package ble

func handleRTSForceDisconnect(v *VectorBLE, msg interface{}) (data []byte, cont bool, err error) {
	v.state.setAuth(false)
	v.state.setClientGUID("")
	v.state.setNonce(nil)
	return nil, false, nil
}
