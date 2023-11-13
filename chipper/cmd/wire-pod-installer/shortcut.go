package main

import (
	"path/filepath"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

func CreateShortcut(is InstallSettings) {
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		panic(err)
	}
	defer unknown.Release()

	wshell, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		panic(err)
	}
	defer wshell.Release()

	cs, err := oleutil.CallMethod(wshell, "CreateShortcut", "C:\\ProgramData\\Microsoft\\Windows\\Start Menu\\Programs\\wire-pod.lnk")
	if err != nil {
		panic(err)
	}
	shortcut := cs.ToIDispatch()
	defer shortcut.Release()

	oleutil.PutProperty(shortcut, "TargetPath", filepath.Join(is.Where, "\\chipper\\chipper.exe"))
	// Set other properties as needed
	oleutil.CallMethod(shortcut, "Save")
}
