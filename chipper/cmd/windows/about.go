package main

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/kercre123/wire-pod/chipper/pkg/podonwin"
)

var AboutWindow fyne.Window
var WindowDefined bool

var FakeWindow fyne.Window

func RunPodAtStartup(installPath string) {
	cmd := fmt.Sprintf(`cmd.exe /C start "" "` + filepath.Join(installPath, "chipper\\chipper.exe") + `" -d`)
	podonwin.UpdateRegistryValueString(podonwin.StartupRunKey, "wire-pod", cmd)
}

func DontRunPodAtStartup(installPath string) {
	podonwin.DeleteRegistryValue(podonwin.StartupRunKey, "wire-pod")
}

func IsPodRunningAtStartup() bool {
	_, err := podonwin.GetRegistryValueString(podonwin.StartupRunKey, "wire-pod")
	if err != nil {
		return false
	}
	return true
}

func GetPodVersion() string {
	podVersion, err := podonwin.GetRegistryValueString(podonwin.SoftwareKey, "PodVersion")
	if err != nil {
		return "v0.0.0"
	}
	return podVersion
}

func DefineAboutWindow(myApp fyne.App) {
	AboutWindow = myApp.NewWindow("wire-pod")
	AboutWindow.Resize(fyne.Size{Width: 400, Height: 100})
	AboutWindow.CenterOnScreen()
	icon, err := os.ReadFile(mBoxIcon)
	var iconRes *fyne.StaticResource
	if err == nil {
		iconRes = fyne.NewStaticResource("podIcon", icon)
		AboutWindow.SetIcon(iconRes)
	} else {
		fmt.Println("error loading icon: " + fmt.Sprint(err))
	}
	card := widget.NewCard("wire-pod", "wire-pod is an Escape Pod alternative which is able to get any Anki/DDL Vector robot setup and working with voice commands.",
		container.NewWithoutLayout())

	version := widget.NewRichTextWithText("Version: " + GetPodVersion())

	runStartup := widget.NewCheck("Run wire-pod when user logs in?", func(checked bool) {
		if checked {
			RunPodAtStartup(InstallPath)
		} else {
			DontRunPodAtStartup(InstallPath)
		}
	})

	runStartup.SetChecked(IsPodRunningAtStartup())

	exitButton := widget.NewButton("Close", func() {
		AboutWindow.Hide()
	})

	AboutWindow.SetContent(container.NewVBox(
		card,
		version,
		runStartup,
		widget.NewSeparator(),
		exitButton,
	))

}

func ShowAbout(myApp fyne.App) {
	if !WindowDefined {
		FakeWindow = myApp.NewWindow("a")
		FakeWindow.Hide()
		DefineAboutWindow(myApp)
	}

	FakeWindow.SetMaster()
	AboutWindow.Show()
}
