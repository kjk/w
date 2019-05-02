package w

import (
	"syscall"
	"unsafe"
)

func EnumWindows(cb func(HWND) bool) {
	innerCb := func(hwnd HWND, lParam uintptr) uintptr {
		cont := cb(hwnd)
		return boolToUintptr(cont)
	}
	innerCbPtr := unsafe.Pointer(syscall.NewCallback(innerCb))
	EnumWindowsSys(innerCbPtr, 0)
}
