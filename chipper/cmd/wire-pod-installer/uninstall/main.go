package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ncruces/zenity"
	"golang.org/x/sys/windows/registry"
)

func StopWirePodIfRunning() {
	confDir, _ := os.UserConfigDir()
	podDir := confDir + "/wire-pod"
	podPid, err := os.ReadFile(podDir + "/runningPID")
	if err == nil {
		pid, _ := strconv.Atoi(string(podPid))
		// doesn't work on unix, but should on Windows
		podProcess, err := os.FindProcess(pid)
		if err == nil {
			fmt.Println("Stopping wire-pod")
			podProcess.Kill()
			podProcess.Wait()
			fmt.Println("Stopped")
		}
	}
}

func main() {
	StopWirePodIfRunning()
	wd, _ := os.Getwd()
	os.RemoveAll(filepath.Join(wd, "chipper"))
	os.RemoveAll(filepath.Join(wd, "vector-cloud"))
	keyPath := `Software\Microsoft\Windows\CurrentVersion\Uninstall\wire-pod`
	registry.DeleteKey(registry.LOCAL_MACHINE, keyPath)
	zenity.Info(
		"wire-pod has been successfully uninstalled.",
		zenity.InfoIcon,
		zenity.Title("wire-pod uninstaller"),
	)
	os.Exit(0)
}
