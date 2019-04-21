package w

import (
	"errors"
	"fmt"
)

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

// TODO: finish me
/*
func CreateShortcut(shortcutPath string, exePath string, args string, description string, iconIndex int) error {
	var lnkPtr unsafe.Pointer
	hr := CoGetClassObject(&IID_IShellLinkW, CLSCTX_ALL, nil, &IID_IClassFactory, &lnkPtr)
	if FAILED(hr) {
		return errorFromHRESULT("CoGetClassObject", hr)
	}
	lnk := (*IClassFactory)(lnkPtr)
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

