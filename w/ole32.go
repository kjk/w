package w

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

var (
	libole32 *windows.LazyDLL

	coGetClassObject *windows.LazyProc
	coCreateInstance *windows.LazyProc
	coInitialize     *windows.LazyProc
	coUninitialize   *windows.LazyProc
)

func init() {
	libole32 = windows.NewLazySystemDLL("ole32.dll")
	coGetClassObject = libole32.NewProc("CoGetClassObject")
	coCreateInstance = libole32.NewProc("CoCreateInstance")
	coInitialize = libole32.NewProc("CoInitialize")
	coUninitialize = libole32.NewProc("CoUninitialize")
}

type COAUTHIDENTITY struct {
	User           *uint16
	UserLength     uint32
	Domain         *uint16
	DomainLength   uint32
	Password       *uint16
	PasswordLength uint32
	Flags          uint32
}

type COAUTHINFO struct {
	DwAuthnSvc           uint32
	DwAuthzSvc           uint32
	PwszServerPrincName  LPWSTR
	DwAuthnLevel         uint32
	DwImpersonationLevel uint32
	PAuthIdentityData    *COAUTHIDENTITY
	DwCapabilities       uint32
}

type COSERVERINFO struct {
	DwReserved1 uint32
	PwszName    LPWSTR
	PAuthInfo   *COAUTHINFO
	DwReserved2 uint32
}

func CoGetClassObject(rclsid *GUID, dwClsContext uint32, pServerInfo *COSERVERINFO, riid *GUID, ppv *unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.Syscall6(coGetClassObject.Addr(), 5,
		uintptr(unsafe.Pointer(rclsid)),
		uintptr(dwClsContext),
		uintptr(unsafe.Pointer(pServerInfo)),
		uintptr(unsafe.Pointer(riid)),
		uintptr(unsafe.Pointer(ppv)),
		0,
	)
	return HRESULT(ret)

}

func CoCreateInstance(rclsid *GUID, pUnkOuter *IUnknown, dwClsContext uint32, riid *GUID, ppv *unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.Syscall6(coCreateInstance.Addr(), 5,
		uintptr(unsafe.Pointer(rclsid)),
		uintptr(unsafe.Pointer(pUnkOuter)),
		uintptr(dwClsContext),
		uintptr(unsafe.Pointer(riid)),
		uintptr(unsafe.Pointer(ppv)),
		0,
	)
	return HRESULT(ret)

}

func CoInitialize(pvReserved unsafe.Pointer) HRESULT {
	ret, _, _ := syscall.Syscall(coInitialize.Addr(), 1,
		uintptr(pvReserved),
		0,
		0,
	)
	return HRESULT(ret)

}

func CoUninitialize() {
	_, _, _ = syscall.Syscall(coUninitialize.Addr(), 0,
		0,
		0,
		0,
	)

}
