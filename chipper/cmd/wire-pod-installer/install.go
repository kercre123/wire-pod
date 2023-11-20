package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func InstallWirePod(is InstallSettings) error {
	if is.SetHostnameEpod {
		UpdateInstallStatus("Setting hostname to escapepod...")
		err := ChangeHostname("escapepod")
		if err != nil {
			return err
		}
	}

	UpdateInstallStatus("Stopping any wire-pod instances...")
	StopWirePodIfRunning()

	UpdateInstallStatus("Uninstalling any previous wire-pod instances (user data will be kept)...")
	DeleteAnyOtherInstallation()

	UpdateInstallStatus("Removing any wire-pod files (if they exist)...")
	os.RemoveAll(is.Where)

	UpdateInstallBar(0)
	UpdateInstallStatus("Starting download...")

	resp, err := http.Get(amd64podURL)
	if err != nil {
		return fmt.Errorf("error getting wire-pod from GitHub: %s", err)
	}
	defer resp.Body.Close()

	totalBytes := resp.ContentLength
	var bytesRead int64 = 0

	tempFile, err := os.CreateTemp("", "wire-pod-*.zip")
	if err != nil {
		return fmt.Errorf("error creating a temp file: %s", err)
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	UpdateInstallStatus("Downloading wire-pod from latest release on GitHub...")
	progressReader := io.TeeReader(resp.Body, tempFile)
	buffer := make([]byte, 32*1024)
	for {
		n, err := progressReader.Read(buffer)
		bytesRead += int64(n)
		if n == 0 || err != nil {
			break
		}
		UpdateInstallBar(float64(40) * float64(bytesRead) / float64(totalBytes))
	}

	if err != nil && err != io.EOF {
		return fmt.Errorf("error while downloading: %s", err)
	}

	UpdateInstallBar(40)
	UpdateInstallStatus("Starting extraction...")

	zipReader, err := zip.OpenReader(tempFile.Name())
	if err != nil {
		return fmt.Errorf("error reading zip file: %s", err)
	}
	defer zipReader.Close()

	for i, f := range zipReader.File {
		if f.Name == "wire-pod/" || strings.HasPrefix(f.Name, "wire-pod/") && f.FileInfo().IsDir() {
			continue
		}

		adjustedPath := strings.TrimPrefix(f.Name, "wire-pod/")
		UpdateInstallStatus("Extracting: " + adjustedPath)
		fpath := filepath.Join(is.Where, adjustedPath)

		if !strings.HasPrefix(fpath, filepath.Clean(is.Where)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return fmt.Errorf("error creating directories: %s", err)
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("error opening file for writing: %s", err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("error opening zip contents: %s", err)
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return fmt.Errorf("error writing file: %s", err)
		}

		UpdateInstallBar(40 + float64(40)*(float64(i)+1)/float64(len(zipReader.File)))
	}

	UpdateInstallBar(81)
	UpdateInstallStatus("Updating registry...")

	UpdateRegistry(is)
	if is.RunAtStartup {
		RunPodAtStartup(is)
	}

	UpdateInstallBar(90)

	UpdateInstallStatus("Creating shortcut...")
	CreateShortcut(is)

	UpdateInstallStatus("Creating firewall rules...")
	AllowThroughFirewall(is)

	UpdateInstallStatus("Done!")

	UpdateInstallBar(100)
	time.Sleep(time.Second / 3)
	return nil
}
