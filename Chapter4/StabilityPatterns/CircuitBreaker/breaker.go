//This is my custom implementation of the Circuit Breaker, for educational purposes

package circuitbreaker

import (
	"fmt"
	"sync"
	"time"
)

type state string

const (
	Open     state = "open"
	HalfOpen state = "half-Open"
	Closed   state = "Closed"
)

var OpenErr error = fmt.Errorf("The Circuit is OPEN\n")

type Breaker struct {
	//represents current state
	State state

	FailCount int
	//if after the given time interval we will have that much fails 'Open' the circuit
	FailThreshold int

	//interval to change from 'Open' to the Half-Open state
	//we need timer, because once we are in the open state we will wait for some time, and change to the 'half Open' state
	//TO-DO implement some better algorithm to adjust the wait time
	WaitTime time.Duration

	//after each specified interval, if circuit is 'Closed' free up the threshold.
	//but how to do it, mb just spawn a goroutine which will do periodic checks?
	RevindInterval time.Duration

	//after consecutive success we swithc from half-open to the close
	ConsecutiveSuccess int
	HalfOpenThreshold  int

	//guard the shared resources from concurent access
	m sync.RWMutex
}

func (b *Breaker) Execute(f func() (any, error)) (any, error) {
	//run early checks
	err := b.runBefore()
	if err != nil {
		return nil, err
	}

	//run the function it self
	resp, err := f()

	//run after checks
	b.runAfter(err)

	return resp, nil

}

// it just checks if its open, if yes early returns
func (b *Breaker) runBefore() error {
	b.m.RLock()
	defer b.m.RUnlock()

	if b.State == Open {
		return OpenErr
	}
	return nil
}

// after func checks, based on error value will make decisions
func (b *Breaker) runAfter(err error) {
	b.m.RLock()
	defer b.m.RUnlock()

	if err == nil {

		if b.State == HalfOpen {
			b.m.Lock()
			b.ConsecutiveSuccess++
			if b.ConsecutiveSuccess >= b.HalfOpenThreshold {

				b.State = Closed
			}
			b.m.Lock()
		}

		return
	}

	//if its close, we will increment the threshold and if its reach the limit open the circuit
	if b.State == Closed {
		b.m.Lock()
		b.FailCount++
		b.m.Unlock()

		if b.FailCount >= b.FailThreshold {
			go b.open() //running this on gorotuine because this func will sleep for 'WaitTime' so its better to return from the runAfter func instead of waiting inside it
			return
		}
	}

	//if its half Open and still failed, back to the open
	if b.State == HalfOpen {
		b.m.Lock()
		b.ConsecutiveSuccess = 0
		b.m.Unlock()
		go b.open()
	}

}

// here we open the circuit and wait for some interval, and change to the half-open state
func (b *Breaker) open() {
	b.m.Lock()
	b.State = Open
	b.m.Unlock()

	time.Sleep(b.WaitTime)
	b.m.Lock()
	b.State = HalfOpen
	b.ConsecutiveSuccess = 0
	b.m.Unlock()
}

// it just in each revindInterval, if the circuit is still closed, will free up the failCount
func (b *Breaker) freeUp() {
	for {
		time.Sleep(b.WaitTime)
		b.m.Lock()
		if b.State == Closed {
			b.FailCount = 0
		}
		b.m.Unlock()
	}
}
