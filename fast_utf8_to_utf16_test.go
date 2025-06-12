package w

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func isUnicodeEqual(s string, u []uint16) bool {
	s2 := FromUTF16(u)
	return s == s2
}

func TestUTF16RoundTrip(t *testing.T) {
	tests := []string{"abc", "Hello, 世界"}
	for _, test := range tests {
		u := ToUTF16ShortLived(test)
		assert.True(t, isUnicodeEqual(test, u))
		FreeShortLivedUTF16(u)
	}
}

func TestToUnicodeShortLived(t *testing.T) {
	s := "abc"
	for i := 0; i < 13; i++ {
		u := ToUTF16ShortLived(s)
		FreeShortLivedUTF16(u)
		s += s
	}
}

func TestToUnicodeShortLivedOutOfOrder(t *testing.T) {
	s := "abcdefgh"
	nNotFast := 0
	for i := 0; i < 1024; i++ {
		u := ToUTF16ShortLived(s)
		// from time to time simulate not freeing immediately
		if i%4 == 0 {
			ToUTF16ShortLived(s + "ha")
			nNotFast++
		}
		FreeShortLivedUTF16(u)
	}
}

var dontOptimezeMe []uint16

func BenchmarkToUnicode(b *testing.B) {
	s := "abcdef"
	for n := 0; n < b.N; n++ {
		for i := 0; i < 1024; i++ {
			dontOptimezeMe = ToUTF16(s)
		}
	}
	dontOptimezeMe = nil
}

func BenchmarkToUnicodeShortLived(b *testing.B) {
	s := "abcdef"
	for n := 0; n < b.N; n++ {
		for i := 0; i < 1024; i++ {
			dontOptimezeMe = ToUTF16ShortLived(s)
			FreeShortLivedUTF16(dontOptimezeMe)
		}
	}
	dontOptimezeMe = nil
}
