package timewheel

import (
	"time"
	"container/list"
	"strings"
	"sync"
)

type SlotJob struct {
	sum int
	do func()
}

type Slot struct {
	lock *sync.Mutex
	hooks *list.List
}

type TimeWheel struct {
	ticker *time.Ticker
	period time.Duration
	tickLife int
	slots []*Slot
	currentTick int
}

func NewTimeWheel(period time.Duration, tickLift int) *TimeWheel {
	ticker := time.NewTicker(period)
	tw := &TimeWheel{
		ticker : ticker,
		period : period,
		tickLife : tickLift,
		slots : func() []*Slot {
			s := make([]*Slot, 0, tickLift + 1)
			for i := 0 ; i < tickLift + 1 ; i ++ {
				s = append(s, &Slot{hooks : list.New(),lock : &sync.Mutex{}})
			}
			return s
		}(),
		currentTick : 0,
	}

	go func() {
		for i := 1; ; i ++ {
			i = i % tw.tickLife
			<- tw.ticker.C
			tw.currentTick = i
			tw.notify(i)
		}
	}()

	return tw
}

func (w *TimeWheel) notify(index int) {
	slots := w.slots[index]
	slots.lock.Lock()
	var n *list.Element
	for e := slots.hooks.Front() ; nil != e ; e = n{
		v := e.Value.(*SlotJob)
		v.sum--
		n = e.Next()
		if v.sum < 0 {
			slots.hooks.Remove(e)
			go v.do()
		}
	}
	slots.lock.Unlock()
}

func (w *TimeWheel) Add(timeout time.Duration, do func()) {
	sum, index := w.getReal(timeout)
	index =  ( index + w.currentTick ) % w.tickLife
	sj := &SlotJob{
		sum : sum,
		do : do,
	}
	slots := w.slots[index]
	slots.lock.Lock()
	slots.hooks.PushBack(sj)
	slots.lock.Unlock()
}

func (w *TimeWheel) getReal(timeout time.Duration) (int,int) {
	var sum,index int
	if int ( timeout % ( w.period * time.Duration(w.tickLife) ) ) == 0 {
		if timeout == 0 {
			sum = 0
			index = 0
		} else {
			sum = int(timeout / ( w.period * time.Duration(w.tickLife) )) - 1
			index = w.tickLife
		}
	} else {
		sum = int ( timeout / ( w.period * time.Duration(w.tickLife) ) )
		tmp :=  (timeout) % ( w.period * time.Duration(w.tickLife) )
		str := w.period.String()
		if strings.Contains(str, "h") {
			return sum, int(tmp / time.Hour)
		}
		if strings.Contains(str, "m"){
			return sum, int(tmp / time.Minute)
		}
		if strings.Contains(str, "ns") {
			return sum, int(tmp / time.Nanosecond)
		}
		if strings.Contains(str, "µs") {
			return sum, int(tmp / time.Microsecond)
		}
		if strings.Contains(str, "ms") {
			return sum, int(tmp / time.Millisecond)
		}
		if strings.Contains(str, "s") {
			return sum, int(tmp / time.Second)
		}
	}
	return sum,index
}
