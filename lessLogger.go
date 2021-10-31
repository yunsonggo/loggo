package loggo

import (
	"encoding/json"
	"sync/atomic"
	"time"
)

type LessLogger struct {
	threshold int
	lastTime  int64
	discarded uint32
}

func NewLessLogger(millis int) *LessLogger {
	return &LessLogger{
		threshold: millis,
		lastTime:  0,
	}
}

func (leg *LessLogger) Error(v ...interface{}) {
	leg.logOrDiscard(func() {
		Error(v...)
	})
}

func (leg *LessLogger) Log(keyValues ...interface{}) error {
	return json.NewEncoder(InfoLog).Encode(keyValues)
}

func (leg *LessLogger) Errorf(format string, v ...interface{}) {
	leg.logOrDiscard(func() {
		ErrorFormat(format, v...)
	})
}

func (leg *LessLogger) logOrDiscard(fn func()) {
	if leg == nil || leg.threshold <= 0 {
		fn()
		return
	}
	currentMillis := time.Now().UnixNano() / int64(time.Millisecond)
	if currentMillis-leg.lastTime > int64(leg.threshold) {
		leg.lastTime = currentMillis
		discarded := atomic.SwapUint32(&leg.discarded, 0)
		if discarded > 0 {
			ErrorFormat("Discarded %d error messages", discarded)
		}
		fn()
	} else {
		atomic.AddUint32(&leg.discarded, 1)
	}
}
