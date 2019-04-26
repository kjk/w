package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/kjk/winapigen/w"
)

func panicIf(cond bool, args ...interface{}) {
	if !cond {
		return
	}
	msg := "condition failed"
	if len(args) > 0 {
		msg = args[0].(string)
		if len(args) > 1 {
			msg = fmt.Sprintf(msg, args[1:]...)
		}
	}
	panic(msg)
}

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

var (
	alphabetForRand = ""
)

func genCharsInRange(start, end byte) string {
	n := int(end) - int(start)
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = start + byte(i)
	}
	return string(b)
}

func genRandChar() byte {
	if alphabetForRand == "" {
		s := genCharsInRange('a', 'z')
		s += genCharsInRange('A', 'Z')
		s += genCharsInRange('0', '9')
		alphabetForRand = s
	}
	pos := rand.Intn(len(alphabetForRand))
	return alphabetForRand[pos]
}

func genRandString(n int) string {
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = genRandChar()
	}
	return string(b)
}

func testReg() {
	fmt.Printf("testReg()\n")
	{
		// test we can open a key (this key should always exist)
		key := `Software\Microsoft\Windows\CurrentVersion\Uninstall`
		hkey, didCreate, err := w.RegCreateKeyEx(w.HKEY_CURRENT_USER, key)
		must(err)
		panicIf(didCreate)
		err = w.RegCloseKey(hkey)
		must(err)
		fmt.Printf("  opened key '%s'\n", key)
	}

	{
		// test we can create a key (with randomly generated name)
		key := `Software\Microsoft\Windows\CurrentVersion\Uninstall\` + genRandString(16)
		hkey, didCreate, err := w.RegCreateKeyEx(w.HKEY_CURRENT_USER, key)
		must(err)
		panicIf(!didCreate)
		err = w.RegCloseKey(hkey)
		must(err)
		fmt.Printf("  created key '%s'\n", key)

		err = w.RegDeleteKeyEx(w.HKEY_CURRENT_USER, key)
		must(err)
		fmt.Printf("  deleted key '%s'\n", key)
	}

	{
		// test opening non-existent key fails with ERROR_FILE_NOT_FOUND (2)
		key := `Software\Microsoft\Windows\CurrentVersion\Uninstall\` + genRandString(16)
		_, err := w.RegOpenKeyForRead(w.HKEY_CURRENT_USER, key)
		panicIf(!strings.Contains(err.Error(), " error code 2"))
	}
}

func testCreateShortcut() {
	fmt.Printf("testCreateShortcut()\n")
	var err error

	err = w.CoInitialize()
	must(err)
	defer w.CoUninitialize()

	dir, err := w.GetKnownFolderPath(w.CSIDL_DESKTOP)
	must(err)

	shortcutPath := filepath.Join(dir, "TestSublime.lnk")

	exePath := `C:\Program Files\Sublime Text 3\sublime_text.exe`
	err = w.CreateShortcut(shortcutPath, exePath, "", "", 0)
	must(err)

	err = os.Remove(shortcutPath)
	must(err)
}

func testGetLogicalDriveStrings() {
	s, err := w.GetLogicalDriveStrings()
	must(err)
	fmt.Printf("Logical drives: %#v\n", s)
}

func main() {
	fmt.Printf("Running windows API tests\n")
	testReg()
	//testGetKnownFolderPath()
	//testCreateShortcut()
	//testGetLogicalDriveStrings()
	fmt.Printf("Finished windows API tests\n")
}
