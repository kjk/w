package w

import (
	"errors"
	"fmt"
	"syscall"

	"golang.org/x/sys/windows"
)

const (
	TRUE  = 1
	FALSE = 0
)

type BOOL int32
type HRESULT int32
type HANDLE uintptr
type HMODULE uintptr
type WCHAR = uint16
type LARGE_INTEGER int64
type ULARGE_INTEGER uint64

// we don't desugar those types because they signal const vs. non-const
type LPCWSTR = *uint16
type LPWSTR = *uint16

type IID = windows.GUID
type GUID = windows.GUID

func panicIf(cond bool, args ...interface{}) {
	if !cond {
		return
	}
	if len(args) == 0 {
		panic("condition failed")
	}
	format := args[0].(string)
	args = args[1:]
	if len(args) == 0 {
		panic(format)
	}
	panic(fmt.Sprintf(format, args...))
}

func winError(s string) error {
	// TODO: use getlasterror, add a callstack
	if s == "" {
		s = "generic windows error"
	}
	return errors.New(s)
}

func sys(trap, nargs, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	return syscall.Syscall(trap, nargs, a1, a2, a3)
}

func sys6(trap, nargs, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	return syscall.Syscall6(trap, nargs, a1, a2, a3, a4, a5, a6)
}

func sys9(trap, nargs, a1, a2, a3, a4, a5, a6, a7, a8, a9 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	return syscall.Syscall9(trap, nargs, a1, a2, a3, a4, a5, a6, a7, a8, a9)
}

func sys12(trap, nargs, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	return syscall.Syscall12(trap, nargs, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12)
}

func sys15(trap, nargs, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	return syscall.Syscall15(trap, nargs, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15)
}
