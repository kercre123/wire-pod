package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kercre123/chipper/pkg/podreg"
	"golang.org/x/sys/windows/registry"
)

var GitHubTag string

func UpdateRegistry(is InstallSettings) {
	UpdateUninstallRegistry(is)
	UpdateSoftwareRegistry(is)
}

func DeleteAnyOtherInstallation() {
	instPath, err := podreg.GetRegistryValueString(podreg.UninstallKey, "InstallPath")
	if err != nil {
		val, err := podreg.GetRegistryValueString(podreg.SoftwareKey, "InstallPath")
		if err != nil {
			return
		}
		fmt.Println("Running uninstaller")
		cmd := exec.Command(val)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "RUN_DISCRETE=true")
		cmd.Run()
	} else {
		os.RemoveAll(instPath)
	}
}

func UpdateUninstallRegistry(is InstallSettings) {
	appName := "wire-pod"
	displayIcon := filepath.Join(is.Where, `\chipper\icons\ico\pod256x256.ico`)
	displayVersion := GitHubTag
	publisher := "github.com/kercre123"
	uninstallString := filepath.Join(is.Where, `\uninstall.exe`)
	installLocation := filepath.Join(is.Where, `\chipper\chipper.exe`)
	err := podreg.UpdateRegistryValueString(podreg.UninstallKey, "DisplayName", appName)
	if err != nil {
		// if this one works, the rest will
		fmt.Printf("Error setting DisplayName: %v\n", err)
		return
	}
	podreg.UpdateRegistryValueString(podreg.UninstallKey, "DisplayIcon", displayIcon)
	podreg.UpdateRegistryValueString(podreg.UninstallKey, "DisplayVersion", displayVersion)
	podreg.UpdateRegistryValueString(podreg.UninstallKey, "Publisher", publisher)
	podreg.UpdateRegistryValueString(podreg.UninstallKey, "UninstallString", uninstallString)
	podreg.UpdateRegistryValueString(podreg.UninstallKey, "InstallLocation", installLocation)
	fmt.Println("Registry entries successfully created")
}

func UpdateSoftwareRegistry(is InstallSettings) {
	err := podreg.UpdateRegistryValueString(podreg.SoftwareKey, "InstallPath", is.Where)
	if err != nil {
		fmt.Printf("Error setting registry key InstallPath: %v\n", err)
		return
	}
	podreg.UpdateRegistryValueString(podreg.SoftwareKey, "PodVersion", GitHubTag)
	podreg.UpdateRegistryValueString(podreg.SoftwareKey, "WebPort", is.WebPort)
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

func RunPodAtStartup(is InstallSettings) {
	key, _ := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	cmd := fmt.Sprintf(`cmd.exe /C start "" "` + filepath.Join(is.Where, "chipper\\chipper.exe") + `" -d`)
	key.SetStringValue("wire-pod", cmd)
}

func AllowThroughFirewall(is InstallSettings) {
	cmdStr := fmt.Sprintf("netsh advfirewall firewall add rule name=\"wire-pod\" dir=in action=allow program=\"%s\\chipper\\chipper.exe\" enable=yes", is.Where)
	fmt.Println("Executing command:", cmdStr)
	cmd := exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
		"name=wire-pod",
		"dir=in",
		"action=allow",
		"profile=any",
		"program="+is.Where+"\\chipper\\chipper.exe",
		"enable=yes")

	out, err := cmd.Output()
	if err != nil {
		fmt.Println(string(out))
		log.Fatalf("Failed to execute command in: %s", err)
	}
	cmd = exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
		"name=wire-pod",
		"dir=out",
		"action=allow",
		"profile=any",
		"program="+is.Where+"\\chipper\\chipper.exe",
		"enable=yes")

	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to execute command out: %s", err)
	}

	log.Println("Firewall rule added successfully.")
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
