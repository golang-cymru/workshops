package main

import (
	"context"
	"log"
	"time"
)

func main() {
	// Create a context with a built in timeout. Timeout will be signalled
	// in 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// Just in case we exit before we timeout, ensure cancel() gets called to
	// stop the goroutine gracefully
	defer cancel()

	// Wait on a channel for the sleeper
	finished := make(chan bool, 1)

	go func(ctx context.Context) {
		i := 0
		for { // Loop forever.. and ever.. and ever...
			select {
			case <-ctx.Done(): // ... unless we get interupted!
				finished <- true
				return // Drop out of the goroutine
			default:
				i++
				time.Sleep(1 * time.Second)
				log.Printf("Slept %d second", i)
			}
		}

	}(ctx)

	<-finished
	log.Println("Timeout!")
}
