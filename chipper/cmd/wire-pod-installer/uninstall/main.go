package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ncruces/zenity"
	"golang.org/x/sys/windows/registry"
)

var discrete bool

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
	CheckWirePodRunningViaRegistry()
}

func CheckWirePodRunningViaRegistry() {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\wire-pod`, registry.WRITE|registry.READ)
	if err != nil {
		fmt.Println("Error reading from registry: " + err.Error())
		return
	}
	defer k.Close()

	// Write a value
	val, _, err := k.GetIntegerValue("LastRunningPID")
	if err != nil {
		fmt.Println("Error reading from registry: " + err.Error())
		return
	}
	// doesn't work on unix, but should on Windows
	podProcess, err := os.FindProcess(int(val))
	if err == nil || errors.Is(err, os.ErrPermission) {
		fmt.Println("Stopping wire-pod")
		podProcess.Kill()
		podProcess.Wait()
		fmt.Println("Stopped")
	}
}

func main() {
	if os.Getenv("RUN_DISCRETE") == "true" {
		discrete = true
	}
	if !discrete {
		err := zenity.Question(
			"Are you sure you want to uninstall wire-pod?",
			zenity.QuestionIcon,
			zenity.Title("wire-pod uninstaller"),
			zenity.OKLabel("Yes"),
		)
		if errors.Is(err, zenity.ErrCanceled) {
			os.Exit(1)
		}
	}
	StopWirePodIfRunning()
	if !discrete {
		err := zenity.Question(
			"Would you like to remove application data, including saved bot settings and API preferences?",
			zenity.ExtraButton("No"),
			zenity.QuestionIcon,
			zenity.NoCancel(),
			zenity.Title("wire-pod uninstaller"),
		)
		if err == nil {
			conf, _ := os.UserConfigDir()
			os.RemoveAll(filepath.Join(conf, "wire-pod"))
		}
	}
	keyPath := `Software\wire-pod`
	k, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.QUERY_VALUE)
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
	keyPath = `Software\Microsoft\Windows\CurrentVersion\Uninstall\wire-pod`
	registry.DeleteKey(registry.LOCAL_MACHINE, keyPath)
	keyPath = `Software\wire-pod`
	registry.DeleteKey(registry.CURRENT_USER, keyPath)

	DontRunPodAtStartup()

	fmt.Println(val)

	os.RemoveAll(filepath.Join(val, "chipper"))
	os.RemoveAll(filepath.Join(val, "vector-cloud"))
	os.Remove("C:\\ProgramData\\Microsoft\\Windows\\Start Menu\\Programs\\wire-pod.lnk")
	if !discrete {
		zenity.Info(
			"wire-pod has successfully been uninstalled.",
			zenity.InfoIcon,
			zenity.Title("wire-pod uninstaller"),
		)
	}
	os.RemoveAll(val)
	os.Exit(0)
}

func DontRunPodAtStartup() {
	key, _ := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	key.DeleteValue("wire-pod")
}
