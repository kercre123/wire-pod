package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/getlantern/systray"
	"github.com/kercre123/chipper/pkg/logger"
	"github.com/kercre123/chipper/pkg/vars"
	botsetup "github.com/kercre123/chipper/pkg/wirepod/setup"
	stt "github.com/kercre123/chipper/pkg/wirepod/stt/vosk"
	"github.com/ncruces/zenity"
	"golang.org/x/sys/windows/registry"
)

// this directory contains code which compiled a single program for end users. gui elements are implemented.

var mBoxTitle = "wire-pod"
var mBoxError = `There was an error starting wire-pod: `
var mBoxAlreadyRunning = "Wire-pod is already running. You must quit that instance before starting another one. Exiting."
var mBoxSuccess = `Wire-pod has started successfully! It is now running in the background and can be managed in the system tray.`
var mBoxIcon = "./icons/start-up-full.png"

func getNeedsSetupMsg() string {
	return `Wire-pod is now running in the background. You must set it up by heading to http://` + botsetup.GetOutboundIP().String() + `:` + vars.WebPort + ` in a browser.`
}

func main() {
	vars.Packaged = true
	conf, err := os.UserConfigDir()
	if err != nil {
		ErrMsg(err)
	}
	pidFile, err := os.ReadFile(conf + "/runningPID")
	if err == nil {
		pid, _ := strconv.Atoi(string(pidFile))
		_, err := os.FindProcess(pid)
		if err == nil {
			zenity.Error(
				"Wire-pod is already running.",
				zenity.ErrorIcon,
				zenity.Title(mBoxTitle),
			)
			os.Exit(1)
		}
	}
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\wire-pod`, registry.WRITE|registry.READ)
	if err != nil {
		fmt.Println("Error reading from registry: " + err.Error())
		return
	}
	defer k.Close()

	pidPre, _, err := k.GetIntegerValue("LastRunningPID")
	if err == nil {
		fmt.Println(int(pidPre))
		_, err := os.FindProcess(int(pidPre))
		if err == nil || errors.Is(err, os.ErrPermission) {
			zenity.Error(
				"Wire-pod is already running.",
				zenity.ErrorIcon,
				zenity.Title(mBoxTitle),
			)
			os.Exit(1)
		}
	}

	err = k.SetQWordValue("LastRunningPID", uint64(os.Getpid()))
	if err != nil {
		fmt.Println("Error writing to registry: " + err.Error())
		return
	}

	val, _, err := k.GetStringValue("InstallPath")
	if err != nil {
		ErrMsg(fmt.Errorf("error getting value from the registry: " + err.Error()))
	}
	err = os.Chdir(filepath.Join(val, "chipper"))
	fmt.Println("Working directory: " + val)
	if err != nil {
		ErrMsg(fmt.Errorf("error setting directory to " + val))
	}
	systray.Run(onReady, onExit)
}

func ExitProgram(code int) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\wire-pod`, registry.WRITE|registry.READ)
	if err != nil {
		fmt.Println("Error reading from registry: " + err.Error())
		os.Exit(code)
	}
	k.DeleteValue("LastRunningPID")
	os.Exit(code)
}

func onExit() {
	ExitProgram(0)
}

func onReady() {
	// windows-specific

	os.Setenv("STT_SERVICE", "vosk")
	os.Setenv("DEBUG_LOGGING", "true")

	systrayIcon, err := os.ReadFile("./icons/start-up-24x24.ico")
	if err != nil {
		zenity.Error(
			"Error, could not load systray icon. Exiting.",
			zenity.Title(mBoxTitle),
		)
		os.Exit(1)
	}

	systray.SetIcon(systrayIcon)
	systray.SetTitle("wire-pod")
	systray.SetTooltip("wire-pod is starting...")
	mQuit := systray.AddMenuItem("Quit", "Quit wire-pod")
	mBrowse := systray.AddMenuItem("Web interface", "Open web UI")
	mConfig := systray.AddMenuItem("Config folder", "Open config folder in case you need to. The web UI should have everything you need.")

	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				zenity.Info(
					"Wire-pod will now exit.",
					zenity.Icon(mBoxIcon),
					zenity.Title(mBoxTitle),
				)
				ExitProgram(0)
			case <-mBrowse.ClickedCh:
				go openBrowser("http://" + botsetup.GetOutboundIP().String() + ":" + vars.WebPort)
			case <-mConfig.ClickedCh:
				conf, _ := os.UserConfigDir()
				go openFileExplorer(filepath.Join(conf, vars.PodName))
			}
		}
	}()

	StartFromProgramInit(stt.Init, stt.STT, stt.Name)
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		go zenity.Warning(
			"Error opening browser: "+err.Error(),
			zenity.WarningIcon,
			zenity.Title(mBoxTitle),
		)
		logger.Println(err)
	}
}

func openFileExplorer(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Start()
}
