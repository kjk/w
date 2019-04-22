package w

import (
	"syscall"
	"unsafe"
)

const (
	CSIDL_DESKTOP                 = 0x0000
	CSIDL_INTERNET                = 0x0001
	CSIDL_PROGRAMS                = 0x0002
	CSIDL_CONTROLS                = 0x0003
	CSIDL_PRINTERS                = 0x0004
	CSIDL_MYDOCUMENTS             = 0x0005
	CSIDL_FAVORITES               = 0x0006
	CSIDL_STARTUP                 = 0x0007
	CSIDL_RECENT                  = 0x0008
	CSIDL_SENDTO                  = 0x0009
	CSIDL_BITBUCKET               = 0x000a
	CSIDL_STARTMENU               = 0x000b
	CSIDL_MYMUSIC                 = 0x000d
	CSIDL_MYVIDEO                 = 0x000e
	CSIDL_DESKTOPDIRECTORY        = 0x0010
	CSIDL_DRIVES                  = 0x0011
	CSIDL_NETWORK                 = 0x0012
	CSIDL_NETHOOD                 = 0x0013
	CSIDL_FONTS                   = 0x0014
	CSIDL_TEMPLATES               = 0x0015
	CSIDL_COMMON_STARTMENU        = 0x0016
	CSIDL_COMMON_PROGRAMS         = 0X0017
	CSIDL_COMMON_STARTUP          = 0x0018
	CSIDL_COMMON_DESKTOPDIRECTORY = 0x0019
	CSIDL_APPDATA                 = 0x001a
	CSIDL_PRINTHOOD               = 0x001b
	CSIDL_LOCAL_APPDATA           = 0x001c
	CSIDL_ALTSTARTUP              = 0x001d
	CSIDL_COMMON_ALTSTARTUP       = 0x001e
	CSIDL_COMMON_FAVORITES        = 0x001f
	CSIDL_INTERNET_CACHE          = 0x0020
	CSIDL_COOKIES                 = 0x0021
	CSIDL_HISTORY                 = 0x0022
	CSIDL_COMMON_APPDATA          = 0x0023
	CSIDL_WINDOWS                 = 0x0024
	CSIDL_SYSTEM                  = 0x0025
	CSIDL_PROGRAM_FILES           = 0x0026
	CSIDL_MYPICTURES              = 0x0027
	CSIDL_PROFILE                 = 0x0028
	CSIDL_SYSTEMX86               = 0x0029
	CSIDL_PROGRAM_FILESX86        = 0x002a
	CSIDL_PROGRAM_FILES_COMMON    = 0x002b
	CSIDL_PROGRAM_FILES_COMMONX86 = 0x002c
	CSIDL_COMMON_TEMPLATES        = 0x002d
	CSIDL_COMMON_DOCUMENTS        = 0x002e
	CSIDL_COMMON_ADMINTOOLS       = 0x002f
	CSIDL_ADMINTOOLS              = 0x0030
	CSIDL_CONNECTIONS             = 0x0031
	CSIDL_COMMON_MUSIC            = 0x0035
	CSIDL_COMMON_PICTURES         = 0x0036
	CSIDL_COMMON_VIDEO            = 0x0037
	CSIDL_RESOURCES               = 0x0038
	CSIDL_RESOURCES_LOCALIZED     = 0x0039
	CSIDL_COMMON_OEM_LINKS        = 0x003a
	CSIDL_CDBURN_AREA             = 0x003b
	CSIDL_COMPUTERSNEARME         = 0x003d
	CSIDL_FLAG_CREATE             = 0x8000
	CSIDL_FLAG_DONT_VERIFY        = 0x4000
	CSIDL_FLAG_DONT_UNEXPAND      = 0x2000
	CSIDL_FLAG_NO_ALIAS           = 0x1000
	CSIDL_FLAG_PER_USER_INIT      = 0x0800
)

type SHITEMID struct {
	Cb   uint16
	AbID [1]uint8
}

type ITEMIDLIST struct {
	Mkid SHITEMID
}

const (
	SLGP_SHORTPATH        = 0x1
	SLGP_UNCPRIORITY      = 0x2
	SLGP_RAWPATH          = 0x4
	SLGP_RELATIVEPRIORITY = 0x8
)

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