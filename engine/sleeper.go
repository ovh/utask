package engine

import "time"

type sleeper struct {
	sleepCount int
}

func newSleeper() *sleeper {
	return &sleeper{}
}

func (s *sleeper) sleep() {
	if s.sleepCount >= 10 {
		time.Sleep(time.Second * 10)
	} else {
		time.Sleep(time.Second * time.Duration(s.sleepCount))
		s.sleepCount++
	}
}

func (s *sleeper) wakeup() {
	s.sleepCount = 0
}
