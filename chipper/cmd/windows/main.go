package main

import (
	"image/color"
	"io"
	"net/http"
	"os"
	"strings"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/getlantern/systray"
	"github.com/kercre123/chipper/pkg/logger"
	stt "github.com/kercre123/chipper/pkg/wirepod/stt/vosk"
	"github.com/ncruces/zenity"
)

var (
	editor widget.Editor
	list   layout.List
)

var mBoxTitle = "wire-pod"
var mBoxNeedsSetupMsg = `Wire-pod is now running in the background. You must set it up by heading to "http://localhost:8080" in a browser.`
var mBoxError = `There was an error starting wire-pod: `
var mBoxAlreadyRunning = "Wire-pod is already running. You must quit that instance before starting another one. Exiting."
var mBoxExit = `Wire-pod is exiting.`
var mBoxSuccess = `Wire-pod has started successfully! It is now running in the background. It can be stopped in the system tray.`
var mBoxIcon = "./icons/start-up-full.png"

func main() {
	hcl := http.Client{}
	hcl.Timeout = 2
	resp, err := hcl.Get("http://localhost:8080/api/is_running")
	if err == nil {
		body, _ := io.ReadAll(resp.Body)
		if strings.TrimSpace(string(body)) == "true" {
			zenity.Error(mBoxAlreadyRunning,
				zenity.Title(mBoxTitle))
			os.Exit(0)
		} else {
			zenity.Error("Port 8080 is in use by another program. Close that program before starting wire-pod. Exiting.",
				zenity.Title(mBoxTitle))
			os.Exit(0)
		}
	}
	go app.Main()
	systray.Run(onReady, onExit)
}

func onExit() {
	zenity.Info(
		mBoxExit,
		zenity.Icon(mBoxIcon),
		zenity.Title(mBoxTitle),
	)
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
	mLogs := systray.AddMenuItem("View Logs", "Open the logs")

	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				zenity.Info(
					mBoxExit,
					zenity.Icon(mBoxIcon),
					zenity.Title(mBoxTitle),
				)
				os.Exit(0)
			case <-mLogs.ClickedCh:
				go ShowLogs()
			}
		}
	}()

	StartFromProgramInit(stt.Init, stt.STT, stt.Name)
}

func ShowLogs() error {
	w := app.NewWindow(app.Title("wire-pod log viewer"), app.Size(unit.Dp(600), unit.Dp(400)))
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	var ops op.Ops
	editor := new(widget.Editor)
	editor.SingleLine = false
	editor.ReadOnly = true // Make the editor read-only

	go func() {
		chann := logger.GetChan()
		for range chann {
			w.Invalidate()
		}
	}()

	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)

			// Update the editor's text with logger.TrayLogList
			editor.SetText(logger.TrayLogList)

			layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					editor := material.Editor(th, editor, "")
					editor.TextSize = unit.Sp(16)
					editor.Color = color.NRGBA{R: 127, G: 0, B: 0, A: 255}
					return editor.Layout(gtx)
				}),
			)
			e.Frame(gtx.Ops)
		}
	}
}
