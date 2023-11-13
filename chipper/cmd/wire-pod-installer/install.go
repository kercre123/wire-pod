package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func InstallWirePod(is InstallSettings) error {
	UpdateInstallStatus("Stopping any wire-pod instances...")
	StopWirePodIfRunning()

	UpdateInstallStatus("Removing any wire-pod files (if they exist)...")
	os.RemoveAll(is.Where)

	UpdateInstallBar(0)
	UpdateInstallStatus("Starting download...")

	// Start downloading the file
	resp, err := http.Get(amd64podURL)
	if err != nil {
		return fmt.Errorf("error getting wire-pod from GitHub: %s", err)
	}
	defer resp.Body.Close()

	totalBytes := resp.ContentLength
	var bytesRead int64 = 0

	// Create a temporary file to store the download
	UpdateInstallStatus("Creating temp file...")
	tempFile, err := os.CreateTemp("", "wire-pod-*.zip")
	if err != nil {
		return fmt.Errorf("error creating a temp file: %s", err)
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name()) // Clean up

	// Copy the download stream to the temp file with progress tracking
	UpdateInstallStatus("Downloading wire-pod from latest release on GitHub...")
	progressReader := io.TeeReader(resp.Body, tempFile)
	buffer := make([]byte, 32*1024) // 32KB buffer
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

	// Open the zip file
	zipReader, err := zip.OpenReader(tempFile.Name())
	if err != nil {
		return fmt.Errorf("error reading zip file: %s", err)
	}
	defer zipReader.Close()

	// Process each file in the zip
	for i, f := range zipReader.File {
		// Skip the root directory
		if f.Name == "wire-pod/" || strings.HasPrefix(f.Name, "wire-pod/") && f.FileInfo().IsDir() {
			continue
		}

		// Adjust the file path to exclude the 'wire-pod/' prefix
		adjustedPath := strings.TrimPrefix(f.Name, "wire-pod/")
		UpdateInstallStatus("Extracting: " + adjustedPath)
		fpath := filepath.Join(is.Where, adjustedPath)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(fpath, filepath.Clean(is.Where)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		// Create all directories needed for the file path
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Create the directories if necessary
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return fmt.Errorf("error creating directories: %s", err)
		}

		// Extract the file
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

		// Update the progress bar for each file processed
		UpdateInstallBar(40 + float64(40)*(float64(i)+1)/float64(len(zipReader.File)))
	}

	// Update status and progress bar for the final phase
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
	return nil
}
