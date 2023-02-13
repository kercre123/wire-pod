package ble

const (
	rtsV2 = 2
	rtsV3 = 3
	rtsV4 = 4
	rtsV5 = 5
)

var rtsHandlers = map[string]func(v *VectorBLE, msg interface{}) ([]byte, bool, error){
	"RtsConnRequest":             handleRtsConnRequest,
	"RtsNonceMessage":            handleRTSNonceRequest,
	"RtsChallengeMessage":        handleRTSChallengeMessage,
	"RtsChallengeSuccessMessage": handleRTSChallengeSuccessMessage,
	"RtsStatusResponse2":         handleRSTStatusResponse,
	"RtsStatusResponse3":         handleRSTStatusResponse,
	"RtsStatusResponse4":         handleRSTStatusResponse,
	"RtsStatusResponse5":         handleRSTStatusResponse,
	"RtsWifiScanResponse2":       handleRSTWifiScanResponse,
	"RtsWifiScanResponse3":       handleRSTWifiScanResponse,
	"RtsWifiConnectResponse":     handleRSTWifiConnectionResponse,
	"RtsWifiConnectResponse3":    handleRSTWifiConnectionResponse,
	"RtsOtaUpdateResponse":       handleRSTOtaUpdateResponse,
	"RtsCloudSessionResponse":    handleRSTCloudSessionResponse,
	"RtsSdkProxyResponse":        handleRSTSDKProxyResponse,
	"RtsWifiIpResponse":          handleRSTWifiIPResponse,
	"RtsWifiAccessPointResponse": handleRSTWifiAccessPointResponse,
	"RtsWifiForgetResponse":      handleRSTWifiForgetResponse,
	"RtsLogResponse":             handleRtsLogResponse,
	"RtsFileDownload":            handleRtsFileDownload,
	"RtsForceDisconnect":         handleRTSForceDisconnect,
}
