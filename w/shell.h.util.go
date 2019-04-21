package w

import (
	"errors"
	"fmt"
)

// TODO: how can I auto-generate those
var CLSID_ShellLink = IID{0x00021401, 0x0000, 0x0000,
	[8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

func SUCCEEDED(hr HRESULT) bool {
	return hr >= 0
}

func FAILED(hr HRESULT) bool {
	return hr < 0
}

func newError(message string) error {
	return errors.New(message)
}

func errorFromHRESULT(funcName string, hr HRESULT) error {
	return newError(fmt.Sprintf("%s: Error %d", funcName, hr))
}

/*
bool CreateShortcut(const WCHAR* shortcutPath, const WCHAR* exePath, const WCHAR* args, const WCHAR* description,
                    int iconIndex) {
    ScopedCom com;

    ScopedComPtr<IShellLink> lnk;
    if (!lnk.Create(CLSID_ShellLink))
        return false;

    ScopedComQIPtr<IPersistFile> file(lnk);
    if (!file)
        return false;

    HRESULT hr = lnk->SetPath(exePath);
    if (FAILED(hr))
        return false;

    lnk->SetWorkingDirectory(AutoFreeW(path::GetDir(exePath)));
    // lnk->SetShowCmd(SW_SHOWNORMAL);
    // lnk->SetHotkey(0);
    lnk->SetIconLocation(exePath, iconIndex);
    if (args)
        lnk->SetArguments(args);
    if (description)
        lnk->SetDescription(description);

    hr = file->Save(shortcutPath, TRUE);
    return SUCCEEDED(hr);
}
*/

/*
HRESULT CreateLink(LPCWSTR lpszPathObj, LPCSTR lpszPathLink, LPCWSTR lpszDesc)
{
    HRESULT hres;
    IShellLink* psl;

    // Get a pointer to the IShellLink interface. It is assumed that CoInitialize
    // has already been called.
    hres = CoCreateInstance(CLSID_ShellLink, NULL, CLSCTX_INPROC_SERVER, IID_IShellLink, (LPVOID*)&psl);
    if (SUCCEEDED(hres))
    {
        IPersistFile* ppf;

        // Set the path to the shortcut target and add the description.
        psl->SetPath(lpszPathObj);
        psl->SetDescription(lpszDesc);

        // Query IShellLink for the IPersistFile interface, used for saving the
        // shortcut in persistent storage.
        hres = psl->QueryInterface(IID_IPersistFile, (LPVOID*)&ppf);

        if (SUCCEEDED(hres))
        {
            WCHAR wsz[MAX_PATH];

            // Ensure that the string is Unicode.
            MultiByteToWideChar(CP_ACP, 0, lpszPathLink, -1, wsz, MAX_PATH);

            // Add code here to check return value from MultiByteWideChar
            // for success.

            // Save the link by calling IPersistFile::Save.
            hres = ppf->Save(wsz, TRUE);
            ppf->Release();
        }
        psl->Release();
    }
	return hres;
*/

// TODO: finish me
// Based on https://docs.microsoft.com/en-us/windows/desktop/shell/links
// https://stackoverflow.com/questions/3906974/how-to-programmatically-create-a-shortcut-using-win32

/*
func CreateShortcut(shortcutPath string, exePath string, args string, description string, iconIndex int) error {
	//IShellLink * psl

	var pslPtr unsafe.Pointer
	hr := CoCreateInstance(&CLSID_ShellLink, nil, CLSCTX_INPROC_SERVER, &IID_IShellLink, &pslPtr)
	if FAILED(hr) {
		return errorFromHRESULT("CoGetClassObject", hr)
	}
	psl := (*IClassFactory)(lnkPtr)
	defer lnk.Release()

	var filePtr unsafe.Pointer

	hr = lnk.CreateInstance(nil, &IID_IPersistFile, &filePtr)
	if FAILED(hr) {
		return errorFromHRESULT("IClassFactory.CreateInstance", hr)
	}

	file := (*IPersistFile)(filePtr)
	defer file.Release()

	exePathW := ToUnicode(exePath)
	hr = lnk.SetPath(exePathW)
	FreeUnicode(exePathW)
	if FAILED(hr) {
		return errorFromHRESULT("lnk.SetPath", hr)
	}

	shortcutPathW := ToUnicode(shortcutPath)
	hr = file.Save(shortcutPathW, TRUE)
	if FAILED(hr) {
		return errorFromHRESULT("file.Save", hr)
	}
	FreeUnicode(shortcutPathW)

	return nil
}
*/
