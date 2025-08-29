package timer

import (
	"sync/atomic"
	"time"
)

// Time contains the unix-time in milliseconds updated every [Resolution] milliseconds
var Time = new(atomic.Int64)

func Now() time.Time {
	millis := Time.Load()
	return time.Unix(millis/1000, (millis%1000)*1e6)
}

// Resolution is the frequency at which time is updated. Default 500ms are
// precise enough for setting I/O deadlines
const Resolution = 500 * time.Millisecond

func init() {
	// In real world this won't matter, however removing the line causes the timer test fail
	// in about half of all runs.
	Time.Store(time.Now().UnixMilli())

	go func() {
		for {
			time.Sleep(Resolution)
			Time.Store(time.Now().UnixMilli())
		}
	}()
}
