package main

import (
	"context"
	"log"
	"time"
)

func main() {
	// Create a context with a cancellation. The cancel functiom is returned
	// for us to use at our leisure (for example in a graceful shutdown routine)
	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context) {
		i := 0
		for { // Loop forever.. and ever.. and ever...
			select {
			case <-ctx.Done(): // ... unless we get interupted!
				log.Println("I'm being cancelled!")
				return // Drop out of the goroutine
			default:
				i++
				time.Sleep(1 * time.Second)
				log.Printf("Slept %d second", i)
			}
		}
	}(ctx)

	// Sleep for 3 seconds, then call the cancel.
	time.Sleep(3 * time.Second)
	cancel()
	log.Println("Cancelled called!!")

	// The cancel won't finish instantaneously, so let's wait a seond, just so
	// it can finish doing what it's doing.
	time.Sleep(1 * time.Second)
	log.Println("Done")
}
