package main

import (
	"fmt"
	"strings"
	"sync"
)

var wg sync.WaitGroup

func main() {
	queue := []string{"h,e,l,l,o", "w,o,r,l,d", "t,h,i,s,", "i,s,,,", "a,,,,", "d,i,c,t,ionary", "w,i,t,h,", "w,o,r,d,s", "o,f,,,", "v,a,r,y,ing", "l,e,n,g,th"}
	cDLQ := make(chan string)

	for _, msg := range queue {
		wg.Add(1)
		go validate(cDLQ, msg)
	}

	fmt.Println("waiting....")

	go func() {
		for dl := range cDLQ {
			//send message to a dead letter queue
			fmt.Println("deadletter: " + dl)
		}
	}()

	wg.Wait()
}

func validate(ch chan string, msg string) {
	defer wg.Done()

	contents := strings.Split(msg, ",")

	for _, c := range contents {
		count := len(c)
		if count > 1 {
			ch <- msg
			break
		}
	}
}