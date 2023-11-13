# wire-pod on Windows

testing grounds for running wire-pod on Windows.

Status:

	-	build.sh creates a .zip which contains chipper.exe and its required DLLs
	-	this .zip can be unzipped on a Windows machine and it runs successfully (all the way down to Windows 7)
	-	i have added a cmd/windows folder which adds some GUI elements
		-	message box telling you whether wire-pod has started or needs setup, systray will tell you status and will let you quit or see logs
	-	all filepaths get set to os.UserConfigDir()/wire-pod/+file
	-	all features work on windows!
	-	installer ui is implemented, but it doesn't actually install yet

Need to:

	-	add easy way to get it to run at startup
	-	add way to delete apiConfig.json
	-	finish installer
	-	fancify GUI elements

COOL ICON FROM https://www.flaticon.com/authors/dooder
