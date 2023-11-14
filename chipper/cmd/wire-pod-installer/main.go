package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/ncruces/zenity"
)

//go:embed ico
var iconData embed.FS

var amd64podURL string = "https://github.com/kercre123/wire-pod/releases/latest/download/wire-pod-win-amd64.zip"

//var amd64podURL string = "http://192.168.1.2:82/wire-pod-win-amd64.zip"

var DefaultInstallationDirectory string = "C:\\Program Files\\wire-pod"

var icon *fyne.StaticResource

var installerStatusUpdate chan string
var installerBarUpdate chan float64

type InstallSettings struct {
	RunAtStartup bool
	AutoUpdate   bool
	Where        string
}

func UpdateInstallStatus(status string) {
	select {
	case installerStatusUpdate <- status:
	default:
	}
}

func UpdateInstallBar(status float64) {
	select {
	case installerBarUpdate <- status / 100:
	default:
	}
}

func GetBarChan() chan float64 {
	return installerBarUpdate
}

func GetStatusChan() chan string {
	return installerStatusUpdate
}

func ExecuteDetached(program string) error {
	cmd := exec.Command(program)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
	return cmd.Start()
}

func PostInstall(myApp fyne.App, is InstallSettings) {
	var shouldStartPod bool = true
	window := myApp.NewWindow("wire-pod installer")
	window.Resize(fyne.Size{Width: 600, Height: 100})
	window.SetIcon(icon)
	window.CenterOnScreen()

	finished := widget.NewRichText(&widget.TextSegment{
		Text: "wire-pod has finished installing!",
	})

	startpod := widget.NewCheck("Start wire-pod after exit?", func(checked bool) {
		shouldStartPod = checked
	})

	startpod.SetChecked(true)

	exitButton := widget.NewButton("Exit", func() {
		if shouldStartPod {
			window.Hide()
			ExecuteDetached(filepath.Join(is.Where, "chipper/chipper.exe"))
		}
		os.Exit(0)
	})

	window.SetContent(container.NewVBox(
		finished,
		startpod,
		exitButton,
	))

	window.Show()
}

func DoInstall(myApp fyne.App, is InstallSettings) {
	window := myApp.NewWindow("wire-pod installer")
	window.Resize(fyne.Size{Width: 600, Height: 100})
	window.CenterOnScreen()
	window.SetIcon(icon)
	bar := widget.NewProgressBar()
	card := widget.NewCard("Installing wire-pod", "Starting installation...",
		container.NewWithoutLayout())

	window.SetContent(container.NewVBox(
		card,
		bar,
	))

	barChan := GetBarChan()
	statusChan := GetStatusChan()

	window.Show()
	go func() {
		for val := range barChan {
			bar.SetValue(val)
			card.Refresh()
		}
	}()
	go func() {
		for val := range statusChan {
			card.SetSubTitle(val)
			card.Refresh()
		}
	}()
	err := InstallWirePod(is)
	if err != nil {
		fmt.Println("error installing wire-pod: " + err.Error())
	}
	window.Hide()
	PostInstall(myApp, is)
}

func GetPreferences(myApp fyne.App) {
	var is InstallSettings
	window := myApp.NewWindow("wire-pod installer")
	window.SetIcon(icon)
	window.Resize(fyne.Size{Width: 600, Height: 200})
	window.CenterOnScreen()
	launchOnStartup := widget.NewCheck("Automatically launch wire-pod after login?", func(checked bool) {
		is.RunAtStartup = checked
	})

	launchOnStartup.SetChecked(true)
	is.RunAtStartup = true

	installDir := widget.NewEntry()
	installDir.SetText(DefaultInstallationDirectory)

	selectDirButton := widget.NewButton("Select Directory", func() {
		dlg := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if uri != nil {
				installDir.SetText(filepath.Join(uri.Path(), "wire-pod"))
			}
		}, window)
		dlg.Show()
	})

	nextButton := widget.NewButton("Next", func() {
		is.Where = installDir.Text
		if !ValidateInstallDirectory(is.Where) {
			zenity.Warning(
				"The directory you have provided ("+is.Where+") is invalid. Please provide a valid path or use the default one.",
				zenity.WarningIcon,
				zenity.Title("wire-pod installer"),
			)
		} else {
			window.Hide()
			DoInstall(myApp, is)
		}
	})

	// Add widgets to the window
	window.SetContent(container.NewVBox(
		widget.NewRichText(&widget.TextSegment{
			Text: "This program will install wire-pod with the following settings.",
		}),
		launchOnStartup,
		widget.NewSeparator(),
		widget.NewRichText(&widget.TextSegment{
			Text: "Installation Directory",
		}),
		installDir,
		selectDirButton,
		widget.NewSeparator(),
		nextButton,
	))
	window.Show()
}

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

func ValidateInstallDirectory(dir string) bool {
	var dirWithoutLast string
	splitDir := strings.Split(dir, "\\")
	dirWithoutLast = splitDir[0]
	for in, str := range splitDir {
		if in == len(splitDir)-1 || in == 0 {
			continue
		}
		dirWithoutLast = dirWithoutLast + "\\" + str
	}
	if statted, err := os.Stat(dirWithoutLast); err == nil {
		if statted.IsDir() {
			return true
		}
	}
	return false
}

func main() {
	if !CheckIfElevated() {
		fmt.Println("installer must be run as administrator")
		os.Exit(0)
	}
	fmt.Println("Getting tag from GitHub")
	tag, err := GetLatestReleaseTag("kercre123", "wire-pod")
	if err != nil {
		fmt.Println("Error getting: " + err.Error())
		zenity.Error(
			"Error getting latest GitHub tag from GitHub, exiting: "+err.Error(),
			zenity.ErrorIcon,
			zenity.Title("wire-pod installer"),
		)
		os.Exit(0)
	}
	GitHubTag = tag
	fmt.Println(tag)
	iconBytes, err := iconData.ReadFile("ico/pod.png")
	if err != nil {
		fmt.Println(err)
	}
	installerBarUpdate = make(chan float64)
	installerStatusUpdate = make(chan string)
	icon = fyne.NewStaticResource("icon", iconBytes)
	myApp := app.New()
	GetPreferences(myApp)
	myApp.Run()
	os.Exit(0)
}

func CheckIfElevated() bool {
	drv, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		return false
	}
	drv.Close()
	return true
}
