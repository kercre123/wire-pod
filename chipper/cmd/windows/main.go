package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/digital-dream-labs/hugh/log"
	"github.com/getlantern/systray"
	"github.com/kercre123/chipper/pkg/logger"
	"github.com/kercre123/chipper/pkg/wirepod/sdkapp"
	botsetup "github.com/kercre123/chipper/pkg/wirepod/setup"
	stt "github.com/kercre123/chipper/pkg/wirepod/stt/vosk"
	"github.com/ncruces/zenity"
)

var mBoxTitle = "wire-pod"
var mBoxError = `There was an error starting wire-pod: `
var mBoxAlreadyRunning = "Wire-pod is already running. You must quit that instance before starting another one. Exiting."
var mBoxSuccess = `Wire-pod has started successfully! It is now running in the background. It can be stopped in the system tray.`
var mBoxIcon = "./icons/start-up-full.png"

func getNeedsSetupMsg() string {
	return `Wire-pod is now running in the background. You must set it up by heading to http://` + botsetup.GetOutboundIP().String() + `:` + sdkapp.WebPort + ` in a browser.`
}

func main() {
	if os.Getenv("WEBSERVER_PORT") != "" {
		if _, err := strconv.Atoi(os.Getenv("WEBSERVER_PORT")); err == nil {
			sdkapp.WebPort = os.Getenv("WEBSERVER_PORT")
		} else {
			logger.Println("WEBSERVER_PORT contains letters, using default of 8080")
			sdkapp.WebPort = "8080"
		}
	} else {
		sdkapp.WebPort = "8080"
	}
	resp, err := http.Get("http://" + botsetup.GetOutboundIP().String() + ":" + sdkapp.WebPort + "/api/is_running")
	if err == nil {
		body, _ := io.ReadAll(resp.Body)
		if strings.TrimSpace(string(body)) == "true" {
			zenity.Error(mBoxAlreadyRunning,
				zenity.ErrorIcon,
				zenity.Title(mBoxTitle))
			os.Exit(1)
		} else {
			zenity.Error("Port "+sdkapp.WebPort+" is in use by another program. Close that program before starting wire-pod. Exiting.",
				zenity.ErrorIcon,
				zenity.Title(mBoxTitle))
			os.Exit(1)
		}
	}
	systray.Run(onReady, onExit)
}

func onExit() {
	os.Exit(0)
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
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	mBrowse := systray.AddMenuItem("Web interface", "Open web UI")

	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				zenity.Info(
					"Wire-pod will now exit.",
					zenity.Icon(mBoxIcon),
					zenity.Title(mBoxTitle),
				)
				os.Exit(0)
			case <-mBrowse.ClickedCh:
				go openBrowser("http://" + botsetup.GetOutboundIP().String() + ":" + sdkapp.WebPort)
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
		log.Fatal(err)
	}
}
