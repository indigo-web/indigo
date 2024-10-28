package transport

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
	_ "unsafe"
)

func BenchmarkTimeNow(b *testing.B) {
	b.Run("time.Now()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			time.Now().Add(5 * time.Second)
		}
	})

	const resolution = 50 * time.Millisecond

	t := new(atomic.Int64)
	go func() {
		for {
			t.Store(time.Now().Add(5 * time.Second).Unix())
			time.Sleep(resolution)
		}
	}()

	b.Run("atomic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = time.Unix(t.Load(), 0)
		}
	})

	var ct time.Time
	lock := new(sync.RWMutex)
	go func() {
		for {
			lock.Lock()
			ct = time.Now().Add(5 * time.Second)
			lock.Unlock()
			time.Sleep(resolution)
		}
	}()

	b.Run("rwmutex", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			lock.RLock()
			_ = ct
			lock.RUnlock()
		}
	})
}
