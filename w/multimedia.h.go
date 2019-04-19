package w

import (
	"syscall"
	"unsafe"
)

var IID_IUnknown = IID{0x00000000, 0x0000, 0x0000, [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

type IUnknownVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
}

type IUnknown struct {
	Vtbl *IUnknownVtbl
}

func (i *IUnknown) QueryInterface(riid *GUID, ppvObject *uintptr) HRESULT {
	ret, _, _ := syscall.Syscall(i.Vtbl.QueryInterface, 3,
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(riid)),
		uintptr(unsafe.Pointer(ppvObject)),
	)
	return HRESULT(ret)

}

func (i *IUnknown) AddRef() uint32 {
	ret, _, _ := syscall.Syscall(i.Vtbl.AddRef, 1,
		uintptr(unsafe.Pointer(i)),
		0,
		0,
	)
	return uint32(ret)

}

func (i *IUnknown) Release() uint32 {
	ret, _, _ := syscall.Syscall(i.Vtbl.Release, 1,
		uintptr(unsafe.Pointer(i)),
		0,
		0,
	)
	return uint32(ret)

}
