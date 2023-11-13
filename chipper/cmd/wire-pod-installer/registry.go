package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

func UpdateRegistry(is InstallSettings) {
	keyPath := `Software\Microsoft\Windows\CurrentVersion\Uninstall\wire-pod`
	appName := "wire-pod"
	displayIcon := filepath.Join(is.Where, `\chipper\icons\ico\pod256x256.ico`)
	displayVersion := "1.0.0"
	publisher := "github.com/kercre123"
	uninstallString := filepath.Join(is.Where, `\uninstall.exe`)
	installLocation := filepath.Join(is.Where, `\chipper\chipper.exe`)
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		k, _, err = registry.CreateKey(registry.LOCAL_MACHINE, keyPath, registry.ALL_ACCESS)
		if err != nil {
			fmt.Printf("Error creating registry key: %v\n", err)
			return
		}
	}
	defer k.Close()

	err = k.SetStringValue("DisplayName", appName)
	if err != nil {
		fmt.Printf("Error setting DisplayName: %v\n", err)
		return
	}
	k.SetStringValue("DisplayIcon", displayIcon)
	k.SetStringValue("DisplayVersion", displayVersion)
	k.SetStringValue("Publisher", publisher)
	k.SetStringValue("UninstallString", uninstallString)
	k.SetStringValue("InstallLocation", installLocation)
	k.SetStringValue("InstallPath", is.Where)
	fmt.Println("Registry entries successfully created")
}

func RunPodAtStartup(is InstallSettings) {
	key, _ := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	key.SetStringValue("wire-pod", filepath.Join(is.Where, "\\chipper\\chipper.exe"))
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
