package main

import (
	"fmt"
	"sync"
)

// Funnel will recieve the one or many input channels, and return the single destination channel, for each source channel in argument, sepereate GOROUTINE will be craeted
// which will read the incoming data and write to the destination. For synchronization used waitGroups, when source channel is closed, gorutine will decrement the wg.Count\
// when all of the sources are closed, the destination channel will also be closed.
func Funnel(sources ...<-chan int) <-chan int {
	dest := make(chan int)
	wg := sync.WaitGroup{}

	wg.Add(len(sources))

	for _, source := range sources {
		go func(ch <-chan int) {
			//decrement before leaving gorotuine
			defer wg.Done()
			for val := range ch {
				dest <- val
			}
		}(source)
	}

	go func() {
		wg.Wait()
		close(dest)
	}()

	return dest
}

func main() {
	sources := make([]<-chan int, 0)

	for i := 0; i < 3; i++ {
		ch := make(chan int)
		sources = append(sources, ch)

		go func(ch chan int) {
			defer close(ch)
			for j := 0; j < 5; j++ {
				ch <- j
			}
		}(ch)
	}

	dest := Funnel(sources...)

	for i := range dest {
		fmt.Printf("Received: %d\n", i)
	}
}
