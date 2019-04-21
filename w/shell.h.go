package w

import (
	"syscall"
	"unsafe"
)

const (
	SLGP_SHORTPATH        = 0x1
	SLGP_UNCPRIORITY      = 0x2
	SLGP_RAWPATH          = 0x4
	SLGP_RELATIVEPRIORITY = 0x8
)

type SHITEMID struct {
	Cb   uint16
	AbID [1]uint8
}

type ITEMIDLIST struct {
	Mkid SHITEMID
}

const (
	SLR_NO_UI                     = 0x1
	SLR_ANY_MATCH                 = 0x2
	SLR_UPDATE                    = 0x4
	SLR_NOUPDATE                  = 0x8
	SLR_NOSEARCH                  = 0x10
	SLR_NOTRACK                   = 0x20
	SLR_NOLINKINFO                = 0x40
	SLR_INVOKE_MSI                = 0x80
	SLR_NO_UI_WITH_MSG_PUMP       = 0x101
	SLR_OFFER_DELETE_WITHOUT_FILE = 0x200
	SLR_KNOWNFOLDER               = 0x400
	SLR_MACHINE_IN_LOCAL_TARGET   = 0x800
	SLR_UPDATE_MACHINE_AND_SID    = 0x1000
)

var IID_IShellLinkW = IID{0x000214F9, 0x0000, 0x0000, [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IShellLinkWVtbl struct {
	IUnknownVtbl
	GetPath             uintptr
	GetIDList           uintptr
	SetIDList           uintptr
	GetDescription      uintptr
	SetDescription      uintptr
	GetWorkingDirectory uintptr
	SetWorkingDirectory uintptr
	GetArguments        uintptr
	SetArguments        uintptr
	GetHotkey           uintptr
	SetHotkey           uintptr
	GetShowCmd          uintptr
	SetShowCmd          uintptr
	GetIconLocation     uintptr
	SetIconLocation     uintptr
	SetRelativePath     uintptr
	Resolve             uintptr
	SetPath             uintptr
}

type IShellLinkW struct {
	IUnknown
	Vtbl *IShellLinkWVtbl
}

func (i *IShellLinkW) GetPath(pszFile LPWSTR, cch int32, pfd *WIN32_FIND_DATAW, fFlags uint32) HRESULT {
	ret, _, _ := syscall.Syscall6(i.Vtbl.GetPath, 5,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(pszFile)),
		uintptr(cch),
		uintptr(unsafe.Pointer(pfd)),
		uintptr(fFlags),
		0,
	)
	return HRESULT(ret)

}

func (i *IShellLinkW) GetIDList(ppidl **ITEMIDLIST) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.GetIDList, 2,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(ppidl)),
		0,
	)
	return HRESULT(ret)

}

func (i *IShellLinkW) SetIDList(pidl *ITEMIDLIST) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.SetIDList, 2,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(pidl)),
		0,
	)
	return HRESULT(ret)

}

func (i *IShellLinkW) GetDescription(pszName LPWSTR, cch int32) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.GetDescription, 3,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(pszName)),
		uintptr(cch),
	)
	return HRESULT(ret)

}

func (i *IShellLinkW) SetDescription(pszName LPCWSTR) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.SetDescription, 2,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(pszName)),
		0,
	)
	return HRESULT(ret)

}

func (i *IShellLinkW) GetWorkingDirectory(pszDir LPWSTR, cch int32) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.GetWorkingDirectory, 3,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(pszDir)),
		uintptr(cch),
	)
	return HRESULT(ret)

}

func (i *IShellLinkW) SetWorkingDirectory(pszDir LPCWSTR) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.SetWorkingDirectory, 2,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(pszDir)),
		0,
	)
	return HRESULT(ret)

}

func (i *IShellLinkW) GetArguments(pszArgs LPWSTR, cch int32) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.GetArguments, 3,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(pszArgs)),
		uintptr(cch),
	)
	return HRESULT(ret)

}

func (i *IShellLinkW) SetArguments(pszArgs LPCWSTR) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.SetArguments, 2,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(pszArgs)),
		0,
	)
	return HRESULT(ret)

}

func (i *IShellLinkW) GetHotkey(pwHotkey *uint16) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.GetHotkey, 2,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(pwHotkey)),
		0,
	)
	return HRESULT(ret)

}

func (i *IShellLinkW) SetHotkey(wHotkey uint16) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.SetHotkey, 2,
		uintptr(unsafe.Pointer(i)),
		uintptr(wHotkey),
		0,
	)
	return HRESULT(ret)

}

func (i *IShellLinkW) GetShowCmd(piShowCmd *int32) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.GetShowCmd, 2,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(piShowCmd)),
		0,
	)
	return HRESULT(ret)

}

func (i *IShellLinkW) SetShowCmd(iShowCmd int32) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.SetShowCmd, 2,
		uintptr(unsafe.Pointer(i)),
		uintptr(iShowCmd),
		0,
	)
	return HRESULT(ret)

}

func (i *IShellLinkW) GetIconLocation(pszIconPath LPWSTR, cch int32, piIcon *int32) HRESULT {
	ret, _, _ := syscall.Syscall6(i.Vtbl.GetIconLocation, 4,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(pszIconPath)),
		uintptr(cch),
		uintptr(unsafe.Pointer(piIcon)),
		0,
		0,
	)
	return HRESULT(ret)

}

func (i *IShellLinkW) SetIconLocation(pszIconPath LPCWSTR, iIcon int32) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.SetIconLocation, 3,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(pszIconPath)),
		uintptr(iIcon),
	)
	return HRESULT(ret)

}

func (i *IShellLinkW) SetRelativePath(pszPathRel LPCWSTR, dwReserved uint32) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.SetRelativePath, 3,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(pszPathRel)),
		uintptr(dwReserved),
	)
	return HRESULT(ret)

}

func (i *IShellLinkW) Resolve(hwnd HWND, fFlags uint32) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.Resolve, 3,
		uintptr(unsafe.Pointer(i)),
		uintptr(hwnd),
		uintptr(fFlags),
	)
	return HRESULT(ret)

}

func (i *IShellLinkW) SetPath(pszFile LPCWSTR) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.SetPath, 2,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(pszFile)),
		0,
	)
	return HRESULT(ret)

}
