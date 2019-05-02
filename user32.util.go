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

func GetWindowText(hWnd HWND) (string, error) {
	var buf [512]WCHAR
	// not sure if GetWindowText terminates truncated
	// text so -1 to ensure there's always 0 at the end
	nMaxCount := int32(len(buf)) - 1
	res := GetWindowTextWSys(hWnd, &buf[0], nMaxCount)
	if res == 0 {
		return "", newGetLastError()
	}

	return FromUTF16(buf[:len(buf)]), nil
}
