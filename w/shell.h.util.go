package w

import (
	"errors"
	"fmt"
	"unsafe"
)

// TODO: I want to auto-generate CLSID_* but they don't seem
// to be present in the api info

// CLSID_ShellLink represents id of IShellLink
var CLSID_ShellLink = IID{0x00021401, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

// HrOk returns true if hr represents a "success" HRESULT
func HrOk(hr HRESULT) bool {
	return hr >= 0
}

// HrFailed returns true if hr represents a "failed" HRESULT
func HrFailed(hr HRESULT) bool {
	return hr < 0
}

// SUCCEEDED returns true if hr represents a "success" HRESULT
func SUCCEEDED(hr HRESULT) bool {
	return hr >= 0
}

// FAILED returns true if hr represents a "failed" HRESULT
func FAILED(hr HRESULT) bool {
	return hr < 0
}

func newError(message string) error {
	return errors.New(message)
}

func errorFromHRESULT(funcName string, hr HRESULT) error {
	return newError(fmt.Sprintf("%s: Error %d", funcName, hr))
}

// Based on https://docs.microsoft.com/en-us/windows/desktop/shell/links
// https://stackoverflow.com/questions/3906974/how-to-programmatically-create-a-shortcut-using-win32
func CreateShortcut(shortcutPath string, exePath string, args string, description string, iconIndex int) error {
	//IShellLink * psl

	var pslPtr unsafe.Pointer
	hr := CoCreateInstance(&CLSID_ShellLink, nil, CLSCTX_INPROC_SERVER, &IID_IShellLinkW, &pslPtr)
	if FAILED(hr) {
		return errorFromHRESULT("CoGetClassObject", hr)
	}
	psl := (*IShellLinkW)(pslPtr)
	defer psl.Release()

	{
		s := ToUnicodeShortLived(exePath)
		hr = psl.SetPath(s)
		hr2 := psl.SetIconLocation(s, int32(iconIndex))
		FreeShortLivedUnicode(s)
		if FAILED(hr) {
			return errorFromHRESULT("psl.SetPath", hr)
		}
		if FAILED(hr2) {
			return errorFromHRESULT("psl.SetIconLocation", hr2)
		}
	}

	if len(description) > 0 {
		s := ToUnicodeShortLived(description)
		hr = psl.SetDescription(s)
		FreeShortLivedUnicode(s)
		if FAILED(hr) {
			return errorFromHRESULT("psl.SetPath", hr)
		}
	}

	if len(args) > 0 {
		s := ToUnicodeShortLived(args)
		hr = psl.SetArguments(s)
		FreeShortLivedUnicode(s)
		if FAILED(hr) {
			return errorFromHRESULT("psl.SetArguments", hr)
		}
	}

	var ppfPtr unsafe.Pointer
	hr = psl.QueryInterface(&IID_IPersistFile, &ppfPtr)
	if FAILED(hr) {
		return errorFromHRESULT("psl.QueryInterface", hr)
	}

	ppf := (*IPersistFile)(ppfPtr)
	defer ppf.Release()

	{
		s := ToUnicodeShortLived(shortcutPath)
		hr = ppf.Save(s, TRUE)
		FreeShortLivedUnicode(s)
		if FAILED(hr) {
			return errorFromHRESULT("ppf.Save", hr)
		}
	}

	return nil
}
