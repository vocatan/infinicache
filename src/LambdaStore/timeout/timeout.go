package timer

import (
	"github.com/wangaoone/LambdaObjectstore/lib/logger"
	"math"
	"time"
)

const TICK = int64(100 * time.Millisecond)
// For Lambdas below 0.5vCPU(896M).
const TICK_1_ERROR_EXTEND = int64(10000 * time.Millisecond)
const TICK_1_ERROR = int64(-1)
// For Lambdas with 0.5vCPU(896M) and above.
const TICK_5_ERROR_EXTEND = int64(1000 * time.Millisecond)
const TICK_5_ERROR = int64(30 * time.Millisecond)
// For Lambdas with 1vCPU(1792M) and above.
const TICK_10_ERROR_EXTEND = int64(1000 * time.Millisecond)
const TICK_10_ERROR = int64(2 * time.Millisecond)

var (
	TICK_ERROR_EXTEND = TICK_10_ERROR_EXTEND
	TICK_ERROR = TICK_10_ERROR
)

type Timeout struct {
	*time.Timer
	Start         time.Time

	lastExtension int64
	log           logger.ILogger
}

func New(d time.Duration) *Timeout {
	return &Timeout{
		Timer: time.NewTimer(d),
		lastExtension: TICK_ERROR,
		log: logger.NilLogger,
	}
}

func (t *Timeout) Restart() time.Time {
	t.Start = time.Now()
	return t.Start
}

func (t *Timeout) Since() time.Duration {
	return time.Since(t.Start)
}

func (t *Timeout) Reset() {
	// Drain the timer to be accurate and safe to reset.
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
	timeout := t.getTimeout(t.lastExtension)
	t.Timer.Reset(timeout)
	t.log.Debug("Timeout reset: %v", timeout)
}

func (t *Timeout) ResetWithExtension(ext int64) {
	t.lastExtension = ext
	t.Reset()
}

func (t *Timeout) SetLogger(log logger.ILogger) {
	t.log = log
}

func (t *Timeout) getTimeout(ext int64) time.Duration {
	if ext < 0 {
		return 0
	}

	now := time.Now().Sub(t.Start).Nanoseconds()
	return time.Duration(int64(math.Ceil(float64(now + ext) / float64(TICK)))*TICK - TICK_ERROR - now)
}