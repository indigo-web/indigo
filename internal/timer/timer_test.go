package timer

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
	_ "unsafe"

	"github.com/stretchr/testify/require"
)

func TestTime(t *testing.T) {
	const (
		threshold = 200 * time.Millisecond
		// use 1.5*Resolution in order to avoid test failures because of the Resolution+1ms error,
		// which happens rarely (approx. once every 20 runs), but better to not happen at all
		resolution = Resolution + Resolution/2
	)

	for range 2 * time.Second / threshold {
		now := Now()
		if time.Now().Sub(now) > resolution {
			require.Fail(t, "the timer is too slow")
		}

		time.Sleep(threshold)
	}
}

func BenchmarkTimeNow(b *testing.B) {
	b.Run("time.Now()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			time.Now().Add(5 * time.Second)
		}
	})

	const resolution = 50 * time.Millisecond

	t := new(atomic.Int64)
	t.Store(time.Now().UnixMilli())
	go func() {
		for {
			t.Store(time.Now().UnixMilli())
			time.Sleep(resolution)
		}
	}()

	var currtime time.Time

	b.Run("atomic (current)", func(b *testing.B) {
		for range b.N {
			millis := Time.Load()
			currtime = time.Unix(millis/1000, (millis%1000)*1e6)
		}
	})

	//nop
	func(time.Time) {}(currtime)

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
