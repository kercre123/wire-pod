package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/getlantern/systray"
	"github.com/kercre123/chipper/pkg/logger"
	"github.com/kercre123/chipper/pkg/podonwin"
	"github.com/kercre123/chipper/pkg/vars"
	botsetup "github.com/kercre123/chipper/pkg/wirepod/setup"
	stt "github.com/kercre123/chipper/pkg/wirepod/stt/vosk"
	"github.com/ncruces/zenity"
)

// this directory contains code which compiled a single program for end users. gui elements are implemented.

var fyneApp fyne.App

var InstallPath string

var mBoxTitle = "wire-pod"
var mBoxError = `There was an error starting wire-pod: `
var mBoxAlreadyRunning = "Wire-pod is already running. You must quit that instance before starting another one. Exiting."
var mBoxSuccess = `Wire-pod has started successfully! It is now running in the background and can be managed in the system tray.`
var mBoxIcon = "./icons/start-up-full.png"

func getNeedsSetupMsg() string {
	return `Wire-pod is now running in the background. You must set it up by heading to http://` + botsetup.GetOutboundIP().String() + `:` + vars.WebPort + ` in a browser.`
}

func checkIfRestartNeeded() bool {
	host, _ := os.Hostname()
	val, err := podonwin.GetRegistryValueString(podonwin.SoftwareKey, "RestartNeeded")
	if err != nil {
		return false
	}
	if val == "true" && host != "escapepod" {
		return true
	} else if val == "true" && host == "escapepod" {
		podonwin.DeleteRegistryValue(podonwin.SoftwareKey, "RestartNeeded")
		return false
	}
	return false
}

func main() {

	defer func() {
		if r := recover(); r != nil {
			conf, _ := os.UserConfigDir()
			os.WriteFile(filepath.Join(conf, "dump.txt"), []byte(fmt.Sprint(r)), 0777)
			fmt.Printf("panic!: %v\n", r)
			zenity.Error("wire-pod has crashed. dump located in %APPDATA%/wire-pod/dump.txt. exiting",
				zenity.ErrorIcon,
				zenity.Title("wire-pod crash :("))
			ExitProgram(1)
		}
	}()

	podonwin.Init()
	if checkIfRestartNeeded() {
		zenity.Error(
			"You must restart your computer before starting wire-pod.",
			zenity.ErrorIcon,
			zenity.Title(mBoxTitle),
		)
		os.Exit(1)
	}
	vars.Packaged = true
	conf, err := os.UserConfigDir()
	if err != nil {
		ErrMsg(err)
	}
	pidFile, err := os.ReadFile(conf + "/runningPID")
	if err == nil {
		pid, _ := strconv.Atoi(string(pidFile))
		if is, _ := podonwin.IsProcessRunning(pid); is {
			zenity.Error(
				"Wire-pod is already running.",
				zenity.ErrorIcon,
				zenity.Title(mBoxTitle),
			)
			os.Exit(1)
		}
	}
	pid, err := podonwin.GetRegistryValueInt(podonwin.SoftwareKey, "LastRunningPID")
	if err == nil {
		if is, _ := podonwin.IsProcessRunning(pid); is {
			zenity.Error(
				"Wire-pod is already running.",
				zenity.ErrorIcon,
				zenity.Title(mBoxTitle),
			)
			os.Exit(1)
		}
	}

	err = podonwin.UpdateRegistryValueInt(podonwin.SoftwareKey, "LastRunningPID", os.Getpid())
	if err != nil {
		ErrMsg(fmt.Errorf("Error writing to the registry (lastrunningpid): " + err.Error()))
	}

	val, err := podonwin.GetRegistryValueString(podonwin.SoftwareKey, "InstallPath")
	if err != nil {
		ErrMsg(fmt.Errorf("error getting InstallPath value from the registry: " + err.Error()))
	}
	InstallPath = val
	err = os.Chdir(filepath.Join(val, "chipper"))
	fmt.Println("Working directory: " + val)
	if err != nil {
		ErrMsg(fmt.Errorf("error setting runtime directory to " + val))
	}

	webPort, err := podonwin.GetRegistryValueString(podonwin.SoftwareKey, "WebPort")
	if err == nil {
		os.Setenv("WEBSERVER_PORT", webPort)
	}

	go systray.Run(onReady, onExit)
	// for the about window to work
	// fine since everything uses `os.Exit()` to exit the program
	fyneApp = app.New()
	fyneApp.Run()
}

func ExitProgram(code int) {
	systray.Quit()
	podonwin.DeleteRegistryValue(podonwin.SoftwareKey, "LastRunningPID")
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
	mAbout := systray.AddMenuItem("About", "About wire-pod")

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
			case <-mAbout.ClickedCh:
				ShowAbout(fyneApp)
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
