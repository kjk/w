package w

import (
	"sync/atomic"
	"unicode/utf8"

	"golang.org/x/sys/windows"
)

// fast conversion of utf8 => utf16 conversion for short-lived strings

const (
	bufSize = 16 * 1024
)

// ToUTF16 converts an UTF-8 string to Window's UTF-16 unicode
func ToUTF16(s string) []uint16 {
	// TODO:
	// - a way for the app to control what happens on error
	//  (ignore, crash or maybe call a registered callback)
	// - optimize for temporary allocations
	a, _ := windows.UTF16FromString(s)
	return a
}

// FromUTF16 converts UTF-16 Windows string to UTF-8 Go string
func FromUTF16(s []uint16) string {
	// allows to shorten the callers
	if len(s) == 0 || s[0] == 0 {
		return ""
	}
	return windows.UTF16ToString(s)
}

var buf []uint16 = make([]uint16, bufSize)
var bufTaken atomic.Bool

// ToUTF16ShortLived converts s to a zero-terminated UTF-16
// Windows string. Returns a slice that doesn't include
// terminating zero (but it's there)
// TODO: explicilty return if buffer was taken
func ToUTF16ShortLived(s string) []uint16 {
	n := len(s)
	if n+1 >= bufSize {
		// if the string is too long, use the regular allocator
		return ToUTF16(s)
	}
	taken := bufTaken.CompareAndSwap(false, true)
	if taken {
		return ToUTF16(s)
	}

	// now it's taken by us
	// optimistically try a fast path for just ascii
	isASCII := true
	res := buf
	for i := 0; i < n; i++ {
		c := s[i]
		if c > utf8.RuneSelf {
			isASCII = false
			break
		}
		res[i] = uint16(c)
	}
	if isASCII {
		// add terminating 0
		res[n] = 0
		return res[:n]
	}

	// buffer is not taken
	bufTaken.Store(false)
	return ToUTF16(s)
}

func FreeShortLivedUTF16(s []uint16) {
	taken := bufTaken.Load()
	if !taken {
		return
	}
	// is taken. could be by us or by someone else
	if cap(s) != bufSize {
		// not our size so not take by us
		return
	}
	// check if s is the same as buf by comparing the address of the first element
	if &s[0] != &buf[0] {
		// not our buffer, so not taken by us
		return
	}
	bufTaken.Store(false)
}
