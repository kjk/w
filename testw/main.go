package main

import (
	"fmt"
	"log"
	"os"
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
	var err error

	err = w.CoInitialize()
	if err != nil {
		log.Fatalf("w.CoInitialize() failed with %s\n", err)
	}
	defer w.CoUninitialize()

	dir, err := w.GetKnownFolderPath(w.CSIDL_DESKTOP)
	must(err)

	shortcutPath := filepath.Join(dir, "foo-sublime.lnk")

	exePath := `C:\Program Files\Sublime Text 3\sublime_text.exe`
	err = w.CreateShortcut(shortcutPath, exePath, "", "", 0)
	must(err)
	// TODO: figure out why shortcut is not created
	err = os.Remove(shortcutPath)
	//must(err)
}

func testGetLogicalDriveStrings() {
	s, err := w.GetLogicalDriveStrings()
	must(err)
	fmt.Printf("Logical drives: %#v\n", s)
}

func main() {
	testGetKnownFolderPath()
	testCreateShortcut()
	testGetLogicalDriveStrings()
	fmt.Printf("Hello from w tester\n")
}
