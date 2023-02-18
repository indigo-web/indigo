package split

import (
	"strings"
	"testing"
)

func BenchmarkSplit(b *testing.B) {
	shortSample := strings.Repeat("aaaaaaa ", 15)     // 120 characters
	longSample := strings.Repeat("aaaaaaaaaaa ", 100) // 1200 characters

	b.Run("NaiveShort", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := range shortSample {
				if shortSample[j] == ' ' {
					nop()
				}
			}
		}
	})

	b.Run("NaiveLong", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := range longSample {
				if longSample[j] == ' ' {
					nop()
				}
			}
		}
	})

	b.Run("StringIterShort", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			iter := StringIter(shortSample, ' ')

			for {
				_, err := iter()
				if err != nil {
					break
				}
			}
		}
	})

	b.Run("StringIterLong", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			iter := StringIter(longSample, ' ')

			for {
				_, err := iter()
				if err != nil {
					break
				}
			}
		}
	})
}

func nop() {}
