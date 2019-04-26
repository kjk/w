package w

import (
	"sync"
	"unicode/utf8"
	"unsafe"

	"golang.org/x/sys/windows"
)

// fast conversion of utf8 => utf16 conversion for short-lived strings

const (
	stringAllocatorDefaultBufSize = 16 * 1024
)

var (
	stringAllocator   []uint16
	muStringAllocator sync.Mutex

	stringAllocatorCurrPos int
	lastAllocationPos      int
	lastAllocationSize     int

	// counters so that we can see how often we are able to
	// free string in a way that re-uses stringAllocator
	nFrees     int
	nFastFrees int
)

// for testing
func resetAllocator() {
	stringAllocator = nil
	stringAllocatorCurrPos = 0
	lastAllocationPos = 0
	lastAllocationSize = 0
	nFrees = 0
	nFastFrees = 0
}

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

func reserveSpaceInStringAllocator(n int) []uint16 {
	muStringAllocator.Lock()
	lastAllocationSize = n

	nLeft := len(stringAllocator) - stringAllocatorCurrPos
	if nLeft >= n {
		lastAllocationPos = stringAllocatorCurrPos
		res := stringAllocator[stringAllocatorCurrPos : stringAllocatorCurrPos+n]
		stringAllocatorCurrPos += n
		muStringAllocator.Unlock()
		return res
	}

	bufSize := stringAllocatorDefaultBufSize
	if n > bufSize {
		bufSize = n
	}
	stringAllocator = make([]uint16, bufSize, bufSize)
	lastAllocationPos = 0
	stringAllocatorCurrPos = n
	res := stringAllocator[:n]
	muStringAllocator.Unlock()
	return res
}

// ToUTF16ShortLived converts s to a zero-terminated UTF-16
// Windows string. Returns a slice that doesn't include
// terminating zero (but it's there)
func ToUTF16ShortLived(s string) []uint16 {
	// optimistically try a fast path for just ascii
	n := len(s)
	res := reserveSpaceInStringAllocator(n + 1)
	isASCII := true
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

	// TODO: fast non-ascii path. We should be able to re-use
	// res most of the time because s being utf8-encoded string
	// len(s) should be enough Unicode chars to fit it, most of the
	// time.
	// this should work as well, though as FreeUnicode() should never
	// consider this string to be allocated withing stringAllocator buffer
	FreeShortLivedUTF16(res)
	return ToUTF16(s)
}

// FreeShortLivedUTF16 frees a unicode string allocated with ToUTF16ShortLived
func FreeShortLivedUTF16(s []uint16) {
	// if s is the last allocated string, shrink stringAllocator
	// buffer so that next allocation will re-use it
	muStringAllocator.Lock()
	nFrees++

	sp := uintptr(unsafe.Pointer(&s[0]))
	lastAllocationPtr1 := &stringAllocator[lastAllocationPos]
	lastAllocationPtr := uintptr(unsafe.Pointer(lastAllocationPtr1))

	// if s is not the last allocated string, then we'll leave it in
	// the allocator buffer. Eventually the buffer will fill up,
	// we'll allocate new one and the buffer will be GC'ed
	if sp != lastAllocationPtr {
		muStringAllocator.Unlock()
		return
	}

	nFastFrees++

	// s is the last allocated string so free up the space in
	// the buffer for next allocation
	stringAllocatorCurrPos -= lastAllocationSize
	panicIf(stringAllocatorCurrPos < 0)
	lastAllocationPtr = uintptr(1) // should never match any valid address
	muStringAllocator.Unlock()
}
