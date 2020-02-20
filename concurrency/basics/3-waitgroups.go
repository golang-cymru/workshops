package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

func main() {
	queue := []string{"h,e,l,l,o", "w,o,r,l,d", "t,h,i,s,", "i,s,,,", "a,,,,", "d,i,c,t,ionary", "w,i,t,h,", "w,o,r,d,s", "o,f,,,", "v,a,r,y,ing", "l,e,n,g,th"}

	for _, msg := range queue {
		wg.Add(1)
		go validate(msg)
	}

	fmt.Println("waiting....")
	wg.Wait()
}

func validate(msg string) {
	defer wg.Done()

	time.Sleep(2 * time.Second)
	contents := strings.Split(msg, ",")

	for _, c := range contents {
		count := len(c)
		fmt.Println(fmt.Sprintf("%d", count))
	}
}
