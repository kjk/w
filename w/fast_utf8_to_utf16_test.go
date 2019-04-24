package w

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToUnicodeShortLived(t *testing.T) {
	resetAllocator()
	s := "abc"
	for i := 0; i < 13; i++ {
		u := ToUnicodeShortLived(s)
		FreeShortLivedUnicode(u)
		s += s
	}
	// make sure we test allocating more than the buffer
	assert.True(t, len(s) > stringAllocatorDefaultBufSize)
	assert.Equal(t, nFrees, nFastFrees)
	assert.Equal(t, stringAllocatorCurrPos, 0)
}

func TestToUnicodeShortLivedOutOfOrder(t *testing.T) {
	resetAllocator()
	s := "abcdefgh"
	nNotFast := 0
	for i := 0; i < 1024; i++ {
		u := ToUnicodeShortLived(s)
		// from time to time simulate not freeing immediately
		if i%4 == 0 {
			ToUnicodeShortLived(s + "ha")
			nNotFast++
		}
		FreeShortLivedUnicode(u)
	}
	assert.Equal(t, nFastFrees+nNotFast, nFrees)
	assert.Equal(t, stringAllocatorCurrPos, (11+9)*nNotFast)
}

var dontOptimezeMe []uint16

func BenchmarkToUnicode(b *testing.B) {
	s := "abcdef"
	for n := 0; n < b.N; n++ {
		for i := 0; i < 1024; i++ {
			dontOptimezeMe = ToUnicode(s)
		}
	}
	dontOptimezeMe = nil
}

func BenchmarkToUnicodeShortLived(b *testing.B) {
	resetAllocator()
	s := "abcdef"
	for n := 0; n < b.N; n++ {
		for i := 0; i < 1024; i++ {
			dontOptimezeMe = ToUnicodeShortLived(s)
			FreeShortLivedUnicode(dontOptimezeMe)
		}
	}
	dontOptimezeMe = nil
}
