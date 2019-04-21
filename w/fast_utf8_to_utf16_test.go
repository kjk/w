package w

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToUnicodeShortLived(t *testing.T) {
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

func TestToUnicodeShortLived2(t *testing.T) {
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

