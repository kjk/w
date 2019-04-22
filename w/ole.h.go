package w

import (
	"syscall"
	"unsafe"
)

const (
	CLSCTX_INPROC_SERVER          = 0x1
	CLSCTX_INPROC_HANDLER         = 0x2
	CLSCTX_LOCAL_SERVER           = 0x4
	CLSCTX_INPROC_SERVER16        = 0x8
	CLSCTX_REMOTE_SERVER          = 0x10
	CLSCTX_INPROC_HANDLER16       = 0x20
	CLSCTX_RESERVED1              = 0x40
	CLSCTX_RESERVED2              = 0x80
	CLSCTX_RESERVED3              = 0x100
	CLSCTX_RESERVED4              = 0x200
	CLSCTX_NO_CODE_DOWNLOAD       = 0x400
	CLSCTX_RESERVED5              = 0x800
	CLSCTX_NO_CUSTOM_MARSHAL      = 0x1000
	CLSCTX_ENABLE_CODE_DOWNLOAD   = 0x2000
	CLSCTX_NO_FAILURE_LOG         = 0x4000
	CLSCTX_DISABLE_AAA            = 0x8000
	CLSCTX_ENABLE_AAA             = 0x10000
	CLSCTX_FROM_DEFAULT_CONTEXT   = 0x20000
	CLSCTX_ACTIVATE_32_BIT_SERVER = 0x40000
	CLSCTX_ACTIVATE_64_BIT_SERVER = 0x80000
	CLSCTX_ENABLE_CLOAKING        = 0x100000
	CLSCTX_PS_DLL                 = 0x80000000
)

const (
	STGM_DIRECT           = 0x00000000
	STGM_TRANSACTED       = 0x00010000
	STGM_SIMPLE           = 0x08000000
	STGM_READ             = 0x00000000
	STGM_WRITE            = 0x00000001
	STGM_READWRITE        = 0x00000002
	STGM_SHARE_DENY_NONE  = 0x00000040
	STGM_SHARE_DENY_READ  = 0x00000030
	STGM_SHARE_DENY_WRITE = 0x00000020
	STGM_SHARE_EXCLUSIVE  = 0x00000010
	STGM_PRIORITY         = 0x00040000
	STGM_DELETEONRELEASE  = 0x04000000
	STGM_NOSCRATCH        = 0x00100000
	STGM_CREATE           = 0x00001000
	STGM_CONVERT          = 0x00020000
	STGM_FAILIFTHERE      = 0x00000000
	STGM_NOSNAPSHOT       = 0x00200000
	STGM_DIRECT_SWMR      = 0x00400000
)

var IID_IClassFactory = IID{0x00000001, 0x0000, 0x0000, [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IClassFactoryVtbl struct {
	// IUnknown
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	// IClassFactory
	CreateInstance uintptr
	LockServer     uintptr
}

type IClassFactory struct {
	Vtbl *IClassFactoryVtbl
}

// methods for IUnknown

func (i *IClassFactory) QueryInterface(riid *GUID, ppvObject *unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.QueryInterface, 3,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(riid)),
		uintptr(unsafe.Pointer(ppvObject)),
	)
	return HRESULT(ret)
}

func (i *IClassFactory) AddRef() uint32 {
	ret, _, _ := syscall.Syscall(i.Vtbl.AddRef, 1,
		uintptr(unsafe.Pointer(i)),
		0,
		0,
	)
	return uint32(ret)
}

func (i *IClassFactory) Release() uint32 {
	ret, _, _ := syscall.Syscall(i.Vtbl.Release, 1,
		uintptr(unsafe.Pointer(i)),
		0,
		0,
	)
	return uint32(ret)
}

// methods for IClassFactory

func (i *IClassFactory) CreateInstance(pUnkOuter *IUnknown, riid *GUID, ppvObject *unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.Syscall6(i.Vtbl.CreateInstance, 4,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(pUnkOuter)),
		uintptr(unsafe.Pointer(riid)),
		uintptr(unsafe.Pointer(ppvObject)),
		0,
		0,
	)
	return HRESULT(ret)
}

func (i *IClassFactory) LockServer(fLock int32) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.LockServer, 2,
		uintptr(unsafe.Pointer(i)),
		uintptr(fLock),
		0,
	)
	return HRESULT(ret)
}

var IID_IPersistFile = IID{0x0000010b, 0x0000, 0x0000, [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IPersistFileVtbl struct {
	// IUnknown
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	// IPersist
	GetClassID uintptr

	// IPersistFile
	IsDirty       uintptr
	Load          uintptr
	Save          uintptr
	SaveCompleted uintptr
	GetCurFile    uintptr
}

type IPersistFile struct {
	Vtbl *IPersistFileVtbl
}

// methods for IUnknown

func (i *IPersistFile) QueryInterface(riid *GUID, ppvObject *unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.QueryInterface, 3,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(riid)),
		uintptr(unsafe.Pointer(ppvObject)),
	)
	return HRESULT(ret)
}

func (i *IPersistFile) AddRef() uint32 {
	ret, _, _ := syscall.Syscall(i.Vtbl.AddRef, 1,
		uintptr(unsafe.Pointer(i)),
		0,
		0,
	)
	return uint32(ret)
}

func (i *IPersistFile) Release() uint32 {
	ret, _, _ := syscall.Syscall(i.Vtbl.Release, 1,
		uintptr(unsafe.Pointer(i)),
		0,
		0,
	)
	return uint32(ret)
}

// methods for IPersist

func (i *IPersistFile) GetClassID(pClassID *GUID) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.GetClassID, 2,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(pClassID)),
		0,
	)
	return HRESULT(ret)
}

// methods for IPersistFile

func (i *IPersistFile) IsDirty() HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.IsDirty, 1,
		uintptr(unsafe.Pointer(i)),
		0,
		0,
	)
	return HRESULT(ret)
}

func (i *IPersistFile) Load(pszFileName LPCWSTR, dwMode uint32) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.Load, 3,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(pszFileName)),
		uintptr(dwMode),
	)
	return HRESULT(ret)
}

func (i *IPersistFile) Save(pszFileName LPCWSTR, fRemember int32) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.Save, 3,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(pszFileName)),
		uintptr(fRemember),
	)
	return HRESULT(ret)
}

func (i *IPersistFile) SaveCompleted(pszFileName LPCWSTR) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.SaveCompleted, 2,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(pszFileName)),
		0,
	)
	return HRESULT(ret)
}

func (i *IPersistFile) GetCurFile(ppszFileName *LPCWSTR) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.GetCurFile, 2,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(ppszFileName)),
		0,
	)
	return HRESULT(ret)
}

var IID_IPersist = IID{0x0000010c, 0x0000, 0x0000, [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IPersistVtbl struct {
	// IUnknown
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	// IPersist
	GetClassID uintptr
}

type IPersist struct {
	Vtbl *IPersistVtbl
}

// methods for IUnknown

func (i *IPersist) QueryInterface(riid *GUID, ppvObject *unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.QueryInterface, 3,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(riid)),
		uintptr(unsafe.Pointer(ppvObject)),
	)
	return HRESULT(ret)
}

func (i *IPersist) AddRef() uint32 {
	ret, _, _ := syscall.Syscall(i.Vtbl.AddRef, 1,
		uintptr(unsafe.Pointer(i)),
		0,
		0,
	)
	return uint32(ret)
}

func (i *IPersist) Release() uint32 {
	ret, _, _ := syscall.Syscall(i.Vtbl.Release, 1,
		uintptr(unsafe.Pointer(i)),
		0,
		0,
	)
	return uint32(ret)
}

// methods for IPersist

func (i *IPersist) GetClassID(pClassID *GUID) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.GetClassID, 2,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(pClassID)),
		0,
	)
	return HRESULT(ret)
}
