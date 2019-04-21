package w

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToUnicode(t *testing.T) {
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

