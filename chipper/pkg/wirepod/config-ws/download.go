package webserver

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
	"strconv"
	"strings"
	"time"

	"github.com/kercre123/chipper/pkg/logger"
)

var DownloadStatus string = "not downloading"

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
		logger.Println("error opening zip file: %v", err)
		return
	}
	defer zipReader.Close()
	DownloadStatus = "Unpacking model..."

	for _, f := range zipReader.File {
		rc, err := f.Open()
		if err != nil {
			DownloadStatus = "error downloading: " + err.Error()
			logger.Println("error opening zip file: %v", err)
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
				logger.Println("Error creating file: %v", err)
				return
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				DownloadStatus = "error downloading: " + err.Error()
				logger.Println("Error writing to file: %v", err)
				return
			}
		}
	}
}
