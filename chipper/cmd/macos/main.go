package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/getlantern/systray"
	"github.com/kercre123/chipper/pkg/logger"
	"github.com/kercre123/chipper/pkg/vars"
	botsetup "github.com/kercre123/chipper/pkg/wirepod/setup"
	stt "github.com/kercre123/chipper/pkg/wirepod/stt/vosk"
	// stt "github.com/kercre123/chipper/pkg/wirepod/stt/whisper.cpp"
	"github.com/ncruces/zenity"
)

var mBoxTitle = "wire-pod"
var mBoxError = "There was an error starting wire-pod:"
var mBoxAlreadyRunning = "Wire-pod is already running. You must quit that instance before starting another one. Exiting."
var mBoxSuccess = "Wire-pod has started successfully! It is now running in the background and can be managed in the system tray."

func mBoxIcon() string {
	if runtime.GOOS == "windows" {
		return "./icons/start-up-full.png"
	} else {
		appPath, _ := os.Executable()
		return filepath.Dir(appPath) + "/../Resources/start-up-full.png"
	}
}

func getNeedsSetupMsg() string {
	return `wire-pod is now running in the background. You must set it up by heading to http://` + botsetup.GetOutboundIP().String() + `:` + vars.WebPort + ` in a browser.`
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			conf, _ := os.UserConfigDir()
			path := filepath.Join(conf, "dump.txt")
			os.WriteFile(path, []byte(fmt.Sprint(r)), 0777)
			fmt.Printf("panic!: %v\n", r)
			zenity.Error("wire-pod has crashed. dump located in " + path + ". Exiting.",
				zenity.ErrorIcon,
				zenity.Title("wire-pod crash :("))
			ExitProgram(1)
		}
	}()
	vars.Packaged = true
	systray.Run(onReady, onExit)
}

func ExitProgram(code int) {
	os.Exit(code)
}

func onExit() {
	ExitProgram(0)
}

func onReady() {
	appPath, _ := os.Executable()
	os.Setenv("STT_SERVICE", "vosk")
	os.Setenv("DEBUG_LOGGING", "true")
	// os.Setenv("GGML_METAL_PATH_RESOURCES", filepath.Dir(appPath) + "/../Frameworks/chipper/whisper.cpp")

	systrayIcon, err := os.ReadFile(filepath.Dir(appPath) + "/../Resources/start-up-24x24.ico")
	if err != nil {
		zenity.Error(
			"Error, could not load systray icon. Exiting.",
			zenity.Title(mBoxTitle),
		)
		os.Exit(1)
	}

	systray.SetIcon(systrayIcon)
	systray.SetTooltip("wire-pod is starting...")
	mAbout := systray.AddMenuItem("wire-pod", "")
	systray.AddSeparator()
	mBrowse := systray.AddMenuItem("Web Interface", "Open web UI")
	mConfig := systray.AddMenuItem("Config Folder", "Open config folder in case you need to. The web UI should have everything you need.")
	mStartup := systray.AddMenuItem("Run at startup", "")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit wire-pod")

	if IsPodRunningAtStartup() {
		mStartup.Check()
	} else {
		mStartup.Uncheck()
	}

	go func() {
		for {
			select {
			case <-mAbout.ClickedCh:
				zenity.Info("wire-pod is an Escape Pod alternative which is able to get any Anki/DDL Vector robot setup and working with voice commands.",
				zenity.Icon(mBoxIcon()),
				zenity.Title("wire-pod"))

			case <-mBrowse.ClickedCh:
				go openBrowser("http://" + botsetup.GetOutboundIP().String() + ":" + vars.WebPort)

			case <-mConfig.ClickedCh:
				conf, _ := os.UserConfigDir()
				go openFileExplorer(filepath.Join(conf, vars.PodName))

			case <-mStartup.ClickedCh:
				if mStartup.Checked() {
					mStartup.Uncheck()
					DontRunPodAtStartup()
				} else {
					mStartup.Check()
					RunPodAtStartup()
				}

			case <-mQuit.ClickedCh:
				ExitProgram(0)
			}
		}
	}()

	StartFromProgramInit(stt.Init, stt.STT, stt.Name)
}

func RunPodAtStartup() {
	homeDir, _ := os.UserHomeDir()
	launchAgentsDir := filepath.Join(homeDir, "/Library/LaunchAgents")
	if _, err := os.Stat(launchAgentsDir); os.IsNotExist(err) {
		os.Mkdir(launchAgentsDir, 0777)
	}
	executable, _ := os.Executable()
	err := os.WriteFile(launchAgentsDir + "/wire-pod.agent.plist", []byte(`
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>wire-pod.agent</string>
	<key>ProgramArguments</key>
	<array>
		<string>` + executable +  `</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
</dict>
</plist>`), 0777)
	if err != nil {
		go zenity.Error("Error enabling run at startup: " + err.Error(), zenity.Title(mBoxTitle))
	}
}

func DontRunPodAtStartup() {
	homeDir, _ := os.UserHomeDir()
	err := os.Remove(filepath.Join(homeDir, "/Library/LaunchAgents/wire-pod.agent.plist"))
	if err != nil {
		go zenity.Error("Error disabling run at startup: " + err.Error(), zenity.Title(mBoxTitle))
	}
}

func IsPodRunningAtStartup() bool {
	homeDir, _ := os.UserHomeDir()
	return IfFileExist(filepath.Join(homeDir, "/Library/LaunchAgents/wire-pod.agent.plist"))
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
