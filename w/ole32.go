package w

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

var (
	libole32 *windows.LazyDLL

	coGetClassObject *windows.LazyProc
)

func init() {
	libole32 = windows.NewLazySystemDLL("ole32.dll")
	coGetClassObject = libole32.NewProc("CoGetClassObject")
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
