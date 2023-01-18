package tcp

import (
	"testing"
	"time"
)

// Just to demonstrate how much it takes to prepare the socket for reading (the operation of setting
// the deadline has to be done on every read from socket)
// Note: on my MacBook Air M1 2020 it takes about 70ns
func BenchmarkTimeNow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		time.Now().Add(5 * time.Second)
	}
}
