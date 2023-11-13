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
	podPid, err := os.ReadFile(filepath.Join(os.TempDir(), "/wirepodrunningPID"))
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
	err := zenity.Question(
		"Would you like to remove application data, like saved bot settings or API preferences?",
		zenity.ExtraButton("No"),
	)
	if err == nil {
		conf, _ := os.UserConfigDir()
		os.RemoveAll(filepath.Join(conf, "wire-pod"))
	}
	keyPath := `Software\Microsoft\Windows\CurrentVersion\Uninstall\wire-pod`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.QUERY_VALUE)
	if err != nil {
		fmt.Println(err)
		return
	}
	val, _, err := k.GetStringValue("InstallPath")
	if err != nil {
		fmt.Println(err)
		return
	}

	k.Close()
	registry.DeleteKey(registry.LOCAL_MACHINE, keyPath)

	DontRunPodAtStartup()

	fmt.Println(val)

	os.RemoveAll(filepath.Join(val, "chipper"))
	os.RemoveAll(filepath.Join(val, "vector-cloud"))
	os.Remove("C:\\ProgramData\\Microsoft\\Windows\\Start Menu\\Programs\\wire-pod.lnk")
	zenity.Info(
		"wire-pod has successfully been uninstalled.",
		zenity.InfoIcon,
		zenity.Title("wire-pod uninstaller"),
	)
	os.RemoveAll(val)
	os.Exit(0)
}

func DontRunPodAtStartup() {
	key, _ := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	key.DeleteValue("wire-pod")
}
