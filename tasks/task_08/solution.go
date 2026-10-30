package main

import (
	"sync"
	"time"
)

type Clock interface{ Now() time.Time }
type Limiter struct {
	mu     sync.Mutex
	clock  Clock
	rate   float64
	burst  int
	tokens float64
	last   time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {

	var last_val time.Time
	if clock != nil {
		last_val = clock.Now()
	}

	return &Limiter{
		clock:  clock,
		rate:   ratePerSec,
		burst:  burst,
		tokens: float64(burst),
		last:   last_val,
	}
}
func (l *Limiter) AllowN(n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if n <= 0 || n > l.burst || l.burst <= 0 || l.clock == nil {
		return false
	}

	now := l.clock.Now()

	//количество секунд сколько прошло
	duration_time_execute := now.Sub(l.last).Seconds()

	if duration_time_execute > 0 && l.rate > 0 {

		l.tokens += duration_time_execute * l.rate

		if l.tokens > float64(l.burst) {
			l.tokens = float64(l.burst)
		}

		l.last = l.clock.Now()

	}

	//Удаляем n
	res := l.tokens - float64(n)

	//fmt.Printf("n = %d token = %f res = %f \n", n, l.tokens, res)

	if res < 0 {
		return false
	}

	l.tokens = res
	return true

}
