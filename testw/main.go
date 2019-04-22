package main

import (
	"fmt"
	"log"

	"github.com/kjk/winapigen/w"
)

var (
	// import a type from w to make sure it compiles
	f w.HWND
)

func testCreateShortcut() {
	hr := w.CoInitialize(nil)
	if w.HrFailed(hr) {
		log.Fatalf("w.CoInitialize() failed with %d\n", hr)
	}
	defer w.CoUninitialize()

	shortcutPath := ``
	exePath := `C:\Program Files\Sublime Text 3\sublime_text.exe`
	err := w.CreateShortcut(shortcutPath, exePath, "", "", 0)
	if err != nil {
		log.Fatalf("w.CreateShortcut() failed with %s\n", err)
	}
}

func main() {
	testCreateShortcut()
	fmt.Printf("Hello from w tester\n")
}
