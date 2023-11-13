package main

import (
	"embed"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/ncruces/zenity"
)

//go:embed ico
var iconData embed.FS

var DefaultInstallationDirectory string = "C:\\Program Files\\wire-pod"

var icon *fyne.StaticResource

type InstallSettings struct {
	RunAtStartup bool
	AutoUpdate   bool
	Where        string
}

func PostInstall(myApp fyne.App) {
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
			fmt.Println("Would start wire-pod here")
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

func DoInstall(myApp fyne.App) {
	var barStatus float64
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

	window.Show()
	i := 0
	for {
		time.Sleep(time.Second / 2)
		barStatus = barStatus + 0.1
		card.SetSubTitle(fmt.Sprint(i))
		card.Refresh()
		bar.SetValue(barStatus)
		bar.Refresh()
		i = i + 1
		if i == 10 {
			break
		}
	}
	window.Hide()
	PostInstall(myApp)
}

func GetPreferences(myApp fyne.App) {
	var is InstallSettings
	window := myApp.NewWindow("wire-pod installer")
	window.SetIcon(icon)
	window.Resize(fyne.Size{Width: 600, Height: 200})
	window.CenterOnScreen()
	launchOnStartup := widget.NewCheck("Launch on startup?", func(checked bool) {
		is.RunAtStartup = checked
	})
	autoUpdate := widget.NewCheck("Auto-update?", func(checked bool) {
		is.AutoUpdate = checked
	})

	installDir := widget.NewEntry()
	installDir.SetText(DefaultInstallationDirectory)

	selectDirButton := widget.NewButton("Select Directory", func() {
		dlg := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if uri != nil {
				installDir.SetText(uri.Path())
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
			DoInstall(myApp)
		}
	})

	// Add widgets to the window
	window.SetContent(container.NewVBox(
		widget.NewRichText(&widget.TextSegment{
			Text: "This program will install wire-pod with the following settings.",
		}),
		launchOnStartup,
		autoUpdate,
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
	iconBytes, err := iconData.ReadFile("ico/pod.png")
	if err != nil {
		fmt.Println(err)
	}
	icon = fyne.NewStaticResource("icon", iconBytes)
	myApp := app.New()
	GetPreferences(myApp)
	myApp.Run()
	os.Exit(0)
}

// func runElevated() {
// 	drv, err := os.Open("\\\\.\\PHYSICALDRIVE0")
// 	if err != nil {
// 		for _, arg := range os.Args {ico
// 			if strings.Contains(arg, "-esc") {
// 				fmt.Println("Privledge escalation failed")
// 				os.Exit(0)
// 			}
// 		}
// 		verb := "runas"
// 		exe, _ := os.Executable()
// 		cwd, _ := os.Getwd()
// 		args := strings.Join(os.Args[1:], "-esc")

// 		verbPtr, _ := syscall.UTF16PtrFromString(verb)
// 		exePtr, _ := syscall.UTF16PtrFromString(exe)
// 		cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
// 		argPtr, _ := syscall.UTF16PtrFromString(args)

// 		var showCmd int32 = 1 //SW_NORMAL

// 		err := windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
// 		if err != nil {
// 			fmt.Println(err)
// 			os.Exit(0)
// 		}
// 	}
// 	drv.Close()
// }
