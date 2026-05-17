package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

func main() {
	exeDir, err := os.Executable()
	if err != nil {
		msgBox("Error", "Cannot determine executable path:\n"+err.Error())
		os.Exit(1)
	}
	exeDir = filepath.Dir(exeDir)

	electronExe := filepath.Join(exeDir, "node_modules", "electron", "dist", "electron.exe")

	if _, err := os.Stat(electronExe); os.IsNotExist(err) {
		msgBox("CCM Desktop v2.1",
			"Electron runtime not found.\n\n"+
				"Please run setup.bat first to install dependencies.\n\n"+
				"Path: "+electronExe)
		os.Exit(1)
	}

	// Try main.js next to the exe (portable release), then desktop/main.js (dev layout)
	mainJS := filepath.Join(exeDir, "main.js")
	if _, err := os.Stat(mainJS); os.IsNotExist(err) {
		mainJS = filepath.Join(exeDir, "desktop", "main.js")
	}
	if _, err := os.Stat(mainJS); os.IsNotExist(err) {
		msgBox("CCM Desktop v2.1",
			"Entry point not found.\n\n"+
				"Cannot find main.js in this directory.\n"+
				"Make sure all files are extracted correctly.")
		os.Exit(1)
	}

	cmd := exec.Command(electronExe, mainJS)
	cmd.Dir = exeDir
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if err := cmd.Start(); err != nil {
		msgBox("Error", "Failed to start CCM Desktop:\n"+err.Error())
		os.Exit(1)
	}

	// Release the process - we don't wait for it, the launcher exits immediately
	cmd.Process.Release()
}

func msgBox(title, text string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	msgBoxW := user32.NewProc("MessageBoxW")
	msgBoxW.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		uintptr(0x00000010|0x00000000), // MB_ICONERROR | MB_OK
	)

	// Also print to stderr for logging (won't be visible with windowsgui)
	fmt.Fprintln(os.Stderr, title+":", text)
}
