package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/kjk/winapigen/w"
)

func testGetKnownFolderPath() {
	ids := []int{
		w.CSIDL_APPDATA,
		w.CSIDL_LOCAL_APPDATA,
		w.CSIDL_COMMON_APPDATA,
		w.CSIDL_DESKTOP,
		w.CSIDL_PROGRAMS,
		w.CSIDL_COMMON_PROGRAMS,
		w.CSIDL_STARTUP,
		w.CSIDL_ALTSTARTUP,
		w.CSIDL_COMMON_ALTSTARTUP,
		w.CSIDL_STARTMENU,
		w.CSIDL_COMMON_STARTMENU,
	}
	idNames := []string{
		"CSIDL_APPDATA",
		"CSIDL_LOCAL_APPDATA",
		"CSIDL_COMMON_APPDATA",
		"CSIDL_DESKTOP",
		"CSIDL_PROGRAMS",
		"CSIDL_COMMON_PROGRAMS",
		"CSIDL_STARTUP",
		"CSIDL_ALTSTARTUP",
		"CSIDL_COMMON_ALTSTARTUP",
		"CSIDL_STARTMENU",
		"CSIDL_COMMON_STARTMENU",
	}
	for i := 0; i < len(ids); i++ {
		s, err := w.GetKnownFolderPath(ids[i])
		if err != nil {
			log.Fatalf("w.GetKnownFolderPath() failed with %s\n", err)
		}
		fmt.Printf("Known folder for %s: %s\n", idNames[i], s)
	}
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func testCreateShortcut() {
	dir, err := w.GetKnownFolderPath(w.CSIDL_DESKTOP)
	must(err)

	shortcutPath := filepath.Join(dir, "foo-sublime.lnk")

	hr := w.CoInitialize(nil)
	if w.HrFailed(hr) {
		log.Fatalf("w.CoInitialize() failed with %d\n", hr)
	}
	defer w.CoUninitialize()

	exePath := `C:\Program Files\Sublime Text 3\sublime_text.exe`
	err = w.CreateShortcut(shortcutPath, exePath, "", "", 0)
	must(err)
}

func main() {
	testGetKnownFolderPath()
	testCreateShortcut()
	fmt.Printf("Hello from w tester\n")
}
