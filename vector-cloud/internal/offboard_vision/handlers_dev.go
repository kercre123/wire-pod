// +build !shipping

package offboard_vision

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/digital-dream-labs/vector-cloud/internal/clad/vision"

	"github.com/digital-dream-labs/vector-cloud/internal/ipc"
	"github.com/digital-dream-labs/vector-cloud/internal/log"

	"github.com/gwatts/rootcerts"
)

const (
	baseDir   = "/anki/data/assets/cozmo_resources/webserver/cloud/offboard_vision"
	imageDir  = baseDir + "/images"
	cacheDir  = "/data/data/com.anki.victor/cache/offboard_vision"
	imgPrefix = "/offboard_vision/images/"
)

func init() {
	devHandlers = func(s *http.ServeMux) {
		s.HandleFunc("/offboard_vision/", offboardVisionHandler)
		s.HandleFunc("/offboard_vision/request", reqHandler)

		s.Handle(imgPrefix, http.StripPrefix(imgPrefix, http.HandlerFunc(imgHandler)))

		log.Println("Offboard vision dev handlers added")
	}
	devURLReader = fetchURLData
}

func offboardVisionHandler(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir(imageDir)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error listing images: ", err)
		return
	}

	// cache dir is optional, skip it if it causes an error
	if cacheFiles, err := ioutil.ReadDir(cacheDir); err == nil {
		files = append(files, cacheFiles...)
	}

	t, err := template.ParseFiles(baseDir + "/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error parsing template: ", err)
		return
	}

	if err := t.Execute(w, files); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error executing template: ", err)
		return
	}
}

var offboardVisionClient ipc.Conn

func reqHandler(w http.ResponseWriter, r *http.Request) {
	// initialize connection to ipc server
	if offboardVisionClient == nil {
		var err error
		// TODO: refactor this to use chan instead because these messages are being handled by the same process in server_dev.go
		if offboardVisionClient, err = ipc.NewUnixgramClient(ipc.GetSocketPath("offboard_vision_server"), "offboard_vision_dev_client"); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Error connecting to server: ", err)
			return
		}
	}

	reqFile := r.URL.Query().Get("file")
	if reqFile == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Request must include file=[filename]")
		return
	}

	var file string

	var isURL bool
	escapedFile, err := url.PathUnescape(reqFile)
	if err == nil {
		if u, err := url.Parse(escapedFile); err == nil && u.Scheme != "" {
			isURL = true
			file = escapedFile
		}
	}

	// make sure file exists - first try image path, then cache dir
	if !isURL {
		file = path.Join(imageDir, reqFile)
		if _, err := os.Stat(file); os.IsNotExist(err) {
			file = path.Join(cacheDir, reqFile)
			if _, err := os.Stat(file); os.IsNotExist(err) {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, "File does not exist: ", reqFile)
				return
			}
		}
	}

	// NOTE: We are no longer strongly-typing this to ease development
	reqTypeString := r.URL.Query().Get("type")

	var msg vision.OffboardImageReady
	msg.Filename = file
	msg.ProcTypes = append(msg.ProcTypes, reqTypeString)

	var buf bytes.Buffer
	if err := msg.Pack(&buf); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error packing message: ", err)
		return
	}

	if _, err := offboardVisionClient.Write(buf.Bytes()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error sending message: ", err)
		return
	}

	respBuf := offboardVisionClient.ReadBlock()
	var resp vision.OffboardResultReady
	if err := resp.Unpack(bytes.NewBuffer(respBuf)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error unpacking response: ", err)
		return
	}

	fmt.Fprint(w, "image: ", file, "\n\n")

	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, []byte(resp.JsonResult), "", "\t")

	fmt.Fprint(w, prettyJSON.String())
}

func imgHandler(w http.ResponseWriter, r *http.Request) {
	file := path.Join(imageDir, r.URL.Path)
	var dir string
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		dir = imageDir
	} else {
		dir = cacheDir
	}
	http.FileServer(http.Dir(dir)).ServeHTTP(w, r)
}

func fetchURLData(path string) ([]byte, error, bool) {
	u, err := url.Parse(path)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, nil, false
	}

	httpClient := http.DefaultClient
	if u.Scheme == "https" {
		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: rootcerts.ServerCertPool(),
				},
			},
		}
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err, true
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err, true
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err, true
	}
	return buf, nil, true
}
