package w

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

var (
	libkernel32 *windows.LazyDLL

	multiByteToWideChar *windows.LazyProc
)

func init() {
	libkernel32 = windows.NewLazySystemDLL("kernel32.dll")
	multiByteToWideChar = libkernel32.NewProc("MultiByteToWideChar")
}

const (
	MB_PRECOMPOSED       = 0x00000001
	MB_COMPOSITE         = 0x00000002
	MB_USEGLYPHCHARS     = 0x00000004
	MB_ERR_INVALID_CHARS = 0x00000008
)

func MultiByteToWideChar(CodePage uint32, dwFlags uint32, lpMultiByteStr *byte, cbMultiByte int32, lpWideCharStr LPWSTR, cchWideChar int32) int32 {
	ret, _, _ := syscall.Syscall6(multiByteToWideChar.Addr(), 6,
		uintptr(CodePage),
		uintptr(dwFlags),
		uintptr(unsafe.Pointer(lpMultiByteStr)),
		uintptr(cbMultiByte),
		uintptr(unsafe.Pointer(lpWideCharStr)),
		uintptr(cchWideChar),
	)
	return int32(ret)

}
