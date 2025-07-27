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
	go func() {
		for {
			Time.Store(time.Now().UnixMilli())
			time.Sleep(Resolution)
		}
	}()
}
