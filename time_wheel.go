package timewheel

import (
	"time"
	"sync"
)

type TimeWheel struct {
	ticker *time.Ticker
	currentTick int
	tickPeriod time.Duration
	ticksWheel int
	lock sync.RWMutex
	quit chan bool
	slots []chan bool
}

func NewTimeWheel(tickPeriod time.Duration, ticksWheel int) *TimeWheel{
	ticker := time.NewTicker(tickPeriod)
	tw := &TimeWheel {
		ticker : ticker,
		currentTick : 0,
		tickPeriod : tickPeriod,
		ticksWheel : ticksWheel,
		lock : sync.RWMutex{},
		quit : func() chan bool{
			return make(chan bool)
		}(),
		slots : func() []chan bool{
			s := make([]chan bool, 0, ticksWheel)
			for i := 0; i < ticksWheel ; i++ {
				s = append(s, make(chan bool))
			}
			return s
		}(),
	}

	go tw.run()

	return tw
}

func (wheel *TimeWheel) run() {
	for {
		select {
			case <- wheel.quit:
				wheel.ticker.Stop()
			case <- wheel.ticker.C:
				wheel.do()

		}
	}
}

func (wheel *TimeWheel) do() {
	wheel.lock.Lock()
	last := wheel.slots[wheel.currentTick]
	wheel.slots[wheel.currentTick] = make(chan bool)
	wheel.currentTick = (wheel.currentTick + 1) % wheel.ticksWheel
	close(last)
	wheel.lock.Unlock()
}

func (wheel *TimeWheel) After(timeout time.Duration) <-chan bool{
	if timeout >= wheel.tickPeriod * time.Duration(wheel.ticksWheel) {
		panic("cant wait too long")
	}
	wheel.lock.Lock()
	idx := timeout / wheel.tickPeriod
	if idx > 0 {
		idx --
	}
	index := (wheel.currentTick + int(idx)) % wheel.ticksWheel
	c := wheel.slots[index]
	wheel.lock.Unlock()
	return c
}

func (wheel *TimeWheel) Close(timeout time.Duration) {
	wheel.quit <- true
}

