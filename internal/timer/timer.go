package timer

import (
	"sync/atomic"
	"time"
)

var Time = new(atomic.Int64)

func Now() time.Time {
	return time.Unix(Time.Load(), 0)
}

// Resolution depicts how often is the time updated. By default, it's updated once
// every 500ms which is considered precise enough
const Resolution = 500 * time.Millisecond

func init() {
	Time.Store(time.Now().Unix())

	go func() {
		for {
			Time.Store(time.Now().Unix())
			time.Sleep(Resolution)
		}
	}()
}
