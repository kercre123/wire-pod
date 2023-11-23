package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/kercre123/chipper/pkg/podonwin"
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
	RunAtStartup    bool
	AutoUpdate      bool
	Where           string
	WebPort         string
	SetHostnameEpod bool
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

func PostInstall_RestartReq(myApp fyne.App, is InstallSettings) {
	window := myApp.NewWindow("wire-pod installer")
	window.Resize(fyne.Size{Width: 600, Height: 100})
	window.SetIcon(icon)
	window.CenterOnScreen()
	window.SetMaster()

	finished := widget.NewCard("wire-pod installer", "wire-pod has finished installing!", container.NewWithoutLayout())

	tellRestart := widget.NewRichTextWithText("You must restart your computer before you start wire-pod, otherwise Vector won't be able to communicate with it.")

	var clarifyWhenPodStarts *widget.RichText
	if is.RunAtStartup {
		clarifyWhenPodStarts = widget.NewRichTextWithText("wire-pod will automatically start after you reboot.")
	} else {
		clarifyWhenPodStarts = widget.NewRichTextWithText("You must manually run wire-pod after restart because you chose for it not to automatically start at login. This can be changed in wire-pod's 'About' menu.")
	}

	clarifyWhenPodStarts.Wrapping = fyne.TextWrapWord

	rebootNowButton := widget.NewButton("Reboot Now", func() {
		RebootSystem()
	})

	rebootLaterButton := widget.NewButton("Reboot Later", func() {
		podonwin.UpdateRegistryValueString(podonwin.SoftwareKey, "RestartNeeded", "true")
		os.Exit(0)
	})

	window.SetContent(container.NewVBox(
		finished,
		tellRestart,
		clarifyWhenPodStarts,
		widget.NewSeparator(),
		rebootNowButton,
		rebootLaterButton,
	))

	window.Show()
}

func PostInstall(myApp fyne.App, is InstallSettings) {
	var shouldStartPod bool = true
	window := myApp.NewWindow("wire-pod installer")
	window.Resize(fyne.Size{Width: 600, Height: 100})
	window.SetIcon(icon)
	window.CenterOnScreen()
	window.SetMaster()

	finished := widget.NewCard("wire-pod installer", "wire-pod has finished installing!", container.NewWithoutLayout())

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
	window.SetMaster()
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
		zenity.Error(
			"Error installing wire-pod, will revert installation and quit: "+err.Error(),
			zenity.ErrorIcon,
			zenity.Title("wire-pod installer"),
		)
		podonwin.DeleteEverythingFromRegistry()
		os.Exit(1)
	}
	window.Hide()
	if is.SetHostnameEpod {
		PostInstall_RestartReq(myApp, is)
	} else {
		PostInstall(myApp, is)
	}
}

func GetSetNameEscapepod(myApp fyne.App, is InstallSettings) {
	window := myApp.NewWindow("wire-pod installer")
	window.SetIcon(icon)
	window.CenterOnScreen()
	window.SetMaster()
	window.Resize(fyne.NewSize(673, 275))
	content := container.NewVBox()

	//content.Resize(fyne.NewSize(600, 200))
	prefInfo := widget.NewCard("wire-pod installer",
		"Your computer's hostname is not `escapepod`.",
		container.NewWithoutLayout())
	content.Add(prefInfo)

	morePrefInfo := widget.NewRichTextWithText("This means regular Vector robots will not be able to communicate with wire-pod unless you have a special network configuration or set it to `escapepod`.")
	morePrefInfo.Wrapping = fyne.TextWrapWord
	content.Add(morePrefInfo)

	question := widget.NewRichTextWithText("Would you like the installer to set the computer's hostname to `escapepod` during installation?")
	question.Wrapping = fyne.TextWrapWord
	content.Add(question)

	exp := widget.NewRichTextWithText("This will require a computer restart once installation has completed.")
	question.Wrapping = fyne.TextWrapWord
	content.Add(exp)

	content.Add(widget.NewSeparator())

	yesButton := widget.NewButton("Yes", func() {
		is.SetHostnameEpod = true
		window.Hide()
		DoInstall(myApp, is)
	})
	noButton := widget.NewButton("No", func() {
		is.SetHostnameEpod = false
		podonwin.DeleteRegistryValue(podonwin.SoftwareKey, "RestartNeeded")
		window.Hide()
		DoInstall(myApp, is)
	})
	buttonContainer := container.New(layout.NewGridLayout(2), yesButton, noButton)
	content.Add(buttonContainer)
	window.SetContent(content)
	window.Show()
	go func() {
		time.Sleep(time.Millisecond * 500)
		window.Resize(fyne.NewSize(673, 275))
	}()
}

func ValidateWebPort(port string) bool {
	i, err := strconv.Atoi(port)
	if err == nil {
		if i < 1000 || i > 65353 {
			return false
		}
		return true
	}
	return false
}

func GetPreferences(myApp fyne.App) {
	var is InstallSettings
	window := myApp.NewWindow("wire-pod installer")
	window.SetIcon(icon)
	window.Resize(fyne.Size{Width: 600, Height: 200})
	window.CenterOnScreen()
	window.SetMaster()
	launchOnStartup := widget.NewCheck("Automatically launch wire-pod after login?", func(checked bool) {
		is.RunAtStartup = checked
	})

	launchOnStartup.SetChecked(true)
	is.RunAtStartup = true

	installDir := widget.NewEntry()
	installDir.SetText(DefaultInstallationDirectory)
	installDir.Disable()

	webPort := widget.NewEntry()
	webPort.SetText("8080")

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
		is.WebPort = webPort.Text
		if !ValidateInstallDirectory(is.Where) {
			zenity.Warning(
				"The directory you have provided ("+is.Where+") is invalid. Please provide a valid path or use the default one.",
				zenity.WarningIcon,
				zenity.Title("wire-pod installer"),
			)
		} else if !ValidateWebPort(is.WebPort) {
			zenity.Warning(
				"The web port you have provided ("+is.WebPort+") is invalid. It must be an integer between 1000-65353.",
				zenity.WarningIcon,
				zenity.Title("wire-pod installer"),
			)
		} else {
			window.Hide()
			hn, _ := os.Hostname()
			if hn == "escapepod" {
				DoInstall(myApp, is)
			} else {
				GetSetNameEscapepod(myApp, is)
			}
		}
	})

	firstCard := widget.NewCard("wire-pod installer", "This program will install wire-pod with the following settings.", container.NewWithoutLayout())

	window.SetContent(container.NewVBox(
		firstCard,
		widget.NewSeparator(),
		launchOnStartup,
		widget.NewSeparator(),
		widget.NewRichTextWithText("Installation Directory"),
		installDir,
		selectDirButton,
		widget.NewSeparator(),
		widget.NewRichTextWithText("Web Configurator Port"),
		webPort,
		widget.NewSeparator(),
		nextButton,
	))
	window.Show()
}

func StopWirePodIfRunning() {
	podPid, err := os.ReadFile(filepath.Join(os.TempDir(), "/wirepodrunningPID"))
	if err == nil {
		pid, _ := strconv.Atoi(string(podPid))
		if is, _ := podonwin.IsProcessRunning(pid); is {
			podProcess, err := os.FindProcess(pid)
			if err == nil {
				fmt.Println("Stopping wire-pod")
				podProcess.Kill()
				podProcess.Wait()
				fmt.Println("Stopped")
				return
			}
		}
	}
	StopWirePod_Registry()
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
	fmt.Println("Initing registry")
	podonwin.Init()
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

type Release struct {
	TagName string `json:"tag_name"`
}

func GetLatestReleaseTag(owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return "", err
	}

	return release.TagName, nil
}
