package main

import (
	"fmt"
	"strings"
	"time"
)

func main() {
	queue := []string{"h,e,l,l,o", "w,o,r,l,d", "t,h,i,s,", "i,s,,,", "a,,,,", "d,i,c,t,ionary", "w,i,t,h,", "w,o,r,d,s", "o,f,,,", "v,a,r,y,ing", "l,e,n,g,th"}

	for _, msg := range queue {
		validate(msg)
	}

}

func validate(msg string) {
	time.Sleep(2 * time.Second)
	contents := strings.Split(msg, ",")

	for _, c := range contents {
		count := len(c)
		fmt.Println(fmt.Sprintf("%d", count))
	}
}
