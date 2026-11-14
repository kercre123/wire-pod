package localization

/**
 * @website http://albulescu.ro
 * @author Cosmin Albulescu <cosmin@albulescu.ro>
 */
// modified from this person's gist ^^^

import (
	"archive/zip"
	"crypto/tls"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

var URLPrefix string = "https://github.com/kercre123/vosk-models/raw/main/"

//var URLPrefix string = "https://alphacephei.com/vosk/models/"

var (
	downloadStatusMu sync.Mutex
	downloadStatus   = "not downloading"
	downloading      atomic.Bool
)

var downloadClient = &http.Client{
	Transport: &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

func setDownloadStatus(s string) {
	downloadStatusMu.Lock()
	downloadStatus = s
	downloadStatusMu.Unlock()
}

func TakeDownloadStatus() string {
	downloadStatusMu.Lock()
	defer downloadStatusMu.Unlock()
	s := downloadStatus
	if s == "success" || strings.Contains(s, "error") {
		downloadStatus = "not downloading"
	}
	return s
}

func DownloadVoskModel(language string) {
	filename := "vosk-model-small-"
	if language == "en-US" {
		filename = filename + "en-us-0.15.zip"
	} else if language == "it-IT" {
		filename = filename + "it-0.22.zip"
	} else if language == "es-ES" {
		filename = filename + "es-0.42.zip"
	} else if language == "fr-FR" {
		filename = filename + "fr-0.22.zip"
	} else if language == "de-DE" {
		filename = filename + "de-0.15.zip"
	} else if language == "pt-BR" {
		filename = filename + "pt-0.3.zip"
	} else if language == "pl-PL" {
		filename = filename + "pl-0.22.zip"
	} else if language == "zh-CN" {
		filename = filename + "cn-0.22.zip"
	} else if language == "tr-TR" {
		filename = filename + "tr-0.3.zip"
	} else if language == "ru-RU" {
		filename = filename + "ru-0.22.zip"
	} else if language == "nt-NL" {
		filename = filename + "nl-0.22.zip"
	} else if language == "uk-UA" {
		filename = filename + "uk-v3-small.zip"
	} else if language == "vi-VN" {
		filename = filename + "vn-0.4.zip"
	} else if language == "ko-KR" {
		filename = filename + "ko-0.22.zip"
	} else {
		logger.Println("Language not valid? " + language)
		return
	}
	if !downloading.CompareAndSwap(false, true) {
		logger.Println("Not downloading model because download is currently happening")
		return
	}
	defer downloading.Store(false)
	os.MkdirAll(vars.VoskModelPath, 0755)
	url := URLPrefix + filename
	var filep string
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		filep = filepath.Join(vars.AndroidPath, "/"+filename)
	} else {
		tmp, err := os.CreateTemp("", "vosk-model-*.zip")
		if err != nil {
			setDownloadStatus("error: " + err.Error())
			return
		}
		filep = tmp.Name()
		tmp.Close()
	}
	destpath := filepath.Join(vars.VoskModelPath, language) + "/"
	DownloadFile(url, filep)
	UnzipFile(filep, destpath)
	os.Rename(destpath+strings.TrimSuffix(filename, ".zip"), destpath+"model")
	os.Remove(filep)
	vars.AddDownloadedVoskModel(language)
	setDownloadStatus("Reloading voice processor")
	vars.UpdateConfig(func(c *vars.APIConfig) {
		c.STT.Language = language
		c.PastInitialSetup = true
	})
	vars.WriteConfigToDisk()
	ReloadVosk()
	logger.Println("Reloaded voice processor successfully")
	setDownloadStatus("success")
}

func PrintDownloadPercent(done chan int64, path string, total int64) {
	for {
		select {
		case <-done:
			return
		case <-time.After(time.Second / 2):
		}
		fi, err := os.Stat(path)
		if err != nil {
			continue
		}
		size := fi.Size()
		if size == 0 {
			size = 1
		}
		percent := float64(size) / float64(total) * 100
		setDownloadStatus("Model download status: " + fmt.Sprint(math.Floor(percent)) + "%")
	}
}

func DownloadFile(url string, dest string) {
	logger.Println("Downloading " + url + " to " + dest)
	out, _ := os.Create(dest)
	defer out.Close()
	headResp, err := downloadClient.Head(url)
	if err != nil {
		logger.Println(err)
		setDownloadStatus("error: " + err.Error())
		return
	}
	defer headResp.Body.Close()
	size, _ := strconv.Atoi(headResp.Header.Get("Content-Length"))
	done := make(chan int64)
	defer close(done)
	go PrintDownloadPercent(done, dest, int64(size))
	resp, err := downloadClient.Get(url)
	if err != nil {
		setDownloadStatus("error: " + err.Error())
		logger.Println(err)
		return
	}
	defer resp.Body.Close()
	io.Copy(out, resp.Body)
	setDownloadStatus("Completed download")
}

func UnzipFile(file, dest string) {
	zipReader, err := zip.OpenReader(file)
	if err != nil {
		setDownloadStatus("error downloading: " + err.Error())
		logger.Println("error opening zip file:", err)
		return
	}
	defer zipReader.Close()
	setDownloadStatus("Unpacking model...")

	cleanDest, err := filepath.Abs(dest)
	if err != nil {
		setDownloadStatus("error downloading: " + err.Error())
		return
	}
	for _, f := range zipReader.File {
		path := filepath.Join(cleanDest, f.Name)
		if path != cleanDest && !strings.HasPrefix(path, cleanDest+string(os.PathSeparator)) {
			setDownloadStatus("error downloading: zip entry escapes destination: " + f.Name)
			logger.Println("Refusing zip entry outside destination: " + f.Name)
			return
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, 0755)
			continue
		}
		if err := unzipEntry(f, path); err != nil {
			setDownloadStatus("error downloading: " + err.Error())
			logger.Println("Error unpacking zip entry:", err)
			return
		}
	}
}

func unzipEntry(f *zip.File, path string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode().Perm())
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, rc)
	return err
}
