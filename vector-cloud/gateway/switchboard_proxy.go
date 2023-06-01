package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/digital-dream-labs/vector-cloud/internal/log"

	gw_clad "github.com/digital-dream-labs/vector-cloud/internal/clad/gateway"

	"google.golang.org/grpc"
)

// Regex to provide two match results:
// 1: either the start of input and any lowercase characters following it
//    or the capital letters and numbers after the start
// 2: The string of a capital and multiple non-capital letters
//    or end of the input
//
// For example, this will provide matches (format = "<input>" -> [("<match1>", "<match2>"), <subsequent_matches>...]) like the following:
// "TestString" -> [("", "Test"), ("", String)]
// "SomeURLString" -> [("", "Some"), ("URL", "String")]
// "lastString" -> [("last", "String")]
var splitCamelRegex = regexp.MustCompile("(^[^A-Z0-9]*|[A-Z0-9]*)([A-Z0-9][^A-Z]+|$)")

// BLEProxy handles switchboard messages being proxied to grpc-gateway which proxies to the grpc handlers
type BLEProxy struct {
	Address    string
	Client     *http.Client
	StreamURLs []string
}

// streamNameToURL translates a grpc stream name into the equivalent rest endpoint
// ex: "BehaviorControl" -> "v1/behavior_control"
func streamNameToURL(name string) string {
	var url []string

	for _, match := range splitCamelRegex.FindAllStringSubmatch(name, -1) {
		for _, i := range []int{1, 2} {
			if match[i] != "" {
				url = append(url, match[i])
			}
		}
	}

	return strings.ToLower("v1/" + strings.Join(url, "_"))
}

// initialize determines which streams need to be added to the StreamsURL list
// which is used for blacklisting streams from the BLE proxy
func (proxy *BLEProxy) initialize(serviceInfo map[string]grpc.ServiceInfo) {
	if proxy.StreamURLs == nil {
		proxy.StreamURLs = make([]string, 0)
	}
	for _, service := range serviceInfo {
		for _, method := range service.Methods {
			if method.IsClientStream || method.IsServerStream {
				proxy.StreamURLs = append(proxy.StreamURLs, streamNameToURL(method.Name))
			}
		}
	}
}

// isStream checks StreamURLs to determine if given url is a stream
func (proxy *BLEProxy) isStream(url string) bool {
	for _, streamURL := range proxy.StreamURLs {
		if url == streamURL {
			return true
		}
	}
	return false
}

// TruncationResponse to construct a json error response when the message is
// too large for the gateway <-> switchboard interface.
type TruncationResponse struct {
	OriginalStatus uint16 `json:"original_status"`
	Reason         string `json:"reason"`
}

// handle takes a given proxy request and sends it through the proxy pipelines and returns a response
func (proxy *BLEProxy) handle(request *gw_clad.SdkProxyRequest) *gw_clad.SdkProxyResponse {
	log.Printf("Handling a switchboard proxy request id:\"%s\" path:\"%s\" json:\"%s\"\n", request.MessageId, request.Path, request.Json)
	proxyResponse := &gw_clad.SdkProxyResponse{
		MessageId: request.MessageId,
	}
	if proxy.isStream(request.Path) {
		log.Errorln("Unable to run streams through BLE interface")
		proxyResponse.StatusCode = uint16(403)
		proxyResponse.ContentType = "application/text"
		proxyResponse.Content = "Unable to run streams through BLE interface"
		return proxyResponse
	}
	path := request.Path
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	url := fmt.Sprintf("https://%s/%s", proxy.Address, path)

	jsonStr := []byte(request.Json)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", request.ClientGuid))
	req.Header.Set("Content-Type", "application/json")

	resp, err := proxy.Client.Do(req)
	if err != nil {
		log.Errorf("Failed to reach proxy server: %s\n", err.Error())
		proxyResponse.StatusCode = uint16(502)
		proxyResponse.ContentType = "application/text"
		proxyResponse.Content = "Failed to reach proxy server"
		return proxyResponse
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	proxyResponse.StatusCode = uint16(resp.StatusCode)
	proxyResponse.ContentType = resp.Header.Get("Content-Type")
	proxyResponse.Content = string(body)
	responseSize := proxyResponse.Size()
	if responseSize >= 2048 {
		content := TruncationResponse{
			OriginalStatus: uint16(resp.StatusCode),
			Reason:         fmt.Sprintf("Response body too large: %d", responseSize),
		}
		proxyResponse.StatusCode = uint16(500)
		jsonResponse, err := json.Marshal(content)
		if err != nil {
			proxyResponse.ContentType = "application/text"
			proxyResponse.Content = "Response body too large"
		} else {
			proxyResponse.ContentType = "application/json"
			proxyResponse.Content = string(jsonResponse[:])
		}
	}
	return proxyResponse
}
