package localization

/**
 * @website http://albulescu.ro
 * @author Cosmin Albulescu <cosmin@albulescu.ro>
 */
// modified from this person's gist ^^^

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

var URLPrefix string = "https://github.com/kercre123/vosk-models/raw/main/"

//var URLPrefix string = "https://alphacephei.com/vosk/models/"

var DownloadStatus string = "not downloading"

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
	} else {
		logger.Println("Language not valid? " + language)
		return
	}
	os.MkdirAll(vars.VoskModelPath, 0755)
	url := URLPrefix + filename
	var filep string
	if runtime.GOOS == "android" {
		filep = filepath.Join(vars.AndroidPath, "/"+filename)
	} else {
		filep = os.TempDir() + "/" + filename
	}
	destpath := filepath.Join(vars.VoskModelPath, language) + "/"
	DownloadFile(url, filep)
	UnzipFile(filep, destpath)
	os.Rename(destpath+strings.TrimSuffix(filename, ".zip"), destpath+"model")
	os.Remove(filep)
	vars.DownloadedVoskModels = append(vars.DownloadedVoskModels, language)
	DownloadStatus = "Reloading voice processor"
	vars.APIConfig.STT.Language = language
	vars.APIConfig.PastInitialSetup = true
	vars.WriteConfigToDisk()
	ReloadVosk()
	logger.Println("Reloaded voice processor successfully")
	DownloadStatus = "success"
}

func PrintDownloadPercent(done chan int64, path string, total int64) {
	var stop bool = false
	for {
		select {
		case <-done:
			stop = true
		default:
			file, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			fi, err := file.Stat()
			if err != nil {
				log.Fatal(err)
			}
			size := fi.Size()
			if size == 0 {
				size = 1
			}
			var percent float64 = float64(size) / float64(total) * 100
			showPercent := math.Floor(percent)
			DownloadStatus = "Model download status: " + fmt.Sprint(showPercent) + "%"
		}
		if stop {
			DownloadStatus = "completed"
			break
		}
		time.Sleep(time.Second / 2)
	}
}

func DownloadFile(url string, dest string) {
	if strings.Contains(DownloadStatus, "success") || strings.Contains(DownloadStatus, "error") || strings.Contains(DownloadStatus, "not downloading") {
		logger.Println("Downloading " + url + " to " + dest)
		out, _ := os.Create(dest)
		defer out.Close()
		headResp, err := http.Head(url)
		if err != nil {
			logger.Println(err)
			DownloadStatus = "error: " + err.Error()
			return
		}
		defer headResp.Body.Close()
		size, _ := strconv.Atoi(headResp.Header.Get("Content-Length"))
		done := make(chan int64)
		go PrintDownloadPercent(done, dest, int64(size))
		resp, err := http.Get(url)
		if err != nil {
			DownloadStatus = "error: " + err.Error()
			logger.Println(err)
			return
		}
		defer resp.Body.Close()
		n, _ := io.Copy(out, resp.Body)
		done <- n
		DownloadStatus = "Completed download"
	} else {
		logger.Println("Not downloading model because download is currently happening")
	}
}

func UnzipFile(file, dest string) {
	zipReader, err := zip.OpenReader(file)
	if err != nil {
		DownloadStatus = "error downloading: " + err.Error()
		logger.Println("error opening zip file:", err)
		return
	}
	defer zipReader.Close()
	DownloadStatus = "Unpacking model..."

	for _, f := range zipReader.File {
		rc, err := f.Open()
		if err != nil {
			DownloadStatus = "error downloading: " + err.Error()
			logger.Println("error opening zip file:", err)
			return
		}
		defer rc.Close()

		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				DownloadStatus = "error downloading: " + err.Error()
				logger.Println("Error creating file:", err)
				return
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				DownloadStatus = "error downloading: " + err.Error()
				logger.Println("Error writing to file:", err)
				return
			}
		}
	}
}
